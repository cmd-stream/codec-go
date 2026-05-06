// Package codec provides a generic, type-safe codec abstraction used for
// encoding and decoding values in the cmd-stream ecosystem.
package codec

import (
	"reflect"

	tspt "github.com/cmd-stream/cmd-stream-go/transport"
	com "github.com/mus-format/common-go"
	"github.com/mus-format/mus-stream-go"
	bslopts "github.com/mus-format/mus-stream-go/options/byte_slice"
	"github.com/mus-format/mus-stream-go/ord"
	"github.com/mus-format/mus-stream-go/typed"
)

// NewCodec constructs a Codec with default value decoder logic.
//
// Parameters:
//   - types1 lists the Go types that can be encoded.
//   - types2 lists the Go types that can be decoded.
//   - ser is the serializer used for encoding/decoding values.
func NewCodec[T, V any](types1 []reflect.Type, types2 []reflect.Type,
	ser Serializer[T, V], opts ...SetOption,
) (codec Codec[T, V]) {
	return newCodec(types1, types2, ser, decodeValue, opts...)
}

// NewCodecWithDecoder constructs a Codec using a custom value decoder
// function.
//
// decodeValueFn allows overriding the default decoding behavior.
func NewCodecWithDecoder[T, V any](types1 []reflect.Type, types2 []reflect.Type,
	ser Serializer[T, V],
	decodeValueFn DecodeValueFn[T, V],
	opts ...SetOption,
) (codec Codec[T, V]) {
	return newCodec(types1, types2, ser, decodeValueFn, opts...)
}

func newCodec[T, V any](types1 []reflect.Type, types2 []reflect.Type,
	ser Serializer[T, V],
	decodeValueFn DecodeValueFn[T, V],
	opts ...SetOption,
) (codec Codec[T, V]) {
	if len(types1) == 0 {
		panic("codecgnrc:" + "types1 is empty")
	}
	if len(types2) == 0 {
		panic("codecgnrc:" + "types2 is empty")
	}
	o := Options{}
	Apply(opts, &o)
	codec = Codec[T, V]{
		typeMap:       make(map[reflect.Type]com.DTM),
		dtmSl:         make([]reflect.Type, len(types2)),
		ser:           ser,
		decodeValueFn: decodeValueFn,
		bslSer:        newByteSliceSer(o.maxLen),
	}
	for i, t := range types1 {
		codec.typeMap[t] = com.DTM(i)
	}
	copy(codec.dtmSl, types2)
	return
}

// Codec represents a generic type-safe codec for encoding and decoding values.
// T is the type used for encoding, V is the type used for decoding.
type Codec[T, V any] struct {
	typeMap       map[reflect.Type]com.DTM
	dtmSl         []reflect.Type
	ser           Serializer[T, V]
	decodeValueFn DecodeValueFn[T, V]
	bslSer        mus.Serializer[[]byte]
}

// Encode writes a value of type T to the given transport.Writer.
// Returns the total number of bytes written and any error.
func (c Codec[T, V]) Encode(t T, w tspt.Writer) (n int, err error) {
	tp := reflect.TypeOf(t)
	dtm, pst := c.typeMap[tp]
	if !pst {
		err = NewUnrecognizedType(tp)
		return
	}
	n, err = typed.DTMSer.Marshal(dtm, w)
	if err != nil {
		err = NewFailedToMarshalDTM(err)
		return
	}
	bs, err := c.ser.Marshal(t)
	if err != nil {
		err = NewFailedToMarshalValue(t, err)
		return
	}
	n1, err := c.bslSer.Marshal(bs, w)
	n += n1
	if err != nil {
		err = NewFailedToMarshalByteSlice(err)
	}
	return
}

// Decode reads a value of type V from the given transport.Reader.
// Returns the decoded value, total bytes read, and any error.
func (c Codec[T, V]) Decode(r tspt.Reader) (v V, n int, err error) {
	dtm, n, err := typed.DTMSer.Unmarshal(r)
	if err != nil {
		err = NewFailedToUnmarshalDTM(err)
		return
	}
	if dtm < 0 || dtm >= com.DTM(len(c.dtmSl)) {
		err = NewUnrecognizedDTM(dtm)
		return
	}
	tp := c.dtmSl[dtm]
	bs, n1, err := c.bslSer.Unmarshal(r)
	n += n1
	if err != nil {
		err = NewFailedToUnmarshalByteSlice(err)
		return
	}
	v, err = c.decodeValueFn(tp, c.ser, bs)
	return
}

func newByteSliceSer(maxLen int) mus.Serializer[[]byte] {
	if maxLen > 0 {
		return ord.NewValidByteSliceSer(bslopts.WithLenValidator(
			com.ValidatorFn[int](func(length int) error {
				if length > maxLen {
					return com.ErrTooLargeLength
				}
				return nil
			}),
		))
	}
	return ord.ByteSlice
}
