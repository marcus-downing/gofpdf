package types

import (
	"bytes"
	// "fmt"
	"regexp"
	"math"
	"strconv"
)

// ValueType is an enum of the given Types
type ValueType int8

const (
	TypeNull       = 0  // Null
	TypeNumeric    = 1  // Numeric
	TypeToken      = 2  // Token
	TypeHex        = 3  // Hex value
	TypeString     = 4  // String value
	TypeDictionary = 5  // Dictionary
	TypeArray      = 6  // Array
	TypeObjDec     = 7  // Object declaration
	TypeObjRef     = 8  // Object reference
	TypeObject     = 9  // Object
	TypeStream     = 10 // Stream
	TypeBoolean    = 11 // Boolean
	TypeReal       = 12 // Real number
)

// Value is one the given Types
type Value interface {
	Type() ValueType
	Equals(Value) bool
	ToNumeric() Numeric
	ToReal() Real
	ToString() String
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
	return TypeToken
}

// Equals tests if the token matches a given object
func (t Token) Equals(v Value) bool {
	t2, ok := v.(Token)
	return ok && bytes.Equal(t, t2)
}

// ToNumeric converts a value to a Numeric value
func (t Token) ToNumeric() Numeric {
	str := string(t)
	num, err := strconv.ParseInt(str, 0, 64)
	if err != nil {
		// parser.SetWarning(err)
		// return Numeric_NaN
	}
	return Numeric(num)
}

// ToReal converts a value to a Real value
func (t Token) ToReal() Real {
	str := string(t)
	flo, err := strconv.ParseFloat(str, 64)
	if err != nil {
		// parser.SetWarning(err)
		return Real_NaN
	}
	return Real(flo)
}

// ToString converts a value to a String value
func (t Token) ToString() String {
	str := string(t)
	return String(str)
}

type PDFDictionary interface {
	Get(key string) (Value, bool)
}

// Dictionary is a mapping from names to values
type Dictionary map[string]Value

// Type of a Dictionary value
func (d Dictionary) Type() ValueType {
	return TypeDictionary
}

// Equals checks if the value is an equivalent dictionary
func (d Dictionary) Equals(v Value) bool {
	if _, ok := v.(Dictionary); !ok {
		return false
	}
	return false
}

// ToNumeric converts a value to a Numeric value
func (d Dictionary) ToNumeric() Numeric {
	// parser.SetWarning(errors.New("Cannot convert dictionary to numeric"))
	return Numeric_NaN
}

// ToReal converts a value to a Real value
func (d Dictionary) ToReal() Real {
	// parser.SetWarning(errors.New("Cannot convert dictionary to real"))
	return Real_NaN
}

// ToString converts a value to a String value
func (d Dictionary) ToString() String {
	return ""
}

// Get a value from the dictionary
func (d Dictionary) Get(key string) (Value, bool) {
	v, b := d[key]
	return v, b
}

// String is a string value
type String string

// Type of a String value
func (s String) Type() ValueType {
	return TypeString
}

// Equals checks if the value is an equivalent string
func (s String) Equals(v Value) bool {
	s2, ok := v.(String)
	return ok && s == s2
}

// ToNumeric converts a value to a Numeric value
func (s String) ToNumeric() Numeric {
	num, err := strconv.ParseInt(string(s), 10, 64)
	if err != nil {
		// return 0, err
		// return Numeric_NaN
	}
	return Numeric(num)
}

// ToReal converts a value to a Real value
func (s String) ToReal() Real {
	flo, err := strconv.ParseFloat(string(s), 64)
	if err != nil {
		// return 0, err
		return Real_NaN
	}
	return Real(flo)
}

// ToString converts a value to a String value
func (s String) ToString() String {
	return s
}

// String converts a String value to a Go string
func (s String) String() string {
	return string(s)
}

// Hex is a hex value
type Hex string

// Type of a Hex value
func (h Hex) Type() ValueType {
	return TypeHex
}

// Equals checks if the value is an equivalent hex value
func (h Hex) Equals(v Value) bool {
	h2, ok := v.(Hex)
	return ok && h == h2
}

// ToNumeric converts a value to a Numeric value
func (h Hex) ToNumeric() Numeric {
	// str := "Ox"+h
	num, err := strconv.ParseInt(string(h), 16, 64)
	if err != nil {
		// return 0, err
	}
	return Numeric(num)
}

// ToReal converts a value to a Real value
func (h Hex) ToReal() Real {
	// str := "Ox"+h
	num, err := strconv.ParseInt(string(h), 16, 64)
	if err != nil {
		// return 0, err
		return Real_NaN
	}
	return Real(num)
}

// ToString converts a value to a String value
func (h Hex) ToString() String {
	str := string(h)
	return String(str)
}

// Numeric is an integer value
type Numeric int64
var Numeric_NaN Numeric = 0

// Type of a Numeric value
func (n Numeric) Type() ValueType {
	return TypeNumeric
}

// Equals checks if the value is an equivalent number
func (n Numeric) Equals(v Value) bool {
	n2, ok := v.(Numeric)
	return ok && n == n2
}

// ToNumeric converts a value to a Numeric value
func (n Numeric) ToNumeric() Numeric {
	return n
}

// ToReal converts a value to a Real value
func (n Numeric) ToReal() Real {
	return Real(n)
}

// ToInt64 returns the raw integer value
func (n Numeric) ToInt64() int64 {
	return int64(n)
}

// ToString converts a value to a String value
func (n Numeric) ToString() String {
	str := strconv.FormatInt(int64(n), 10)
	return String(str)
}


// Real is a floating-point value
type Real float64
var Real_NaN Real = Real(math.NaN())

// Type of a Real value
func (r Real) Type() ValueType {
	return TypeReal
}

// Equals checks if the value is an equivalent number
func (r Real) Equals(v Value) bool {
	r2, ok := v.(Real)
	return ok && r == r2
}

// ToNumeric converts a value to a Numeric value
func (r Real) ToNumeric() Numeric {
	return Numeric(r)
}

// ToReal converts a value to a Real value
func (r Real) ToReal() Real {
	return r
}

func (r Real) ToFloat64() float64 {
	return float64(r)
}

// ToString converts a value to a String value
func (r Real) ToString() String {
	str := strconv.FormatFloat(float64(r), 'f', -1, 64)
	return String(str)
}

// Boolean is a boolean value
type Boolean bool

// Type of a Boolean value
func (b Boolean) Type() ValueType {
	return TypeBoolean
}

// Equals checks if the value is an equivalent boolean
func (b Boolean) Equals(v Value) bool {
	b2, ok := v.(Boolean)
	return ok && b == b2
}

// ToNumeric converts a value to a Numeric value
func (b Boolean) ToNumeric() Numeric {
	if b {
		return Numeric(1)
	} else {
		return Numeric(0)
	}
}

// ToReal converts a value to a Real value
func (b Boolean) ToReal() Real {
	if b {
		return Real(1.0)
	} else {
		return Real(0.0)
	}
}

// ToString converts a value to a String value
func (b  Boolean) ToString() String {
	if b {
		return "true"
	} else {
		return "false"
	}
}

// Array is a sequence of value
type Array []Value

// Type of an Array value
func (a Array) Type() ValueType {
	return TypeArray
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

// ToNumeric converts a value to a Numeric value
func (a Array) ToNumeric() Numeric {
	// errors.New("Cannot convert array to numeric")
	return Numeric_NaN
}

// ToReal converts a value to a Real value
func (a Array) ToReal() Real {
	// return 0, errors.New("Cannot convert array to real")
	return Real_NaN
}

// ToString converts a value to a String value
func (a Array) ToString() String {
	return ""
}

// Null is an undefined value
type Null struct{}

// Type of a Null value
func (n Null) Type() ValueType {
	return TypeNull
}

// Equals checks if the value is also null
func (n Null) Equals(v Value) bool {
	_, ok := v.(Null)
	return ok
}

// ToNumeric converts a value to a Numeric value
func (n Null) ToNumeric() Numeric {
	// return 0, errors.New("Cannot convert null to numeric")
	return Numeric_NaN
}

// ToReal converts a value to a Real value
func (n Null) ToReal() Real {
	// return 0, errors.New("Cannot convert null to real")
	return Real_NaN
}

// ToString converts a value to a String value
func (n Null) ToString() String {
	return ""
}

// ObjectRef is an indirect object reference
type ObjectRef struct {
	Obj int
	Gen int
}

// Type of a ObjectRef value
func (r ObjectRef) Type() ValueType {
	return TypeObjRef
}

// Equals checks if the value is also null
func (r ObjectRef) Equals(v Value) bool {
	r2, ok := v.(ObjectRef)
	return ok && r == r2
}

// ToNumeric converts a value to a Numeric value
func (r ObjectRef) ToNumeric() Numeric {
	// return 0, errors.New("Cannot convert object ref to numeric")
	return Numeric_NaN
}

// ToReal converts a value to a Real value
func (r ObjectRef) ToReal() Real {
	// return 0, errors.New("Cannot convert object ref to real")
	return Real_NaN
}

// ToString converts a value to a String value
func (r ObjectRef) ToString() String {
	return ""
}

// ObjectDeclaration is a object identifier
type ObjectDeclaration struct {
	ObjectRef
	Values []Value
}

// Type of a ObjectRef value
func (r ObjectDeclaration) Type() ValueType {
	if len(r.Values) > 0 && r.Values[0].Type() == TypeStream {
		return TypeStream
	}
	return TypeObjRef
}

// GetParam looks up values in an object's dictionary (if it has one)
func (r ObjectDeclaration) Get(key string) (Value, bool) {
	for _, v := range r.Values {
		if v.Type() == TypeDictionary {
			d := v.(Dictionary)
			if value, ok := d[key]; ok {
				return value, true
			}
		}
	}
	return nil, false
}

// GetDictionary picks an object's dictionary out of its value set (if it has one)
func (r ObjectDeclaration) GetDictionary() Dictionary {
	for _, v := range r.Values {
		if v.Type() == TypeDictionary {
			return v.(Dictionary)
		}
	}
	return Dictionary(map[string]Value{})
}

// Equals checks if the value is also null
func (r ObjectDeclaration) Equals(v Value) bool {
	r2, ok := v.(ObjectDeclaration)
	return ok && r.ObjectRef == r2.ObjectRef
}

// ToNumeric converts a value to a Numeric value
func (r ObjectDeclaration) ToNumeric() Numeric {
	// return 0, errors.New("Cannot convert object declaration to numeric")
	return Numeric_NaN
}

// ToReal converts a value to a Real value
func (r ObjectDeclaration) ToReal() Real {
	// return 0, errors.New("Cannot convert object declaration to real")
	return Real_NaN
}

// ToString converts a value to a String value
func (r ObjectDeclaration) ToString() String {
	return ""
}

// Stream is a blob value
type Stream struct {
	Parameters Dictionary
	Bytes      []byte
}

// Type of a Stream value
func (s Stream) Type() ValueType {
	return TypeStream
}

// Equals checks if the value is an equivalent stream
func (s Stream) Equals(v Value) bool {
	s2, ok := v.(Stream)
	return ok && bytes.Equal(s.Bytes, s2.Bytes)
}

// ToNumeric converts a value to a Numeric value
func (s Stream) ToNumeric() Numeric {
	// return 0, errors.New("Cannot convert stream to numeric")
	return Numeric_NaN
}

// ToReal converts a value to a Real value
func (s Stream) ToReal() Real {
	// return 0, errors.New("Cannot convert stream to real")
	return Real_NaN
}

// ToString converts a value to a String value
func (s Stream) ToString() String {
	return ""
}
