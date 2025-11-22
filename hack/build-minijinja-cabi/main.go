// Package main builds and installs the minijinja-cabi library and headers.
package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"runtime"
	"syscall"
)

const (
	defaultIncludeDir = "include"
	defaultLibDir     = "lib"
	defaultLogLevel   = slog.LevelInfo

	dirPerm  = 0o755
	filePerm = 0o644
)

var (
	errMustProvideModDir      = errors.New("flag --mod-dir must be provided")
	errMissingCommand         = errors.New("required command not found")
	errMinijinjaCABINotFound  = errors.New("minijinja-cabi package not found")
	errUnsupportedPlatform    = errors.New("unsupported platform")
	errRuntimeDLLNameNotFound = errors.New("runtime DLL filename not found")

	winRuntimeDLLRegexp = regexp.MustCompile(
		`^Name.+(?P<name>minijinja_cabi(?:-\w+)?\.dll)$`,
	)
)

type builder struct {
	logger     *slog.Logger
	includeDir string
	libDir     string
	modDir     string
}

func newBuilder(logger *slog.Logger) *builder {
	return &builder{
		logger: logger,
	}
}

func (b *builder) parseFlags(fs *flag.FlagSet, args []string) error {
	fs.StringVar(
		&b.modDir,
		"mod-dir",
		"",
		"The directory of the minijinja-go module [required]",
	)
	fs.StringVar(
		&b.includeDir,
		"include-dir",
		defaultIncludeDir,
		"The include directory",
	)
	fs.StringVar(&b.libDir, "lib-dir", defaultLibDir, "The lib directory")
	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("failed to parse flags: %w", err)
	}

	if b.modDir == "" {
		return errMustProvideModDir
	}

	return nil
}

// checkRequirements checks if the required external commands are available.
func (b *builder) checkRequirements() error {
	if _, err := exec.LookPath("cargo"); err != nil {
		b.logger.Error(
			"cargo not found",
			slog.String("hint", "install Rust toolchain"),
			slog.String("url", "https://www.rust-lang.org/tools/install"),
		)
		return fmt.Errorf("cargo: %w", errMissingCommand)
	}
	return nil
}

// ensureSrcExists ensures the Rust src directory is present. It creates it if
// not.
func (b *builder) ensureSrcExists() error {
	srcDir := filepath.Join(b.modDir, "src")
	if err := os.MkdirAll(srcDir, dirPerm); err != nil {
		return fmt.Errorf("mkdir src: %w", err)
	}
	libPath := filepath.Join(srcDir, "lib.rs")
	f, err := os.OpenFile( //nolint: gosec
		libPath,
		os.O_RDONLY|os.O_CREATE,
		filePerm,
	)
	if err != nil {
		return fmt.Errorf("touch lib.rs: %w", err)
	}
	_ = f.Close()
	return nil
}

// getMjCabiSourceDir uses the `cargo metadata` command to locate the source
// directory of the minijinja-cabi Rust crate.
func (b *builder) getMjCabiSourceDir(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(
		ctx,
		"cargo",
		"metadata",
		"--format-version=1",
		"--locked",
	)
	cmd.Dir = b.modDir
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("cargo metadata: %w", err)
	}

	var meta struct {
		Packages []struct {
			Name         string `json:"name"`
			ManifestPath string `json:"manifest_path"` //nolint: tagliatelle
		} `json:"packages"`
	}
	if err := json.Unmarshal(out, &meta); err != nil {
		return "", fmt.Errorf("parse cargo metadata JSON: %w", err)
	}
	for _, p := range meta.Packages {
		if p.Name == "minijinja-cabi" {
			return filepath.Dir(p.ManifestPath), nil
		}
	}
	return "", errMinijinjaCABINotFound
}

// getLibName returns the library filename for the current OS.
func getLibName() (string, error) {
	switch runtime.GOOS {
	case "darwin":
		return "libminijinja_cabi.dylib", nil
	case "freebsd", "linux", "netbsd":
		return "libminijinja_cabi.so", nil
	case "windows":
		return "minijinja_cabi.dll", nil
	default:
		return "", fmt.Errorf("%w: %s", errUnsupportedPlatform, runtime.GOOS)
	}
}

// createWindowsDLLSymlink creates a symlink with the runtime library name
// pointing to built library file (Windows only). It uses the `objdump` command
// to read the runtime name from the built library.
func (b *builder) createWindowsDLLSymlink(
	ctx context.Context,
	builtDLLPath, libDir string,
) error {
	cmd := exec.CommandContext(
		ctx,
		"objdump",
		"--all-headers",
		"--section=.rdata",
		builtDLLPath,
	)
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("objdump: %w", err)
	}

	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		line := scanner.Text()
		if matches := winRuntimeDLLRegexp.FindStringSubmatch(line); len(
			matches,
		) > 1 {
			name := matches[1]
			srcPath := filepath.Base(builtDLLPath)
			linkPath := filepath.Join(libDir, name)

			_ = os.Remove(linkPath)
			if err := os.Symlink(srcPath, linkPath); err != nil {
				return fmt.Errorf(
					"failed to create symlink for runtime DLL name: %w",
					err,
				)
			}

			b.logger.Info(
				"runtime DLL alias created",
				slog.String("alias", name),
				slog.String("source", srcPath),
			)

			return nil
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scan objdump output: %w", err)
	}

	return errRuntimeDLLNameNotFound
}

// buildMjCabiLib builds the minijinja-cabi crate using the `cargo build`
// command.
func (b *builder) buildMjCabiLib(ctx context.Context) error {
	cmd := exec.CommandContext(
		ctx,
		"cargo",
		"build",
		"--locked",
		"--package=minijinja-cabi",
		"--release",
	)
	cmd.Dir = b.modDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("cargo build: %w", err)
	}

	libDir := filepath.Join(b.modDir, b.libDir)
	if err := os.RemoveAll(libDir); err != nil {
		return fmt.Errorf("remove lib dir: %w", err)
	}
	if err := os.MkdirAll(libDir, dirPerm); err != nil {
		return fmt.Errorf("mkdir lib dir: %w", err)
	}

	libName, err := getLibName()
	if err != nil {
		return err
	}
	srcPath := filepath.Join(
		b.modDir,
		"target",
		"release",
		libName,
	)
	dstPath := filepath.Join(libDir, libName)

	if err := os.Rename(srcPath, dstPath); err != nil {
		return fmt.Errorf("failed to move library to lib directory: %w", err)
	}
	b.logger.Info("library installed", slog.String("path", dstPath))

	// HACK: For some reason on Windows, the built DLL file is
	// minijinja_cabi.dll, but the linker uses a name inside its headers, usually
	// something like minijinja_cabi-4c3a24a9ee0120b4.dll. Not sure of how to
	// tell the Rust compiler to keep the name consistent or the linker to use
	// the file name. So the regular DLL is required during build-time and the
	// suffixed one during run-time.
	if runtime.GOOS == "windows" {
		if err := b.createWindowsDLLSymlink(ctx, dstPath, libDir); err != nil {
			return err
		}
	}

	return nil
}

// copyMjCabiHeaders copies the C headers from the minijinja-cabi Rust crate
// source directory.
func (b *builder) copyMjCabiHeaders(sourceDir string) error {
	includeDir := filepath.Join(b.modDir, b.includeDir)
	if err := os.RemoveAll(includeDir); err != nil {
		return fmt.Errorf("remove include dir: %w", err)
	}
	if err := os.MkdirAll(includeDir, dirPerm); err != nil {
		return fmt.Errorf("mkdir include dir: %w", err)
	}
	src := filepath.Join(sourceDir, "include", "minijinja.h")
	dst := filepath.Join(includeDir, "minijinja.h")

	in, err := os.Open(src) //nolint: gosec
	if err != nil {
		return fmt.Errorf("open source header: %w", err)
	}
	defer in.Close()
	out, err := os.OpenFile( //nolint: gosec
		dst,
		os.O_CREATE|os.O_WRONLY|os.O_TRUNC,
		filePerm,
	)
	if err != nil {
		return fmt.Errorf("create dst header: %w", err)
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return fmt.Errorf("copy header: %w", err)
	}
	b.logger.Info("header installed", slog.String("path", dst))
	return nil
}

func (b *builder) run(ctx context.Context) error {
	if err := b.checkRequirements(); err != nil {
		return err
	}

	if err := b.ensureSrcExists(); err != nil {
		return err
	}

	b.logger.Info("getting minijinja-cabi source directory")
	mjSource, err := b.getMjCabiSourceDir(ctx)
	if err != nil {
		return err
	}
	b.logger.Info("source dir found", slog.String("path", mjSource))

	b.logger.Info("building minijinja-cabi")
	if err := b.buildMjCabiLib(ctx); err != nil {
		return err
	}

	b.logger.Info("copying minijinja-cabi C headers")
	if err := b.copyMjCabiHeaders(mjSource); err != nil {
		return err
	}

	return nil
}

func main() {
	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: defaultLogLevel,
	})
	logger := slog.New(handler)
	builder := newBuilder(logger)
	if err := builder.parseFlags(flag.CommandLine, os.Args[1:]); err != nil {
		builder.logger.Error("invalid flags", slog.String("err", err.Error()))
		flag.Usage()
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)

	if err := builder.run(ctx); err != nil {
		builder.logger.Error("builder failed", slog.String("err", err.Error()))
		cancel()
		os.Exit(1)
	}
	cancel()
	builder.logger.Info("done")
	os.Exit(0)
}
