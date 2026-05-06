package codec

import (
	"reflect"
	"testing"

	test "github.com/cmd-stream/codec-go/test"
	asserterror "github.com/ymz-ncnk/assert/error"
)

func TestRegistry(t *testing.T) {
	r := NewRegistry(
		WithCmd[any, test.Cmd1](),
		WithCmd[any, test.Cmd2](),
		WithResult[any, test.Result1](),
	)
	asserterror.Equal(t, reflect.TypeFor[test.Cmd1](), r.cmds[0])
	asserterror.Equal(t, reflect.TypeFor[test.Cmd2](), r.cmds[1])
	asserterror.Equal(t, reflect.TypeFor[test.Result1](), r.results[0])
}
