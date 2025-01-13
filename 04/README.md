# Monkey Interpreter 04

이 프로젝트에서는 기존에 구현했던 Monkey 언어의 인터프리터에 새로운 자료형을 추가하여 확장한다.

* String
* Array

하나의 자료형을 추가할 때 거치게 되는 과정

1. Lexer 에 새로운 Token 타입 추가
2. Parser 에 새로운 자료형을 파싱하는 코드 추가 (AST 노드 타입 추가, 파서에 새로운 parseFunction register)
3. Eval 에 새로운 자료형을 처리하는 코드 추가 (Object 타입 추가, eval 함수에 새로운 case 추가)