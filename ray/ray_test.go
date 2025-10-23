package ray

import (
	"structs"
	"testing"
	"unsafe"
)

type ATest struct {
	_       structs.HostLayout
	a, b, c uint16
}

func TestAlignment(t *testing.T) {
	at := [2]ATest{}
	p1 := &at[1]
	p0 := &at[0]
	diff := uintptr(unsafe.Pointer(p1)) - uintptr(unsafe.Pointer(p0))
	t.Errorf("Error by design just to show the value %p %p: %d", p0, p1, diff)
}
