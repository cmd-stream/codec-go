package test

import (
	"context"
	"fmt"

	"github.com/cmd-stream/cmd-stream-go/core"
)

type Interface interface {
	Print()
}

type Struct1 struct {
	X int
}

func (s Struct1) Print() {
	fmt.Println("MyStruct1")
}

type Struct2 struct {
	Y string
}

func (s Struct2) Print() {
	fmt.Println("MyStruct2")
}

type Struct3 struct {
	Z float64
}

func (s Struct3) Print() {
	fmt.Println("MyStruct3")
}

type Cmd1 struct{ A, B int }

func (c Cmd1) Exec(ctx context.Context, _ struct{}, proxy core.Proxy) error {
	return nil
}

type Cmd2 string

func (c Cmd2) Exec(ctx context.Context, _ struct{}, proxy core.Proxy) error {
	return nil
}

type Result1 int

func (r Result1) LastOne() bool { return true }

type Result2 struct {
	Y string
}

func (r Result2) LastOne() bool {
	return true
}
