package test

import (
	"context"
	"fmt"
	"time"

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

func (c Cmd1) Exec(ctx context.Context, seq core.Seq, at time.Time,
	_ struct{}, proxy core.Proxy,
) (err error) {
	return
}

type Cmd2 struct{ A, B int }

func (c Cmd2) Exec(ctx context.Context, seq core.Seq, at time.Time,
	_ struct{}, proxy core.Proxy,
) (err error) {
	return
}

type Result1 int

func (r Result1) LastOne() bool { return true }
