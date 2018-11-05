package js

var DefaultGlobal = NewGlobal()

type Global struct {
	properties map[string]interface{}
}

func NewGlobal() *Global {
	return &Global{
		properties: make(map[string]interface{}),
	}
}

func (g *Global) Register(name string, prop interface{}) {
	g.properties[name] = prop
}

func (g *Global) Get(name string) (interface{}, bool) {
	v, ok := g.properties[name]
	return v, ok
}

func Register(name string, prop interface{}) {
	DefaultGlobal.Register(name, prop)
}
