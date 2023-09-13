package coroutine

import (
	"fmt"
	"reflect"
	"testing"
)

type EasyStruct struct {
	A int
	B string
}

func TestReflect(t *testing.T) {
	withBlankTypeMap(func() {
		intv := int(100)
		intp := &intv
		intpp := &intp

		cases := []interface{}{
			"foo",
			true,
			int(42),
			int64(11),
			int32(10),
			int16(9),
			int8(8),
			uint(42),
			uint64(11),
			uint32(10),
			uint16(9),
			uint8(8),
			intp,
			intpp,
			[2]int{1, 2},
			[]int{1, 2, 3},
			map[string]int{"one": 1, "two": 2},
			EasyStruct{
				A: 52,
				B: "test",
			},
			[]interface{}{},
		}

		for _, x := range cases {
			tm.Add(reflect.TypeOf(x))
		}

		for i, x := range cases {
			x := x
			typ := reflect.TypeOf(x)
			t.Run(fmt.Sprintf("%d-%s", i, typ), func(t *testing.T) {
				var b []byte

				b = Serialize(x, b)
				out, b := Deserialize(b)

				assertEqual(t, x, out)

				if len(b) > 0 {
					t.Fatalf("leftover bytes: %d", len(b))
				}
			})
		}
	})
}

func TestReflectSharing(t *testing.T) {
	testReflect(t, "slice backing array", func(t *testing.T) {
		data := make([]int, 10)
		for i := range data {
			data[i] = i
		}

		type X struct {
			s1 []int
			s2 []int
			s3 []int
		}

		orig := X{
			s1: data[0:3],
			s2: data[2:8],
			s3: data[7:10],
		}
		assertEqual(t, []int{0, 1, 2}, orig.s1)
		assertEqual(t, []int{2, 3, 4, 5, 6, 7}, orig.s2)
		assertEqual(t, []int{7, 8, 9}, orig.s3)

		assertEqual(t, 10, cap(orig.s1))
		assertEqual(t, 3, len(orig.s1))
		assertEqual(t, 8, cap(orig.s2))
		assertEqual(t, 6, len(orig.s2))
		assertEqual(t, 3, cap(orig.s3))
		assertEqual(t, 3, len(orig.s3))

		RegisterType[X]()

		out := assertRoundTrip(t, orig)

		// verify that the initial arrays were shared
		orig.s1[2] = 42
		assertEqual(t, 42, orig.s2[0])
		orig.s2[5] = 11
		assertEqual(t, 11, orig.s3[0])

		// verify the result's underlying array is shared
		out.s1[2] = 42
		assertEqual(t, 42, out.s2[0])
		out.s2[5] = 11
		assertEqual(t, 11, out.s3[0])
	})

	testReflect(t, "struct fields extra pointers", func(t *testing.T) {
		type A struct {
			X, Y int
		}

		type B struct {
			P *int
		}

		type X struct {
			B *B
			A *A
			// putting A after B to make sure A gets serialized
			// first because of dependencies, not just because it's
			// earlier than B in the fields list.
		}

		x := X{
			A: new(A),
			B: new(B),
		}
		x.B.P = &x.A.Y

		// verify the original pointer is correct
		x.A.Y = 42
		assertEqual(t, 42, *x.B.P)

		RegisterType[X]()
		out := assertRoundTrip(t, x)

		// verify the resulting pointer is correct
		out.A.Y = 11
		assertEqual(t, 11, *out.B.P)
	})
}

func assertEqual(t *testing.T, expected, actual any) {
	t.Helper()
	if !reflect.DeepEqual(expected, actual) {
		t.Error("unexpected context")
		t.Logf("   got: %#v", actual)
		t.Logf("expect: %#v", expected)
	}
}

func assertRoundTrip[T any](t *testing.T, orig T) T {
	t.Helper()

	var b []byte
	b = Serialize(orig, b)
	out, b := Deserialize(b)

	assertEqual(t, orig, out)

	if len(b) > 0 {
		t.Fatalf("leftover bytes: %d", len(b))
	}

	return out.(T)
}

func withBlankTypeMap(f func()) {
	oldtm := tm
	tm = newTypeMap()
	defer func() { tm = oldtm }()

	f()
}

func testReflect(t *testing.T, name string, f func(t *testing.T)) {
	t.Helper()
	t.Run(name, func(t *testing.T) {
		withBlankTypeMap(func() {
			f(t)
		})
	})
}
