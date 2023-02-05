package gorm_driver_redis

import (
	"fmt"
	"gorm.io/gorm/schema"
	"reflect"
)

type RedisEntity struct {
	ExpireInSecond int64
}

type GetCacheKeyIDAble interface {
	GetCacheKeyID() string
}

func GetDataKey(sch *schema.Schema, dest interface{}) string {
	getCacheKeyIDAble, ok := dest.(GetCacheKeyIDAble)
	if ok {
		var keyID = getCacheKeyIDAble.GetCacheKeyID()
		return GetDataKeyWithKeyID(sch.Table, keyID)
	}
	var primaryKeyField = sch.PrioritizedPrimaryField
	var reflectValue = reflect.ValueOf(dest)
	reflectValue = reflect.Indirect(reflectValue)
	var primaryKeyValue = reflectValue.FieldByName(primaryKeyField.Name).Interface()
	return GetDataKeyWithKeyID(sch.Table, primaryKeyValue)
}

func GetDataKeyWithKeyID(table string, keyID interface{}) string {
	return fmt.Sprintf("%s.[%v].<Data>",
		table, keyID)
}
