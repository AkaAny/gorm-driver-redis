package mirror

import "unsafe"

type iface struct {
	tab  uintptr //*itab
	data unsafe.Pointer
}

type eface struct {
	_type uintptr //*_type
	data  unsafe.Pointer
}

func efaceOf(ep *any) *eface {
	return (*eface)(unsafe.Pointer(ep))
}
