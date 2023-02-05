package gorm_driver_redis

import (
	"context"
	"fmt"
	"gorm-driver-redis/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"reflect"
	"time"
)

func (rd *RedisDialector) CreateFn(tx *gorm.DB) {
	var toSaveValue = tx.Statement.Dest //dest is to save obj
	var reflectValue = tx.Statement.ReflectValue
	fmt.Println("to save value:", toSaveValue)
	var hmap = map[string]interface{}{}
	var sch = tx.Statement.Schema
	var fields = sch.Fields
	var redisEntityPart = getRedisEntityPart(sch)(reflectValue)
	fmt.Println("redis entity part:", redisEntityPart)
	for _, field := range fields {
		//var fieldReflectValue = tx.Statement.ReflectValue.FieldByName(field.Name)
		primitiveValue, isZero := field.ValueOf(context.Background(), reflectValue)
		//handle auto incr value
		if isZero && field.AutoIncrement {
			//get key like JPA
			var lastInsertIDKey = fmt.Sprintf("%s.%s.<LastInsertID>",
				tx.Statement.Schema.Table, field.Name)
			var cmdHandle = rd.RedisClient.IncrBy(context.Background(), lastInsertIDKey, field.AutoIncrementIncrement)
			if err := cmdHandle.Err(); err != nil {
				_ = tx.AddError(fmt.Errorf("failed to incr last insert id for field:%s with err:%w",
					field.Name, err))
				return
			}
			var incredVal = cmdHandle.Val()
			if redisEntityPart.ExpireInSecond != 0 {
				var cmdHandle = rd.RedisClient.Expire(context.Background(), lastInsertIDKey,
					time.Duration(redisEntityPart.ExpireInSecond)*time.Second)
				if err := cmdHandle.Err(); err != nil {
					_ = tx.AddError(fmt.Errorf("failed to set last insert id expire for field:%s with err:%w",
						field.Name, err))
					return
				}
			}

			if err := field.Set(context.Background(), reflectValue, incredVal); err != nil {
				_ = tx.AddError(fmt.Errorf("failed to set incred val to field:%s with err:%w",
					field.Name, err)) //如果在这之后出错，破坏了redis事务的完整性，但是自增可以不连续
				return
			}
			primitiveValue = incredVal
		}
		hmap[field.DBName] = primitiveValue
	}
	//build search set key: value is json like {pk1_dbname:pk1,pk2_dbname:pk2,pk3_dbname:pk3}
	//and key is full -set of primary keys
	//var primaryKeyMap = map[string]interface{}{}
	//var primaryKeyFields = tx.Statement.Schema.PrimaryFields
	//for _, field := range primaryKeyFields {
	//	//dst has 3 primary keys: a b c, there is several possibilities
	//	//a, ab, ac, abc, bc
	//	//if got any of above existed, the record is existed
	//}

	//build key
	var redisKey = GetDataKey(sch, toSaveValue)
	//set to redis
	var cmdHandle = rd.RedisClient.HSet(context.Background(), redisKey, hmap)
	if err := cmdHandle.Err(); err != nil {
		_ = tx.AddError(fmt.Errorf("failed to hset to redis with err:%w", err))
	}
	if redisEntityPart.ExpireInSecond != 0 {
		var cmdHandle = rd.RedisClient.Expire(context.Background(), redisKey,
			time.Duration(redisEntityPart.ExpireInSecond)*time.Second)
		if err := cmdHandle.Err(); err != nil {
			_ = tx.AddError(fmt.Errorf("failed to set key expire with err:%w", err))
			return
		}
	}
}

func checkIfRedisEntity(fieldObj *schema.Field) bool {
	return fieldObj.FieldType == reflect.TypeOf(RedisEntity{})
}

type StructFieldNameAndType struct {
	Name string
	Type reflect.Type
}

func (x StructFieldNameAndType) Equals(other interface{}) bool {
	otherAs, ok := other.(StructFieldNameAndType)
	if !ok {
		return false
	}
	return x.Name == otherAs.Name && x.Type == otherAs.Type
}

func getRedisEntityPart(sch *schema.Schema) func(reflectValue reflect.Value) RedisEntity {
	var redisEntityType = reflect.TypeOf(RedisEntity{})
	var redisEntityFieldNameAndTypes = make([]StructFieldNameAndType, 0)
	for i := 0; i < redisEntityType.NumField(); i++ {
		var field = redisEntityType.Field(i)
		redisEntityFieldNameAndTypes = append(redisEntityFieldNameAndTypes, StructFieldNameAndType{
			Name: field.Name,
			Type: field.Type,
		})
	}
	return func(reflectValue reflect.Value) RedisEntity {
		var startIndex = -1
		for i := 0; i < len(sch.Fields); i++ {
			if i+len(redisEntityFieldNameAndTypes) > len(sch.Fields)-1 {
				return RedisEntity{}
			}
			var partFields = sch.Fields[i : i+len(redisEntityFieldNameAndTypes)]
			var partFieldNameAndTypes = utils.Map[*schema.Field, StructFieldNameAndType](partFields,
				func(fieldObj *schema.Field) StructFieldNameAndType {
					return StructFieldNameAndType{
						Name: fieldObj.Name,
						Type: fieldObj.FieldType,
					}
				})
			if utils.SliceEqual[StructFieldNameAndType](redisEntityFieldNameAndTypes, partFieldNameAndTypes) {
				startIndex = i
				break
			}
		}
		if startIndex == -1 {
			return RedisEntity{}
		}
		var redisEntityValue = reflect.New(redisEntityType).Elem()
		for i := startIndex; i < startIndex+redisEntityType.NumField(); i++ {
			var fieldObj = sch.Fields[i]
			v, _ := fieldObj.ValueOf(context.Background(), reflectValue)
			var vValue = reflect.ValueOf(v)
			redisEntityValue.FieldByName(fieldObj.Name).Set(vValue)
			//if err := fieldObj.Set(context.Background(), redisEntityValue, v); err != nil {
			//	panic(fmt.Errorf("failed to copy field:%s to new redis entity with err:%w",
			//		fieldObj.Name, err))
			//}
		}
		var result = redisEntityValue.Interface().(RedisEntity)
		return result
	}

}
