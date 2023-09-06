package serde

import (
	"encoding/binary"
	"fmt"
	"math"
	"unsafe"
)

// ID is the unique ID of a pointer in a serialized format.
type ID int64

// Deserializer contains the state of the deserializer.
type Deserializer struct {
	// TODO: make it a slice
	ptrs map[ID]unsafe.Pointer
}

func (d *Deserializer) ReadPtr(b []byte) (unsafe.Pointer, ID, []byte) {
	x, n := binary.Varint(b)
	i := ID(x)
	p := d.ptrs[i]
	return p, i, b[n:]
}

func (d *Deserializer) Store(i ID, p unsafe.Pointer) {
	if d.ptrs[i] != nil {
		panic(fmt.Errorf("trying to overwirte known ID %d with %p", i, p))
	}
	d.ptrs[i] = p
}

func EnsureDeserializer(d *Deserializer) *Deserializer {
	if d == nil {
		d = &Deserializer{
			ptrs: make(map[ID]unsafe.Pointer),
		}
	}
	return d
}

// Serializer contains the state of the serializer.
type Serializer struct {
	ptrs map[unsafe.Pointer]ID
}

func (s *Serializer) WritePtr(p unsafe.Pointer, b []byte) (bool, []byte) {
	if p == nil {
		return true, binary.AppendVarint(b, 0)
	}
	i, ok := s.ptrs[p]
	if !ok {
		i = ID(len(s.ptrs) + 1)
		s.ptrs[p] = i
	}
	return ok, binary.AppendVarint(b, int64(i))
}

func EnsureSerializer(s *Serializer) *Serializer {
	if s == nil {
		s = &Serializer{
			ptrs: make(map[unsafe.Pointer]ID),
		}
	}
	return s
}

// Serializable objects can be serialized to bytes.
type Serializable interface {
	// MarshalAppend marshals the object and appends the resulting bytes to
	// the provided buffer.
	MarshalAppend(b []byte) ([]byte, error)

	// Unmarshal unmarshals an object from a buffer. It returns the number
	// of bytes that were read from the buffer in order to reconstruct the
	// object.
	Unmarshal(b []byte) (n int, err error)
}

// Helpers to write composite types serializers and deserializers.

func SerializeSerializable[T Serializable](x T, b []byte) []byte {
	b, err := x.MarshalAppend(b)
	if err != nil {
		panic(fmt.Errorf("serializing %T: %w", x, err))
	}
	return b
}

func DeserializeSerializable[T Serializable](x T, b []byte) []byte {
	n, err := x.Unmarshal(b)
	if err != nil {
		panic(fmt.Errorf("deserializing %T: %w", x, err))
	}
	return b[n:]
}

func SerializeSliceSize[T any](x []T, b []byte) []byte {
	return binary.AppendVarint(b, int64(len(x)))
}

func DeserializeSliceSize(b []byte) (int, []byte) {
	l, n := binary.Varint(b)
	return int(l), b[n:]
}

// Serializers and deserializers for basic types. Composite types serializers
// and deserializers are generated by cmd/serde.

func SerializeBool(x bool, b []byte) []byte {
	c := byte(0)
	if x {
		c = 1
	}
	return append(b, c)
}

func DeserializeBool(b []byte) (bool, []byte) {
	return b[0] == 1, b[1:]
}

func SerializeUint64(x uint64, b []byte) []byte {
	return binary.LittleEndian.AppendUint64(b, x)
}

func DeserializeUint64(b []byte) (uint64, []byte) {
	return binary.LittleEndian.Uint64(b[:8]), b[8:]
}

func SerializeUint32(x uint32, b []byte) []byte {
	return binary.LittleEndian.AppendUint32(b, x)
}

func DeserializeUint32(b []byte) (uint32, []byte) {
	return binary.LittleEndian.Uint32(b[:4]), b[4:]
}

func SerializeUint16(x uint16, b []byte) []byte {
	return binary.LittleEndian.AppendUint16(b, x)
}

func DeserializeUint16(b []byte) (uint16, []byte) {
	return binary.LittleEndian.Uint16(b[:2]), b[2:]
}

func SerializeUint8(x uint8, b []byte) []byte {
	return append(b, byte(x))
}

func DeserializeUint8(b []byte) (uint8, []byte) {
	return uint8(b[0]), b[1:]
}

func SerializeInt64(x int64, b []byte) []byte {
	return binary.LittleEndian.AppendUint64(b, uint64(x))
}

func DeserializeInt64(b []byte) (int64, []byte) {
	return int64(binary.LittleEndian.Uint64(b[:8])), b[8:]
}

func SerializeInt32(x int32, b []byte) []byte {
	return binary.LittleEndian.AppendUint32(b, uint32(x))
}

func DeserializeInt32(b []byte) (int32, []byte) {
	return int32(binary.LittleEndian.Uint32(b[:4])), b[4:]
}

func SerializeInt16(x int16, b []byte) []byte {
	return binary.LittleEndian.AppendUint16(b, uint16(x))
}

func DeserializeInt16(b []byte) (int16, []byte) {
	return int16(binary.LittleEndian.Uint16(b[:2])), b[2:]
}

func SerializeInt8(x int8, b []byte) []byte {
	return append(b, byte(x))
}

func DeserializeInt8(b []byte) (int8, []byte) {
	return int8(b[0]), b[1:]
}

func SerializeFloat32(x float32, b []byte) []byte {
	return SerializeUint32(math.Float32bits(x), b)
}

func DeserializeFloat32(b []byte) (float32, []byte) {
	u, b := DeserializeUint32(b)
	return math.Float32frombits(u), b
}

func SerializeFloat64(x float64, b []byte) []byte {
	return SerializeUint64(math.Float64bits(x), b)
}

func DeserializeFloat64(b []byte) (float64, []byte) {
	u, b := DeserializeUint64(b)
	return math.Float64frombits(u), b
}

func SerializeComplex64(x complex64, b []byte) []byte {
	b = SerializeFloat32(real(x), b)
	b = SerializeFloat32(imag(x), b)
	return b
}

func DeserializeComplex64(b []byte) (complex64, []byte) {
	r, b := DeserializeFloat32(b)
	i, b := DeserializeFloat32(b)
	return complex(r, i), b
}

func SerializeComplex128(x complex128, b []byte) []byte {
	b = SerializeFloat64(real(x), b)
	b = SerializeFloat64(imag(x), b)
	return b
}
func DeserializeComplex128(b []byte) (complex128, []byte) {
	r, b := DeserializeFloat64(b)
	i, b := DeserializeFloat64(b)
	return complex(r, i), b
}

func SerializeString(s string, b []byte) []byte {
	b = binary.AppendVarint(b, int64(len(s)))
	return append(b, s...)
}

func DeserializeString(b []byte) (string, []byte) {
	l, n := binary.Varint(b)
	b = b[n:]
	return string(b[:l]), b[l:]
}

// TODO: UnsafePointer
