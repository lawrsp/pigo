package parser

import (
	"log"
	"testing"
)

type TT int

type YY = log.Logger

type BB = int
type CC = BB

type DD map[string]interface{}

func TestTypeReduce(t *testing.T) {
	log.Printf("begin test Type:")
	var x **int
	x = new(*int)
	assert(t, x != nil, "x != nil")
	assert(t, *x == nil, "*x == nil")

	y := new(TT)
	assert(t, y != nil, "y != nil")
	assert(t, *y == 0, "*y == nil")
	y2 := TT(10)
	assert(t, y2 == 10, "*y == nil")

	z := new(YY)
	assert(t, z != nil, "z != nil")

	m := &YY{}
	assert(t, m != nil, "m != nil")

	b := CC(0)
	assert(t, b == 0, "b == 0")

	d := DD{}
	assert(t, d != nil, "d != nil")
	d2 := DD(map[string]interface{}{})
	assert(t, d2 != nil, "d2 != nil")
}

func TestTypeEqual(t *testing.T) {
	log.Printf("begin test Type equal:")

	a := MapType(&BasicType{"string"}, &BasicType{"string"})
	b := &Field{
		Type: a,
		Tag:  "",
		name: "Test",
	}

	if !TypeEqual(a, b) {
		t.Errorf("%s != %s", a, b)
	}
	if !TypeEqual(b, a) {
		t.Errorf("%s != %s", b, a)
	}

	if !TypeEqual(a, a) {
		t.Errorf("%s != %s", a, a)
	}
	if !TypeEqual(b, b) {
		t.Errorf("%s != %s", b, b)
	}

	a2 := MapType(&BasicType{"string"}, &InterfaceType{})
	b2 := &Field{
		Type: a2,
		Tag:  "",
		name: "Test",
	}
	if !TypeEqual(a2, b2) {
		t.Errorf("%s != %s", a2, b2)
	}
	if !TypeEqual(b2, a2) {
		t.Errorf("%s != %s", b2, a2)
	}
}

func TestTypeAssignable(t *testing.T) {
	log.Printf("begin test Type equal:")

	a := MapType(&BasicType{"string"}, &BasicType{"interface"})
	b := &Field{
		Type: a,
		Tag:  "",
		name: "Test",
	}

	if !TypeAssignable(a, b) {
		t.Errorf("%s != %s", a, b)
	}
	if !TypeAssignable(b, a) {
		t.Errorf("%s != %s", b, a)
	}

	if !TypeAssignable(a, a) {
		t.Errorf("%s != %s", a, a)
	}
	if !TypeAssignable(b, b) {
		t.Errorf("%s != %s", b, b)
	}

	a2 := &BasicType{"interface"}
	b2 := &namedType{Type: &StructType{}, name: "TestStruct"}
	if !TypeAssignable(a2, b2) {
		t.Errorf("%s !!= %s", a2, b2)
	}
	if TypeAssignable(b2, a2) {
		t.Errorf("%s == %s", b2, a2)
	}

	a3 := &InterfaceType{}
	b3 := &namedType{Type: &StructType{}, name: "TestStruct"}
	if !TypeAssignable(a3, b3) {
		t.Errorf("%s !!= %s", a3, b3)
	}
	if TypeAssignable(b3, a3) {
		t.Errorf("%s == %s", b3, a3)
	}

}
