package utils

import "go.uber.org/dig"

var registrars = []func(*dig.Container){}

func AddRegistrar(f func(*dig.Container)) {
	registrars = append(registrars, f)
}

func Register(di *dig.Container) {
	for _, f := range registrars {
		f(di)
	}
}

func InvokeDI(c *dig.Container, function any, opts ...dig.InvokeOption) {
	err := c.Invoke(function, opts...)
	if err != nil {
		panic(err)
	}
}

func RegisterDI(c *dig.Container, constructor any, opts ...dig.ProvideOption) {
	err := c.Provide(constructor, opts...)
	if err != nil {
		panic(err)
	}
}
