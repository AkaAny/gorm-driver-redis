package mirror

import (
	"fmt"
	"reflect"
	"testing"
)

func TestFlagROLinkName(t *testing.T) {
	fmt.Println("flagRO:", flagROUsingLinkName)
}

func TestSliceValueAddr(t *testing.T) {
	var strType = reflect.TypeOf("")
	var strSliceType = reflect.SliceOf(strType)
	var strSliceValue = reflect.MakeSlice(strSliceType, 0, 0)
	var ptrToStrSliceValue = SliceValueAddr(strSliceValue)
	fmt.Println(ptrToStrSliceValue.Type().String())
}
