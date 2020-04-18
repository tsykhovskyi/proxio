package proxio

type Locator struct {
	services map[string]interface{}
}

func (l Locator) Add(alias string, service interface{}) {
	l.services[alias] = service
}

func (l Locator) Get(alias string) interface{} {
	return l.services[alias]
}

func NewLocator() Locator {
	l := &Locator{}
	l.services = make(map[string]interface{})

	return *l
}
