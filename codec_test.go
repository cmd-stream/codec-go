package codec_test

import (
	"errors"
	"reflect"
	"testing"

	cmock "github.com/cmd-stream/cmd-stream-go/test/mock"
	"github.com/cmd-stream/codec-go"
	test "github.com/cmd-stream/codec-go/test"
	"github.com/cmd-stream/codec-go/test/mock"
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

		ser := mock.NewSerializer[test.Interface, test.Interface]().RegisterMarshal(
			func(t test.Interface) ([]byte, error) {
				return wantBs, nil
			})
		c := codec.NewCodec(
			[]reflect.Type{
				reflect.TypeFor[test.Struct1](),
				reflect.TypeFor[test.Struct2](),
			},
			[]reflect.Type{
				reflect.TypeFor[test.Struct1](),
				reflect.TypeFor[test.Struct2](),
			},
			ser,
		)

		w := cmock.NewWriter().RegisterWriteByte(func(b byte) error {
			assertfatal.Equal(t, b, byte(wantDTM))
			return nil
		}).RegisterWriteByte(func(b byte) error {
			assertfatal.Equal(t, b, byte(wantLen))
			return nil
		}).RegisterWrite(func(p []byte) (n int, err error) {
			assertfatal.EqualDeep(t, p, wantBs)
			return len(p), nil
		})

		n, err := c.Encode(test.Struct1{}, w)
		assertfatal.EqualError(t, err, nil)
		assertfatal.Equal(t, n, wantN)
	})

	t.Run("Failed to marshal DTM", func(t *testing.T) {
		c := codec.NewCodec[test.Interface, test.Interface](
			[]reflect.Type{
				reflect.TypeFor[test.Struct1](),
				reflect.TypeFor[test.Struct2](),
			},
			[]reflect.Type{
				reflect.TypeFor[test.Struct1](),
				reflect.TypeFor[test.Struct2](),
			},
			nil,
		)

		writeErr := errors.New("failed to write DTM")
		wantErr := codec.NewFailedToMarshalDTM(writeErr)

		w := cmock.NewWriter().RegisterWriteByte(func(b byte) error {
			return writeErr
		})
		n, err := c.Encode(test.Struct1{}, w)
		assertfatal.EqualError(t, err, wantErr)
		assertfatal.Equal(t, n, 0)
	})

	t.Run("Decoding should work", func(t *testing.T) {
		wantDTM := 1
		wantV := test.Struct2{Y: "hello"}
		wantBs := []byte{1, 2, 3}
		wantLen := len(wantBs)
		wantN := 1 + 1 + wantLen

		ser := mock.NewSerializer[test.Interface, test.Interface]().RegisterUnmarshal(
			func(bs []byte, v test.Interface) error {
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
				reflect.TypeFor[test.Struct1](),
				reflect.TypeFor[test.Struct2](),
			},
			[]reflect.Type{
				reflect.TypeFor[test.Struct1](),
				reflect.TypeFor[test.Struct2](),
			},
			ser,
		)

		r := cmock.NewReader().RegisterReadByte(func() (b byte, err error) {
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
		c := codec.NewCodec[test.Interface, test.Interface](
			[]reflect.Type{
				reflect.TypeFor[test.Struct1](),
				reflect.TypeFor[test.Struct2](),
			},
			[]reflect.Type{
				reflect.TypeFor[test.Struct1](),
				reflect.TypeFor[test.Struct2](),
			},
			nil,
		)

		v := test.Struct3{Z: 3.14}
		wantType := reflect.TypeOf(v)
		wantErr := codec.NewUnrecognizedType(wantType)

		w := cmock.NewWriter()

		n, err := c.Encode(v, w)
		assertfatal.EqualError(t, err, wantErr)
		assertfatal.Equal(t, n, 0)
	})

	t.Run("Failed to marshal byte slice", func(t *testing.T) {
		ser := mock.NewSerializer[test.Interface, test.Interface]().RegisterMarshal(
			func(t test.Interface) ([]byte, error) {
				return []byte{}, nil
			},
		)
		c := codec.NewCodec(
			[]reflect.Type{
				reflect.TypeFor[test.Struct1](),
				reflect.TypeFor[test.Struct2](),
			},
			[]reflect.Type{
				reflect.TypeFor[test.Struct1](),
				reflect.TypeFor[test.Struct2](),
			},
			ser,
		)

		writeErr := errors.New("failed to write byte slice length")
		wantErr := codec.NewFailedToMarshalByteSlice(writeErr)

		w := cmock.NewWriter().RegisterWriteByte(func(b byte) error {
			return nil
		}).RegisterWriteByte(func(b byte) error {
			return writeErr
		})

		n, err := c.Encode(test.Struct1{}, w)
		assertfatal.EqualError(t, err, wantErr)
		assertfatal.Equal(t, n, 1)
	})

	t.Run("Failed to unmarshal DTM", func(t *testing.T) {
		c := codec.NewCodec[test.Interface, test.Interface](
			[]reflect.Type{
				reflect.TypeFor[test.Struct1](),
				reflect.TypeFor[test.Struct2](),
			},
			[]reflect.Type{
				reflect.TypeFor[test.Struct1](),
				reflect.TypeFor[test.Struct2](),
			},
			nil,
		)

		readErr := errors.New("failed to read DTM")
		wantErr := codec.NewFailedToUnmarshalDTM(readErr)

		r := cmock.NewReader().RegisterReadByte(func() (b byte, err error) {
			return 0, readErr
		})

		_, n, err := c.Decode(r)
		assertfatal.EqualError(t, err, wantErr)
		assertfatal.Equal(t, n, 0)
	})

	t.Run("Unrecognized DTM", func(t *testing.T) {
		c := codec.NewCodec[test.Interface, test.Interface](
			[]reflect.Type{
				reflect.TypeFor[test.Struct1](),
				reflect.TypeFor[test.Struct2](),
			},
			[]reflect.Type{
				reflect.TypeFor[test.Struct1](),
				reflect.TypeFor[test.Struct2](),
			},
			nil,
		)

		const unrecognizedDTM com.DTM = 99
		wantErr := codec.NewUnrecognizedDTM(unrecognizedDTM)

		r := cmock.NewReader().RegisterReadByte(func() (b byte, err error) {
			return byte(unrecognizedDTM), nil
		})

		_, n, err := c.Decode(r)
		assertfatal.EqualError(t, err, wantErr)
		assertfatal.Equal(t, n, 1)
	})

	t.Run("Failed to unmarshal byte slice", func(t *testing.T) {
		c := codec.NewCodec[test.Interface, test.Interface](
			[]reflect.Type{
				reflect.TypeFor[test.Struct1](),
				reflect.TypeFor[test.Struct2](),
			},
			[]reflect.Type{
				reflect.TypeFor[test.Struct1](),
				reflect.TypeFor[test.Struct2](),
			},
			nil,
		)

		readErr := errors.New("failed to read byte slice")
		wantErr := codec.NewFailedToUnmarshalByteSlice(readErr)

		r := cmock.NewReader().RegisterReadByte(func() (b byte, err error) {
			return 0, nil
		}).RegisterReadByte(func() (b byte, err error) {
			return 0, readErr
		})

		_, n, err := c.Decode(r)
		assertfatal.EqualError(t, err, wantErr)
		assertfatal.Equal(t, n, 1)
	})
}
