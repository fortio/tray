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
	if diff != 6 {
		t.Errorf("Alignment changed, expected 6 got %d (%p %p)", diff, p0, p1)
	}
}
