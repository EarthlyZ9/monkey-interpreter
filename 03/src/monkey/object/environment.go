package object

func NewEnclosedEnvironment(outer *Environment) *Environment {
	env := NewEnvironment()
	env.outer = outer
	return env
}

func NewEnvironment() *Environment {
	s := make(map[string]Object)
	return &Environment{store: s, outer: nil}
}

// Environment 는 식별자와 값을 매핑하는 구조체이다.
// 바깥쪽 스코프는 안쪽 스코프를 감싸고 안쪽 스코프는 바깥쪽 스코프를 확장하는 모양새가 된다.
type Environment struct {
	store map[string]Object
	outer *Environment
}

// Get 함수는 주어진 이름에 해당하는 값을 찾아 반환한다.
// 이때 이름이 존재하지 않으면 outer 환경으로 이동하여 찾는다.
func (e *Environment) Get(name string) (Object, bool) {
	obj, ok := e.store[name]
	if !ok && e.outer != nil {
		obj, ok = e.outer.Get(name)
	}
	return obj, ok
}

// Set 함수는 현재 환경 (store) 에 값을 저장한다.
func (e *Environment) Set(name string, val Object) Object {
	e.store[name] = val
	return val
}
