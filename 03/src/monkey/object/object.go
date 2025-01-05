package object

import (
	"bytes"
	"fmt"
	"monkey/ast"
	"strings"
)

type ObjectType string

const (
	NULL_OBJ  = "NULL"
	ERROR_OBJ = "ERROR"

	INTEGER_OBJ = "INTEGER"
	BOOLEAN_OBJ = "BOOLEAN"

	RETURN_VALUE_OBJ = "RETURN_VALUE"

	FUNCTION_OBJ = "FUNCTION"
)

// Object 소스코드를 평가하면서 확인하는 모든 값은 Object 인터페이스로 표현한다.
type Object interface {
	Type() ObjectType
	Inspect() string
}

// Integer 는 정수 값을 나타내는 객체이다.
// 파서가 정수 리터럴을 만나면 우선 ast.IntegerLiteral 노드를 생성할 것이다.
// 그리고 나서 AST 를 평가할 때에는 ast.IntegerLiteral 노드를 평가하여 Integer 객체를 생성할 것이다.
// Integer 의 Value 필드는 전달받은 *ast.IntegerLiteral 노드의 Value 필드와 같아야 한다.
type Integer struct {
	Value int64
}

func (i *Integer) Type() ObjectType { return INTEGER_OBJ }
func (i *Integer) Inspect() string  { return fmt.Sprintf("%d", i.Value) }

type Boolean struct {
	Value bool
}

func (b *Boolean) Type() ObjectType { return BOOLEAN_OBJ }
func (b *Boolean) Inspect() string  { return fmt.Sprintf("%t", b.Value) }

type Null struct{}

func (n *Null) Type() ObjectType { return NULL_OBJ }
func (n *Null) Inspect() string  { return "null" }

// ReturnValue 는 반환값을 나타내는 객체이다.
// ReturnValue 는 단순히 다른 Object 를 감싸고 있는 Wrapper 로,
// 반환값임을 표현할 뿐이며 프로그램은 ReturnValue 를 만나면 그 내부의 Object 를 꺼내어 반환한다.
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
