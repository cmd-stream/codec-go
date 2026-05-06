package test

import (
	"bufio"
	"bytes"
	"reflect"
	"testing"

	"github.com/cmd-stream/cmd-stream-go/core"
	"github.com/cmd-stream/cmd-stream-go/transport"
	assertfatal "github.com/ymz-ncnk/assert/fatal"
)

// Codec represents a generic codec interface for testing.
type Codec[T, V any] interface {
	Encode(t T, w transport.Writer) (n int, err error)
	Decode(r transport.Reader) (v V, n int, err error)
}

// TransportWriter is a utility for testing codecs.
type TransportWriter struct {
	*bufio.Writer
}

func (w *TransportWriter) Flush() error {
	return w.Writer.Flush()
}

// VerifyRoundTripCmd verifies the round-trip encoding and decoding of a Command.
func VerifyRoundTripCmd(t *testing.T, clientCodec Codec[core.Cmd[any], core.Result],
	serverCodec Codec[core.Result, core.Cmd[any]], cmd core.Cmd[any]) {
	VerifyRoundTripCmdWith(t, clientCodec, serverCodec, cmd, func(a, b any) bool {
		return reflect.DeepEqual(a, b)
	})
}

// VerifyRoundTripCmdWith verifies the round-trip encoding and decoding of a
// Command using a custom equality check.
func VerifyRoundTripCmdWith(t *testing.T, clientCodec Codec[core.Cmd[any], core.Result],
	serverCodec Codec[core.Result, core.Cmd[any]],
	cmd core.Cmd[any],
	equal func(expected, actual any) bool,
) {
	var (
		buf = &bytes.Buffer{}
		bw  = bufio.NewWriter(buf)
		w   = &TransportWriter{Writer: bw}
	)

	n, err := clientCodec.Encode(cmd, w)
	assertfatal.EqualError(t, err, nil, "failed to encode")
	w.Flush()

	r := bytes.NewReader(buf.Bytes())
	decoded, n2, err := serverCodec.Decode(r)
	assertfatal.EqualError(t, err, nil, "failed to decode")
	assertfatal.Equal(t, n, n2, "encoded and decoded bytes length mismatch")
	if !equal(cmd, decoded) {
		t.Fatalf("roundtrip failed: expected %v, got %v", cmd, decoded)
	}
}

// VerifyRoundTripResult verifies the round-trip encoding and decoding of a Result.
func VerifyRoundTripResult(t *testing.T, clientCodec Codec[core.Cmd[any], core.Result],
	server Codec[core.Result, core.Cmd[any]],
	res core.Result,
) {
	VerifyRoundTripResultWith(t, clientCodec, server, res, func(a, b any) bool {
		return reflect.DeepEqual(a, b)
	})
}

// VerifyRoundTripResultWith verifies the round-trip encoding and decoding of a
// Result using a custom equality check.
func VerifyRoundTripResultWith(t *testing.T, clientCodec Codec[core.Cmd[any], core.Result],
	serverCodec Codec[core.Result, core.Cmd[any]],
	res core.Result,
	equal func(expected, actual any) bool,
) {
	var (
		buf = &bytes.Buffer{}
		bw  = bufio.NewWriter(buf)
		w   = &TransportWriter{Writer: bw}
	)

	n, err := serverCodec.Encode(res, w)
	assertfatal.EqualError(t, err, nil, "failed to encode")
	w.Flush()

	r := bytes.NewReader(buf.Bytes())
	decoded, n2, err := clientCodec.Decode(r)
	assertfatal.EqualError(t, err, nil, "failed to decode")
	assertfatal.Equal(t, n, n2, "encoded and decoded bytes length mismatch")
	if !equal(res, decoded) {
		t.Fatalf("roundtrip failed: expected %v, got %v", res, decoded)
	}
}

// FuzzDecode ensures that the codec does not panic when decoding random data.
func FuzzDecode[T, V any](codec Codec[T, V], data []byte) {
	r := bytes.NewReader(data)
	_, _, _ = codec.Decode(r)
}
