package object

import (
	"bytes"
	"fmt"
	"hash/fnv"
	"monkey/ast"
	"strings"
)

// BuiltinFunction 내장 함수를 표현하는 타입
// 한 개 이상의 Object 를 받아 하나의 반환값으로서의 Object 를 반환하는 함수
type BuiltinFunction func(args ...Object) Object

type ObjectType string

const (
	NULL_OBJ  = "NULL"
	ERROR_OBJ = "ERROR"

	INTEGER_OBJ = "INTEGER"
	BOOLEAN_OBJ = "BOOLEAN"
	STRING_OBJ  = "STRING" // StringLiteral 을 평가하기 위한 객체

	RETURN_VALUE_OBJ = "RETURN_VALUE"

	FUNCTION_OBJ = "FUNCTION"
	BUILTIN_OBJ  = "BUILTIN"

	ARRAY_OBJ = "ARRAY" // ArrayLiteral 을 평가하기 위한 객체
	HASH_OBJ  = "HASH"
)

type HashKey struct {
	Type  ObjectType
	Value uint64
}

// Hashable 해시 리터럴이나 해시 인덱스 표현식을 평가할 때, 주어진 객체가 해시키로 적절한지 평가하기 위함
type Hashable interface {
	HashKey() HashKey
}

type Object interface {
	Type() ObjectType
	Inspect() string
}

type Integer struct {
	Value int64
}

func (i *Integer) Type() ObjectType { return INTEGER_OBJ }
func (i *Integer) Inspect() string  { return fmt.Sprintf("%d", i.Value) }
func (i *Integer) HashKey() HashKey {
	return HashKey{Type: i.Type(), Value: uint64(i.Value)}
}

type Boolean struct {
	Value bool
}

func (b *Boolean) Type() ObjectType { return BOOLEAN_OBJ }
func (b *Boolean) Inspect() string  { return fmt.Sprintf("%t", b.Value) }
func (b *Boolean) HashKey() HashKey {
	var value uint64

	if b.Value {
		value = 1
	} else {
		value = 0
	}

	return HashKey{Type: b.Type(), Value: value}
}

type Null struct{}

func (n *Null) Type() ObjectType { return NULL_OBJ }
func (n *Null) Inspect() string  { return "null" }

type ReturnValue struct {
	Value Object
}

func (rv *ReturnValue) Type() ObjectType { return RETURN_VALUE_OBJ }
func (rv *ReturnValue) Inspect() string  { return rv.Value.Inspect() }

type Error struct {
	Message string
}

func (e *Error) Type() ObjectType { return ERROR_OBJ }
func (e *Error) Inspect() string  { return "ERROR: " + e.Message }

type Function struct {
	Parameters []*ast.Identifier
	Body       *ast.BlockStatement
	Env        *Environment
}

func (f *Function) Type() ObjectType { return FUNCTION_OBJ }
func (f *Function) Inspect() string {
	var out bytes.Buffer

	params := []string{}
	for _, p := range f.Parameters {
		params = append(params, p.String())
	}

	out.WriteString("fn")
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(") {\n")
	out.WriteString(f.Body.String())
	out.WriteString("\n}")

	return out.String()
}

// String 문자열 리터럴을 평가하기 위한 객체
type String struct {
	Value string
}

func (s *String) Type() ObjectType { return STRING_OBJ }
func (s *String) Inspect() string  { return s.Value }
func (s *String) HashKey() HashKey {
	h := fnv.New64a()
	h.Write([]byte(s.Value))

	return HashKey{Type: s.Type(), Value: h.Sum64()}
}

// Builtin 내장 함수를 표현하는 타입
type Builtin struct {
	Fn BuiltinFunction
}

func (b *Builtin) Type() ObjectType { return BUILTIN_OBJ }
func (b *Builtin) Inspect() string  { return "builtin function" }

// Array 배열 리터럴을 평가하기 위한 객체
type Array struct {
	Elements []Object
}

func (ao *Array) Type() ObjectType { return ARRAY_OBJ }
func (ao *Array) Inspect() string {
	var out bytes.Buffer

	elements := []string{}
	for _, e := range ao.Elements {
		elements = append(elements, e.Inspect())
	}

	out.WriteString("[")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString("]")

	return out.String()
}

// HashPair 맵 자료형의 값을 표현하는 객체
// monkey 언에서 map 은 map[object.HashKey]object.HashPair 로 표현되는데,
// key: value 가 아닌 hashkey: hashpair 로 표현하는 이유는
// key 가 겉보기에는 같더라도 실제로는 메모리상 다른 주소를 가진 값일 수 있기 때문이다. 이 경우 제대로 value 를 찾아낼 수 없다.
// 그렇기에 hashkey 를 통해 hashpair 를 찾고 겉보기의 key 값을 hashpair 에 저장한다.
type HashPair struct {
	Key   Object
	Value Object
}

// Hash 해시 리터럴을 평가하기 위한 객체
type Hash struct {
	Pairs map[HashKey]HashPair
}

func (h *Hash) Type() ObjectType { return HASH_OBJ }
func (h *Hash) Inspect() string {
	var out bytes.Buffer

	pairs := []string{}
	for _, pair := range h.Pairs {
		pairs = append(pairs, fmt.Sprintf("%s: %s",
			pair.Key.Inspect(), pair.Value.Inspect()))
	}

	out.WriteString("{")
	out.WriteString(strings.Join(pairs, ", "))
	out.WriteString("}")

	return out.String()
}
