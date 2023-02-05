package gorm_driver_redis

import (
	"context"
	"fmt"
	"gorm.io/gorm"
)

func (rd *RedisDialector) UpdateFn(tx *gorm.DB) {
	fmt.Println(tx)
	var selectedUpdateColumnNames = tx.Statement.Selects
	fmt.Println(selectedUpdateColumnNames)
	//every primary key has value when update
	var sch = tx.Statement.Schema
	var hmap = map[string]interface{}{}
	for _, fieldObj := range sch.Fields {
		v, _ := fieldObj.ValueOf(context.Background(), tx.Statement.ReflectValue)
		hmap[fieldObj.DBName] = v
	}
	//build key
	var redisKey = GetDataKey(sch, tx.Statement.Dest) //TODO: support batch update
	var cmdHandle = rd.RedisClient.HSet(context.Background(), redisKey, hmap)
	if err := cmdHandle.Err(); err != nil {
		_ = tx.AddError(fmt.Errorf("failed to hset with err:%w", err))
		return
	}
}
