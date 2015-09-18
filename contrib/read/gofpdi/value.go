package gofpdi

import (
	"bytes"
	"regexp"
)

// ValueType is an enum of the given types
type ValueType int8

const (
	typeNull       = 0  // Null
	typeNumeric    = 1  // Numeric
	typeToken      = 2  // Token
	typeHex        = 3  // Hex value
	typeString     = 4  // String value
	typeDictionary = 5  // Dictionary
	typeArray      = 6  // Array
	typeObjDec     = 7  // Object declaration
	typeObjRef     = 8  // Object reference
	typeObject     = 9  // Object
	typeStream     = 10 // Stream
	typeBoolean    = 11 // Boolean
	typeReal       = 12 // Real number
)

// Value is one the given types
type Value interface {
	Type() ValueType
	Equals(Value) bool
}

// Token is a unit of PDF syntax.
// Not all tokens can be safely represented as strings.
type Token []byte

// String converts a token to a string
func (t Token) String() string {
	return string(t)
}

// ToRegex creates a regular expression to match the token
func (t Token) ToRegex() string {
	return regexp.QuoteMeta(string(t))
}

// Type of a token value
func (t Token) Type() ValueType {
	return typeToken
}

// Equals tests if the token matches a given object
func (t Token) Equals(v Value) bool {
	t2, ok := v.(Token)
	return ok && bytes.Equal(t, t2)
}

// Dictionary is a mapping from names to values
type Dictionary map[string]Value

// Type of a Dictionary value
func (d Dictionary) Type() ValueType {
	return typeDictionary
}

// Equals checks if the value is an equivalent dictionary
func (d Dictionary) Equals(v Value) bool {
	if _, ok := v.(Dictionary); !ok {
		return false
	}
	return false
}

// Get a value from the dictionary
// func (d Dictionary) Get(key string) Value {

// }

// String is a string value
type String string

// Type of a String value
func (s String) Type() ValueType {
	return typeString
}

// Equals checks if the value is an equivalent string
func (s String) Equals(v Value) bool {
	s2, ok := v.(String)
	return ok && s == s2
}

// Stream is a blob value
type Stream []byte

// Type of a Stream value
func (s Stream) Type() ValueType {
	return typeStream
}

// Equals checks if the value is an equivalent stream
func (s Stream) Equals(v Value) bool {
	s2, ok := v.(Stream)
	return ok && bytes.Equal(s, s2)
}

// Hex is a hex value
type Hex string

// Type of a Hex value
func (h Hex) Type() ValueType {
	return typeHex
}

// Equals checks if the value is an equivalent hex value
func (h Hex) Equals(v Value) bool {
	h2, ok := v.(Hex)
	return ok && h == h2
}

// Numeric is an integer value
type Numeric int64

// Type of a Numeric value
func (n Numeric) Type() ValueType {
	return typeNumeric
}

// Equals checks if the value is an equivalent number
func (n Numeric) Equals(v Value) bool {
	n2, ok := v.(Numeric)
	return ok && n == n2
}

// Real is a floating-point value
type Real float64

// Type of a Real value
func (r Real) Type() ValueType {
	return typeReal
}

// Equals checks if the value is an equivalent number
func (r Real) Equals(v Value) bool {
	r2, ok := v.(Real)
	return ok && r == r2
}

// Boolean is a boolean value
type Boolean bool

// Type of a Boolean value
func (b Boolean) Type() ValueType {
	return typeBoolean
}

// Equals checks if the value is an equivalent boolean
func (b Boolean) Equals(v Value) bool {
	b2, ok := v.(Boolean)
	return ok && b == b2
}

// Array is a sequence of value
type Array []Value

// Type of an Array value
func (a Array) Type() ValueType {
	return typeArray
}

// Equals checks if the value is an equivalent array
func (a Array) Equals(v Value) bool {
	a2, ok := v.(Array)
	if !ok {
		return false
	}
	if len(a) != len(a2) {
		return false
	}
	for i, v := range a {
		v2 := a2[i]
		if !v.Equals(v2) {
			return false
		}
	}
	return true
}

// Null is an undefined value
type Null struct{}

// Type of a Null value
func (n Null) Type() ValueType {
	return typeNull
}

// Equals checks if the value is also null
func (n Null) Equals(v Value) bool {
	_, ok := v.(Null)
	return ok
}

// ObjectRef is an indirect object reference
type ObjectRef struct {
	Obj int
	Gen int
}

// Type of a ObjectRef value
func (r ObjectRef) Type() ValueType {
	return typeObjRef
}

// Equals checks if the value is also null
func (r ObjectRef) Equals(v Value) bool {
	r2, ok := v.(ObjectRef)
	return ok && r == r2 // r.obj == r2.obj && r.gen == r2.gen
}

// ObjectDeclaration is a object identifier
type ObjectDeclaration struct {
	Obj int
	Gen int
	Values []Value
}

// Type of a ObjectRef value
func (r ObjectDeclaration) Type() ValueType {
	if len(r.Values) > 0 && r.Values[0].Type() == typeStream {
		return typeStream
	}
	return typeObjRef
}

// Equals checks if the value is also null
func (r ObjectDeclaration) Equals(v Value) bool {
	r2, ok := v.(ObjectDeclaration)
	return ok && r == r2 // r.obj == r2.obj && r.gen == r2.gen
}