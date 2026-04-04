package codec_test

import (
	"errors"
	"reflect"
	"testing"

	tmock "github.com/cmd-stream/cmd-stream-go/test/mock/transport"
	"github.com/cmd-stream/codec-generic-go"
	test "github.com/cmd-stream/codec-generic-go/test"
	"github.com/cmd-stream/codec-generic-go/test/mock"
	com "github.com/mus-format/common-go"
	assertfatal "github.com/ymz-ncnk/assert/fatal"
)

func TestCodec(t *testing.T) {
	t.Run("Encoding should work", func(t *testing.T) {
		var (
			wantDTM = 0
			wantBs  = []byte{1, 2, 3}
			wantLen = len(wantBs)
			wantN   = 1 + 1 + wantLen
		)

		ser := mock.NewSerializer[test.MyInterface, test.MyInterface]().RegisterMarshal(
			func(t test.MyInterface) ([]byte, error) {
				return wantBs, nil
			})
		c := codec.NewCodec(
			[]reflect.Type{
				reflect.TypeFor[test.MyStruct1](),
				reflect.TypeFor[test.MyStruct2](),
			},
			[]reflect.Type{
				reflect.TypeFor[test.MyStruct1](),
				reflect.TypeFor[test.MyStruct2](),
			},
			ser,
		)

		w := tmock.NewWriter().RegisterWriteByte(func(b byte) error {
			assertfatal.Equal(t, b, byte(wantDTM))
			return nil
		}).RegisterWriteByte(func(b byte) error {
			assertfatal.Equal(t, b, byte(wantLen))
			return nil
		}).RegisterWrite(func(p []byte) (n int, err error) {
			assertfatal.EqualDeep(t, p, wantBs)
			return len(p), nil
		})

		n, err := c.Encode(test.MyStruct1{}, w)
		assertfatal.EqualError(t, err, nil)
		assertfatal.Equal(t, n, wantN)
	})

	t.Run("Failed to marshal DTM", func(t *testing.T) {
		c := codec.NewCodec[test.MyInterface, test.MyInterface](
			[]reflect.Type{
				reflect.TypeFor[test.MyStruct1](),
				reflect.TypeFor[test.MyStruct2](),
			},
			[]reflect.Type{
				reflect.TypeFor[test.MyStruct1](),
				reflect.TypeFor[test.MyStruct2](),
			},
			nil,
		)

		writeErr := errors.New("failed to write DTM")
		wantErr := codec.NewFailedToMarshalDTM(writeErr)

		w := tmock.NewWriter().RegisterWriteByte(func(b byte) error {
			return writeErr
		})
		n, err := c.Encode(test.MyStruct1{}, w)
		assertfatal.EqualError(t, err, wantErr)
		assertfatal.Equal(t, n, 0)
	})

	t.Run("Decoding should work", func(t *testing.T) {
		wantDTM := 1
		wantV := test.MyStruct2{Y: "hello"}
		wantBs := []byte{1, 2, 3}
		wantLen := len(wantBs)
		wantN := 1 + 1 + wantLen

		ser := mock.NewSerializer[test.MyInterface, test.MyInterface]().RegisterUnmarshal(
			func(bs []byte, v test.MyInterface) error {
				assertfatal.EqualDeep(t, bs, wantBs)

				rv := reflect.ValueOf(v)
				if rv.Kind() == reflect.Ptr && !rv.IsNil() {
					rv.Elem().Set(reflect.ValueOf(wantV))
				}

				return nil
			},
		)
		c := codec.NewCodec(
			[]reflect.Type{
				reflect.TypeFor[test.MyStruct1](),
				reflect.TypeFor[test.MyStruct2](),
			},
			[]reflect.Type{
				reflect.TypeFor[test.MyStruct1](),
				reflect.TypeFor[test.MyStruct2](),
			},
			ser,
		)

		r := tmock.NewReader().RegisterReadByte(func() (b byte, err error) {
			return byte(wantDTM), nil
		}).RegisterReadByte(func() (b byte, err error) {
			return byte(wantLen), nil
		}).RegisterRead(func(p []byte) (n int, err error) {
			copy(p, wantBs)
			return wantLen, nil
		})

		v, n, err := c.Decode(r)
		assertfatal.EqualError(t, err, nil)
		assertfatal.Equal(t, n, wantN)
		assertfatal.EqualDeep(t, v, wantV)
	})

	t.Run("Unrecognized type", func(t *testing.T) {
		c := codec.NewCodec[test.MyInterface, test.MyInterface](
			[]reflect.Type{
				reflect.TypeFor[test.MyStruct1](),
				reflect.TypeFor[test.MyStruct2](),
			},
			[]reflect.Type{
				reflect.TypeFor[test.MyStruct1](),
				reflect.TypeFor[test.MyStruct2](),
			},
			nil,
		)

		v := test.MyStruct3{Z: 3.14}
		wantType := reflect.TypeOf(v)
		wantErr := codec.NewUnrecognizedType(wantType)

		w := tmock.NewWriter()

		n, err := c.Encode(v, w)
		assertfatal.EqualError(t, err, wantErr)
		assertfatal.Equal(t, n, 0)
	})

	t.Run("Failed to marshal byte slice", func(t *testing.T) {
		ser := mock.NewSerializer[test.MyInterface, test.MyInterface]().RegisterMarshal(
			func(t test.MyInterface) ([]byte, error) {
				return []byte{}, nil
			},
		)
		c := codec.NewCodec(
			[]reflect.Type{
				reflect.TypeFor[test.MyStruct1](),
				reflect.TypeFor[test.MyStruct2](),
			},
			[]reflect.Type{
				reflect.TypeFor[test.MyStruct1](),
				reflect.TypeFor[test.MyStruct2](),
			},
			ser,
		)

		writeErr := errors.New("failed to write byte slice length")
		wantErr := codec.NewFailedToMarshalByteSlice(writeErr)

		w := tmock.NewWriter().RegisterWriteByte(func(b byte) error {
			return nil
		}).RegisterWriteByte(func(b byte) error {
			return writeErr
		})

		n, err := c.Encode(test.MyStruct1{}, w)
		assertfatal.EqualError(t, err, wantErr)
		assertfatal.Equal(t, n, 1)
	})

	t.Run("Failed to unmarshal DTM", func(t *testing.T) {
		c := codec.NewCodec[test.MyInterface, test.MyInterface](
			[]reflect.Type{
				reflect.TypeFor[test.MyStruct1](),
				reflect.TypeFor[test.MyStruct2](),
			},
			[]reflect.Type{
				reflect.TypeFor[test.MyStruct1](),
				reflect.TypeFor[test.MyStruct2](),
			},
			nil,
		)

		readErr := errors.New("failed to read DTM")
		wantErr := codec.NewFailedToUnmarshalDTM(readErr)

		r := tmock.NewReader().RegisterReadByte(func() (b byte, err error) {
			return 0, readErr
		})

		_, n, err := c.Decode(r)
		assertfatal.EqualError(t, err, wantErr)
		assertfatal.Equal(t, n, 0)
	})

	t.Run("Unrecognized DTM", func(t *testing.T) {
		c := codec.NewCodec[test.MyInterface, test.MyInterface](
			[]reflect.Type{
				reflect.TypeFor[test.MyStruct1](),
				reflect.TypeFor[test.MyStruct2](),
			},
			[]reflect.Type{
				reflect.TypeFor[test.MyStruct1](),
				reflect.TypeFor[test.MyStruct2](),
			},
			nil,
		)

		const unrecognizedDTM com.DTM = 99
		wantErr := codec.NewUnrecognizedDTM(unrecognizedDTM)

		r := tmock.NewReader().RegisterReadByte(func() (b byte, err error) {
			return byte(unrecognizedDTM), nil
		})

		_, n, err := c.Decode(r)
		assertfatal.EqualError(t, err, wantErr)
		assertfatal.Equal(t, n, 1)
	})

	t.Run("Failed to unmarshal byte slice", func(t *testing.T) {
		c := codec.NewCodec[test.MyInterface, test.MyInterface](
			[]reflect.Type{
				reflect.TypeFor[test.MyStruct1](),
				reflect.TypeFor[test.MyStruct2](),
			},
			[]reflect.Type{
				reflect.TypeFor[test.MyStruct1](),
				reflect.TypeFor[test.MyStruct2](),
			},
			nil,
		)

		readErr := errors.New("failed to read byte slice")
		wantErr := codec.NewFailedToUnmarshalByteSlice(readErr)

		r := tmock.NewReader().RegisterReadByte(func() (b byte, err error) {
			return 0, nil
		}).RegisterReadByte(func() (b byte, err error) {
			return 0, readErr
		})

		_, n, err := c.Decode(r)
		assertfatal.EqualError(t, err, wantErr)
		assertfatal.Equal(t, n, 1)
	})
}
