package codec

import (
	"reflect"

	"github.com/cmd-stream/cmd-stream-go/core"
)

type Registry[T any] struct {
	cmds    []reflect.Type
	results []reflect.Type
}

func (r *Registry[T]) Cmds() []reflect.Type {
	return r.cmds
}

func (r *Registry[T]) Results() []reflect.Type {
	return r.results
}

func NewRegistry[T any](opts ...func(*Registry[T])) *Registry[T] {
	r := &Registry[T]{}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

func WithCmd[T any, C core.Cmd[T]]() func(*Registry[T]) {
	return func(r *Registry[T]) {
		r.cmds = append(r.cmds, reflect.TypeFor[C]())
	}
}

func WithResult[T any, R core.Result]() func(*Registry[T]) {
	return func(r *Registry[T]) {
		r.results = append(r.results, reflect.TypeFor[R]())
	}
}
