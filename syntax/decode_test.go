package syntax

import (
	"reflect"
	"testing"
)

func TestDecodeUnsupported(t *testing.T) {
	var y struct {
		Strings []string
	}
	read, err := Unmarshal(z8, &y)
	if err == nil || read != 0 {
		t.Fatalf("Agreed to unmarshal an unsupported type")
	}

	var yi int
	read, err = Unmarshal(z8, yi)
	if err == nil || read != 0 {
		t.Fatalf("Agreed to unmarshal to a non-pointer")
	}

	read, err = Unmarshal(z8, nil)
	if err == nil || read != 0 {
		t.Fatalf("Agreed to unmarshal to a nil pointer")
	}
}

func TestDecodeBasicTypes(t *testing.T) {
	var y8 uint8
	read, err := Unmarshal(z8, &y8)
	if err != nil || y8 != x8 || read != 1 {
		t.Fatalf("uint8 decode failed [%v] [%x]", err, y8)
	}

	var y16 uint16
	read, err = Unmarshal(z16, &y16)
	if err != nil || y16 != x16 || read != 2 {
		t.Fatalf("uint16 decode failed [%v] [%x]", err, y16)
	}

	var y32 uint32
	read, err = Unmarshal(z32, &y32)
	if err != nil || y32 != x32 || read != 4 {
		t.Fatalf("uint32 decode failed [%v] [%x]", err, y32)
	}

	var y64 uint64
	read, err = Unmarshal(z64, &y64)
	if err != nil || y64 != x64 || read != 8 {
		t.Fatalf("uint64 decode failed [%v] [%x]", err, y64)
	}

	read, err = Unmarshal(z8[:0], &y8)
	if err == nil || read != 0 {
		t.Fatalf("Allowed uint8 decode from an empty buffer")
	}

	read, err = Unmarshal(z64[:2], &y64)
	if err == nil || read != 0 {
		t.Fatalf("Allowed uint64 decode from an incomplete buffer")
	}
}

func TestDecodeArray(t *testing.T) {
	var ya [5]uint16
	read, err := Unmarshal(za, &ya)
	if err != nil || !reflect.DeepEqual(ya, xa) || read != 10 {
		t.Fatalf("[5]uint16 decode failed [%v] [%x]", err, ya)
	}
}

func TestDecodeSlice(t *testing.T) {
	var yv20 struct {
		V []byte `tls:"head=1"`
	}
	read, err := Unmarshal(zv20, &yv20)
	if err != nil || !reflect.DeepEqual(yv20, xv20) || read != len(zv20) {
		t.Fatalf("[0x20]uint8 decode failed [%v] [%x]", err, yv20.V)
	}

	var yv200 struct {
		V []byte `tls:"head=2"`
	}
	read, err = Unmarshal(zv200, &yv200)
	if err != nil || !reflect.DeepEqual(yv200, xv200) || read != len(zv200) {
		t.Fatalf("[0x200]uint8 decode failed [%v] [%x]", err, yv200.V)
	}

	var yv20000 struct {
		V []byte `tls:"head=3"`
	}
	read, err = Unmarshal(zv20000, &yv20000)
	if err != nil || !reflect.DeepEqual(yv20000, xv20000) || read != len(zv20000) {
		t.Fatalf("[0x20000]uint8 decode failed [%v] [%x]", err, yv20000.V)
	}

	var yvEhead struct {
		V []byte
	}
	read, err = Unmarshal(zv20, &yvEhead)
	if err == nil || read != 0 {
		t.Fatalf("Allowed a vector decode with no head")
	}

	var yvEmax struct {
		V []byte `tls:"head=1,max=31"`
	}
	read, err = Unmarshal(zv20, &yvEmax)
	if err == nil || read != 0 {
		t.Fatalf("Allowed a vector decode with length exceeding max")
	}

	var yvEmin struct {
		V []byte `tls:"head=1,min=33"`
	}
	read, err = Unmarshal(zv20, &yvEmin)
	if err == nil || read != 0 {
		t.Fatalf("Allowed a vector decode with length below min")
	}

	read, err = Unmarshal(zv200[:0], &yv200)
	if err == nil || read != 0 {
		t.Fatalf("Allowed a vector decode from an empty buffer")
	}

	read, err = Unmarshal(zv200[:1], &yv200)
	if err == nil || read != 0 {
		t.Fatalf("Allowed a vector decode with length too short")
	}

	read, err = Unmarshal(zv200[:2], &yv200)
	if err == nil || read != 0 {
		t.Fatalf("Allowed a vector decode shorter than declared length [%x]", yv200.V)
	}

	read, err = Unmarshal(zv200[:5], &yv200)
	if err == nil || read != 0 {
		t.Fatalf("Allowed a vector decode shorter than declared length [%x]", yv200.V)
	}
}

func TestDecodeStruct(t *testing.T) {
	var ys1 struct {
		A uint16
		B []uint8 `tls:"head=2"`
		C [4]uint32
	}
	read, err := Unmarshal(zs1, &ys1)
	if err != nil || !reflect.DeepEqual(ys1, xs1) || read != len(zs1) {
		t.Fatalf("struct decode failed [%v] [%v]", err, ys1)
	}
}

func TestIgnoreExtraData(t *testing.T) {
	zsExtra := append(zs1, 0)
	var ys1 struct {
		A uint16
		B []uint8 `tls:"head=2"`
		C [4]uint32
	}
	read, err := Unmarshal(zsExtra, &ys1)
	if err != nil || !reflect.DeepEqual(ys1, xs1) || read != len(zs1) {
		t.Fatalf("struct decode failed [%v] [%v]", err, ys1)
	}
}
