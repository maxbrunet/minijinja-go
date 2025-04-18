#requires -Version 5.1
<#
.SYNOPSIS
    Builds the minijinja-cabi Rust library and copies the C header.

.DESCRIPTION
    This script performs the following tasks:
      - Builds the Rust minijinja-cabi library.
      - Moves the built .dll to the library directory.
      - Copies the minijinja-cabi header file to the include directory.
#>

#region Script Configuration

if (-not $env:MZ_CABI_INCLUDE_DIR) { $env:MZ_CABI_INCLUDE_DIR = 'include' }
if (-not $env:MZ_CABI_LIB_DIR)     { $env:MZ_CABI_LIB_DIR     = 'lib' }

#endregion

#region Functions

function Check-Requirements {
    if (-not (Get-Command cargo -ErrorAction SilentlyContinue)) {
        Write-Error @"
>>> A Rust toolchain is required to build minijinja.
>>> Please see https://www.rust-lang.org/tools/install for instructions.
"@
        exit 1
    }
}

function Ensure-SrcExists {
    if (-not (Test-Path 'src')) {
        New-Item -ItemType Directory -Path $srcDir -Force | Out-Null
    }
    if (-not (Test-Path 'src\lib.rs')) {
        New-Item -ItemType File -Path $libRs -Force | Out-Null
    }
}

function Get-MJCabiSourceDir {
    $metadata = cargo metadata --format-version 1 --locked | ConvertFrom-Json

    $package = $metadata.packages | Where-Object { $_.name -eq "minijinja-cabi" }

    $manifestPath = $package.manifest_path
    $sourceDir = Split-Path -Path $manifestPath -Parent
    return $sourceDir
}

function Build-MJCabiLib {
    cargo build --locked --package=minijinja-cabi --release

    $libDir = $env:MZ_CABI_LIB_DIR
    if (Test-Path $libDir) {
        Remove-Item -Path $libDir -Recurse -Force
    }
    New-Item -ItemType Directory -Path $libDir -Force | Out-Null

    $destinationLibPath = Join-Path -Path $libDir -ChildPath "minijinja_cabi.dll"
    Move-Item -Path "target\release\minijinja_cabi.dll" -Destination $destinationLibPath -Force

    # HACK: For some reason, the built DLL file is minijinja_cabi.dll, but
    # the linker uses a name inside its headers, usually something like
    # minijinja_cabi-4c3a24a9ee0120b4.dll. Not sure of how to tell the Rust
    # compiler to keep the name consistent or the linker to use the file name.
    # So the regular DLL is required during build-time and the suffixed one
    # during run-time.
    Write-Host ">>> Creating DLL symlink with name used at runtime..."
    $dllName = objdump --all-headers --section=.rdata $destinationLibPath
        | ForEach-Object {
            if ($_ -match '^Name.+(?<name>minijinja_cabi(?:-\w+)?\.dll)$') {
                $matches.name
            }
        }
    $symLinkPath = Join-Path -Path $libDir -ChildPath $dllName
    New-Item -ItemType SymbolicLink -Path $symLinkPath -Target "minijinja_cabi.dll"
}

function Copy-MJCabiHeaders {
    param(
        [Parameter(Mandatory)]
        [string]$SourceDir
    )

    $includeDir = $env:MZ_CABI_INCLUDE_DIR
    if (Test-Path $includeDir) {
        Remove-Item -Path $includeDir -Recurse -Force
    }
    New-Item -ItemType Directory -Path $includeDir -Force | Out-Null

    $headerSource = Join-Path -Path $SourceDir -ChildPath "include\minijinja.h"
    Copy-Item -Path $headerSource -Destination $includeDir -Force
}

#endregion

#region Main

function Main {
    Check-Requirements

    Set-Location -Path $PSScriptRoot

    Ensure-SrcExists

    Write-Host ">>> Getting minijinja-cabi source directory..."
    $mjCabiSourceDir = Get-MJCabiSourceDir

    Write-Host ">>> Building minijinja-cabi..."
    Build-MJCabiLib

    Write-Host ">>> Copying minijinja-cabi C header..."
    Copy-MJCabiHeaders -SourceDir $mjCabiSourceDir

    Write-Host '>>> Done!'
}

Main

#endregion
