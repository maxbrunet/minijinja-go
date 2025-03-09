package minijinja

import "strconv"

// valueKind represents the kind of a value.
type valueKind uint

const (
	valueKindUndefined valueKind = iota
	valueKindNone
	valueKindBool
	valueKindNumber
	valueKindString
	valueKindBytes
	valueKindSeq
	valueKindMap
	valueKindIterable
	valueKindPlain
	valueKindInvalid
)

var valueKindNames = []string{
	valueKindUndefined: "undefined",
	valueKindNone:      "none",
	valueKindBool:      "bool",
	valueKindNumber:    "number",
	valueKindString:    "string",
	valueKindBytes:     "bytes",
	valueKindSeq:       "seq",
	valueKindMap:       "map",
	valueKindIterable:  "iterable",
	valueKindPlain:     "plain",
	valueKindInvalid:   "invalid",
}

func (vk valueKind) String() string {
	if uint(vk) < uint(len(valueKindNames)) {
		return valueKindNames[vk]
	}
	return "valueKind" + strconv.FormatUint(uint64(vk), 10)
}
