# Monkey Interpreter 02
이 프로젝트에서는 추상구문트리를 정의하고 렉서가 생성한 토큰들을 추상구문트리로 표현하는 파서를 구현한다.

### AST 와 우선순위 테이블
우선순위 테이블을 이용하여 연산자의 우선순위를 정의하고 AST 를 생성할 때 핵심이 되는
컨셉은 더 높은 우선순위를 가진 연산자를 가진 표현식이 트리상 더 깊은 곳에 위치하게 만드는 것이다.

이를 달성하기 위해 `parseExpression` 함수를 호출할 때 현재의 우선순위를 넘기는 것이다.
그렇기 때문에 표현식을 파싱하기 시작하는 시점에는 우선순위가 `LOWEST` 로 시작하게 된다.

> `parseExpression` 이 호출될 때, `precedence` 값은 `parseExpression` 매서드를 호출하는 시점에서
> 갖게 되는 '오른쪽으로 묶이는 힘(right-binding power)' 을 나타낸다.

#### Right Binding Power 
RBP 가 강할 수록 현재의 표현식 오른쪽에 더 많은 토큰/피연산자/연산자를 묶을 수 있다.
RBP 가 최대라면,
즉, 현재 precedence 가 가장 높아서 다음 토큰의 precedence 보다 높고 `parseExpression` 의 for 문 조건을 만족하지 않는다면
더 이상 표현식을 파싱할 수 없게 된다.

반대로 생각하면 Left Binding Power 는 결국 `peekPrecedence` 가 될 것이다.

간단하게 정리하면, RBP 보다 LBP 가 높다면 그 시점까지 파싱한 노드가 다음 연산자에 의해 
왼쪽에서 오른쪽으로 빨려 들어간다.