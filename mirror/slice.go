package mirror

import (
	"reflect"
	"unsafe"
)

//go:linkname flagROUsingLinkName reflect.flagRO
var flagROUsingLinkName flag

func SliceValueAddr(v reflect.Value) reflect.Value {
	//cancel check
	//if v.flag&flagAddr == 0 {
	//	panic("reflect.Value.Addr of unaddressable value")
	//}

	// Preserve flagRO instead of using v.flag.ro() so that
	// v.Addr().Elem() is equivalent to v (#32772)
	var mirrorOldValue = *(*Value)(unsafe.Pointer(&v))
	fl := mirrorOldValue.flag & flagRO
	//equivalent with reflect.PtrTo but for *rtype instance instead of reflect.Type interface
	//var modelTypeInterface interface{}= unsafe.Pointer(mirrorOldValue.typ)
	var modelTypeInterface = reflect.TypeOf("") //model, now eface.typ is *rtype(*runtime._type) and ptr is an instance
	//replace its ptr to value.typ to avoid copy entire *rtype out of reflect package
	//it is a good method to make a mirror of complicated unexported types
	var modelTypeInterfaceValue = reflect.ValueOf(modelTypeInterface)

	var typInterfaceValue = SetValuePtrAsRaw(modelTypeInterfaceValue, mirrorOldValue.typ)
	var typInterface = typInterfaceValue.Interface().(reflect.Type)
	var ptrToTypeInterface = reflect.PtrTo(typInterface)
	var ptrToTypeInterfaceValue = reflect.ValueOf(ptrToTypeInterface)
	var mirrorAddrValue = Value{
		typ:  GetValuePtrAsRaw(ptrToTypeInterfaceValue),
		ptr:  mirrorOldValue.ptr,
		flag: fl | flag(reflect.Pointer),
	}
	var asReflectValue = *(*reflect.Value)(unsafe.Pointer(&mirrorAddrValue))
	//return Value{v.typ.ptrTo(), v.ptr, fl | flag(reflect.Pointer)}
	return asReflectValue
}

func SetValuePtrAsRaw(v reflect.Value, newPtr uintptr) reflect.Value {
	var mirrorOldValue = *(*Value)(unsafe.Pointer(&v))
	mirrorOldValue.ptr = unsafe.Pointer(newPtr)
	var asReflectValue = *(*reflect.Value)(unsafe.Pointer(&mirrorOldValue))
	return asReflectValue
}

func GetValuePtrAsRaw(v reflect.Value) uintptr {
	var mirrorOldValue = *(*Value)(unsafe.Pointer(&v))
	return uintptr(mirrorOldValue.ptr)
}
