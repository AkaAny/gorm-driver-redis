package gorm_driver_redis

import (
	"fmt"
	"gorm-driver-redis/mirror"
	"gorm.io/gorm"
	"reflect"
)

func (rd *RedisDialector) DeleteFn(tx *gorm.DB) {
	fmt.Println(tx)
	var sch = tx.Statement.Schema
	//build dest
	var sliceType = reflect.SliceOf(sch.ModelType)
	var reflectValue = reflect.MakeSlice(sliceType, 0, 20)    //make(type,0,20)
	var reflectPtrValue = mirror.SliceValueAddr(reflectValue) //&[]type
	tx.Statement.ReflectValue = reflectPtrValue
	//direct full table scan
	var findMethod = reflect.ValueOf(tx.Find)
	findMethod.Call([]reflect.Value{reflect.ValueOf(tx), reflectValue})
	var destLen = reflectValue.Len()
	fmt.Println(destLen)
	//now tx dest is our target
}
