// Package codec provides a generic, type-safe codec abstraction used for
// encoding and decoding values in the cmd-stream ecosystem.
package codec

import (
	"reflect"

	tspt "github.com/cmd-stream/cmd-stream-go/transport"
	com "github.com/mus-format/common-go"
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
	ser Serializer[T, V],
) (codec Codec[T, V]) {
	return newCodec(types1, types2, ser, decodeValue)
}

// NewCodecWithDecoder constructs a Codec using a custom value decoder
// function.
//
// decodeValueFn allows overriding the default decoding behavior.
func NewCodecWithDecoder[T, V any](types1 []reflect.Type, types2 []reflect.Type,
	ser Serializer[T, V],
	decodeValueFn DecodeValueFn[T, V],
) (codec Codec[T, V]) {
	return newCodec(types1, types2, ser, decodeValueFn)
}

func newCodec[T, V any](types1 []reflect.Type, types2 []reflect.Type,
	ser Serializer[T, V],
	decodeValueFn DecodeValueFn[T, V],
) (codec Codec[T, V]) {
	if len(types1) == 0 {
		panic("codecgnrc:" + "types1 is empty")
	}
	if len(types2) == 0 {
		panic("codecgnrc:" + "types2 is empty")
	}
	codec = Codec[T, V]{
		typeMap:       make(map[reflect.Type]com.DTM),
		dtmSl:         make([]reflect.Type, len(types2)),
		ser:           ser,
		decodeValueFn: decodeValueFn,
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
	n1, err := ord.ByteSlice.Marshal(bs, w)
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
	bs, n1, err := ord.ByteSlice.Unmarshal(r)
	n += n1
	if err != nil {
		err = NewFailedToUnmarshalByteSlice(err)
		return
	}
	v, err = c.decodeValueFn(tp, c.ser, bs)
	return
}
