# Monkey Interpreter 03

이 프로젝트에서는 파서로 만든 AST 를 번역하는 트리 순회 인터프리터를 구현하여 
symbol 에 실제 의미를 담는다.

### Closure
```plaintext
let newAdder = fn(x) { fn(y) { x + y } };
let addTwo = newAdder(2);
addTwo(3); // 5
let addThree = newAdder(3);
addThree(10); // 13
```

`enclosedEnvironment` 를 이용하여 클로저를 구현할 수 있다. 위 코드를 보자.

`newAdder` 는 함수를 반환한다는 면에서 고차 함수이다. 
`addTwo` 에 인수 2를 넘겨 호출했을 때 `addTwo` 에는 반환한 클로저가 바인딩 되어 있다. (즉 `addTwo` 는 클로저이다.)

클로저라는 특별한 명칭이 붙은 이유는 
addTwo 가 호출되었을 때 'y' 인 3에 접근할 수 있는 것뿐만 아니라 `addTwo` 가 정의되었을 당시의 `x` 인 2에도 접근할 수 있기 때문이다.
`x` 가 `addTwo` 의 스코프에서 한참을 벗어나 있고 이론적으로 `addTwo` 를 정의했던 스코프는 사라졌지만,
여전히 `x` 에 접근할 수 있는 것이다.
즉, 정의된 당시의 환경에 접근할 수 있다는 것이고 여기서 '정의된 당시'는 newAdder 함수 몸체의 마지막 행이 평가되었을 때이다.
마지막 행은 함수 리터럴이므로 이것이 평가될 대에는 object.Function 을 만들어낼 것이고 이때 현재 환경에 대한 참조를 .Env 필드에 저장했다.
나중에 addTwo 의 몸체를 평가할 때 이 object.Function 객체의 Env 에서 환경을 꺼내어 확장한 다음 확장된 환경을 사용하여 addTwo 를 평가한다.
결국 정의한 당시에 사용했던 환경에 접근할 수 있는 것이다.

