package evaluator

import (
	"fmt"
	"monkey/object"
)

// 내장함수를 모아둔 map 객체
var builtins = map[string]*object.Builtin{
	"len": &object.Builtin{Fn: func(args ...object.Object) object.Object {
		// 내장 함수 len 은 하나의 인자만을 받을 수 있다
		if len(args) != 1 {
			return newError("wrong number of arguments. got=%d, want=1",
				len(args))
		}

		switch arg := args[0].(type) {
		case *object.Array:
			return &object.Integer{Value: int64(len(arg.Elements))}
		case *object.String:
			// String 타입이라면 len 은 문자열의 길이를 반환
			return &object.Integer{Value: int64(len(arg.Value))}
		default:
			// 내장 함수 len 이 지원하는 타입은 String, Array 뿐이다
			return newError("argument to `len` not supported, got %s",
				args[0].Type())
		}
	},
	},
	"puts": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			for _, arg := range args {
				fmt.Println(arg.Inspect())
			}

			return NULL
		},
	},
	// 내장함수 first 는 배열을 받아 첫 번째 요소를 반환한다
	"first": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			// 내장함수 first 는 하나의 인자만을 받을 수 있다
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1",
					len(args))
			}
			// 인자로 받은 객체가 배열 타입인지 확인
			if args[0].Type() != object.ARRAY_OBJ {
				return newError("argument to `first` must be ARRAY, got %s",
					args[0].Type())
			}

			arr := args[0].(*object.Array)
			if len(arr.Elements) > 0 {
				return arr.Elements[0]
			}
			// 인자로 받은 배열이 비어있는 경우 NULL 반환
			return NULL
		},
	},
	// 내장함수 last 는 배열을 받아 마지막 요소를 반환한다
	"last": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1",
					len(args))
			}
			if args[0].Type() != object.ARRAY_OBJ {
				return newError("argument to `last` must be ARRAY, got %s",
					args[0].Type())
			}

			arr := args[0].(*object.Array)
			length := len(arr.Elements)
			if length > 0 {
				return arr.Elements[length-1]
			}
			// 인자로 받은 배열이 비어있는 경우 NULL 반환
			return NULL
		},
	},
	// 내장함수 rest 는 배열을 받아 첫 번째 요소를 제외한 나머지 요소를 반환한다
	"rest": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1",
					len(args))
			}
			if args[0].Type() != object.ARRAY_OBJ {
				return newError("argument to `rest` must be ARRAY, got %s",
					args[0].Type())
			}

			arr := args[0].(*object.Array)
			length := len(arr.Elements)
			//
			if length > 0 {
				newElements := make([]object.Object, length-1, length-1)
				// 새로 만든 배열에 기존 배열의 두 번째 요소부터 복사
				copy(newElements, arr.Elements[1:length])
				return &object.Array{Elements: newElements}
			}
			// 인자로 받은 배열이 비어있는 경우 NULL 반환
			return NULL
		},
	},
	// 내장함수 push 는 배열의 끝에 새로운 요소를 추가한다
	"push": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			// 내장함수 push 는 두 개의 인자를 받을 수 있다
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2",
					len(args))
			}
			// 첫번째 인자는 무조건 요소를 추가할 배열 타입이어야 한다.
			if args[0].Type() != object.ARRAY_OBJ {
				return newError("argument to `push` must be ARRAY, got %s",
					args[0].Type())
			}

			arr := args[0].(*object.Array)
			length := len(arr.Elements)

			// 기존 길이보다 1 큰 배열을 만들어 기존 요소를 복사하고 새로운 요소를 추가
			newElements := make([]object.Object, length+1, length+1)
			copy(newElements, arr.Elements)
			newElements[length] = args[1]

			return &object.Array{Elements: newElements}
		},
	},
}
