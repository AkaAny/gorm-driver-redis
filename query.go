package gorm_driver_redis

import (
	"context"
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
	"reflect"
	"strings"
)

func GetWhereClause(clauses map[string]clause.Clause) (clause.Where, bool) {
	whereClause, ok := clauses["WHERE"].Expression.(clause.Where)
	if !ok {
		return clause.Where{}, false
	}
	return whereClause, true
}

func (rd *RedisDialector) handleFullTableScan(tx *gorm.DB, whereClause clause.Where,
	sch *schema.Schema, reflectValue reflect.Value) error {
	var keyPattern = fmt.Sprintf(`%s\.\[*\]\.<Data>`, sch.Table)
	var cmdHandle = rd.RedisClient.Keys(context.Background(), keyPattern)
	var keys = make([]string, 0)
	if err := cmdHandle.Err(); err != nil {
		return fmt.Errorf("failed to get all related key with err:%w", err)
	}
	if err := cmdHandle.ScanSlice(&keys); err != nil {
		return fmt.Errorf("failed to unmarshal keys with err:%w", err)
	}
	var redisRows = NewRedisRows(sch)
	for _, key := range keys {
		var cmdHandle = rd.RedisClient.HGetAll(context.Background(), key)
		if err := cmdHandle.Err(); err != nil {
			return fmt.Errorf("failed to call hgetall from redis with err:%w", err)
		}
		rowMap, err := redisClientScanUsingMakeStruct(sch, cmdHandle)
		if err != nil {
			return fmt.Errorf("failed to get and scan key with err:%w", err)
		}
		if predicateFn(whereClause, sch, rowMap) {
			redisRows.AppendRow(rowMap)
		}
	}
	gorm.Scan(redisRows, tx, 0)
	return nil
}

func predicateFn(whereClause clause.Where, sch *schema.Schema, rowMap map[string]interface{}) bool {
	for _, subClause := range whereClause.Exprs {
		switch subClause.(type) {
		case clause.Eq:
			var eqClause = subClause.(clause.Eq)
			var column = eqClause.Column.(clause.Column)
			var fieldObj = sch.LookUpField(column.Name)
			v, _ := rowMap[fieldObj.DBName]
			if v != eqClause.Value {
				return false
			}
		case clause.Like:
			var likeClause = subClause.(clause.Like)
			var column = likeClause.Column.(clause.Column)
			var fieldObj = sch.LookUpField(column.Name)
			v, _ := rowMap[fieldObj.DBName]
			var pattern = fmt.Sprintf("%v", likeClause.Value) //%s% s% %s
			return checkLikePatternMatch(pattern, v)
		default:
			panic("unsupported clause")
		}
	}
	return true
}

func checkLikePatternMatch(pattern string, v interface{}) bool {
	//pattern is str, so we convert all of them to str before compare
	var vStr = fmt.Sprintf("%v", v)
	if pattern == vStr { //fast path:exact match first
		return true
	}
	hasPrefix, hasSuffix := strings.HasPrefix(pattern, "%"), strings.HasSuffix(pattern, "%")
	var sub = strings.Trim(pattern, "%")
	if hasPrefix && hasSuffix { //%s%
		return strings.Contains(vStr, sub)
	}
	if hasPrefix && !hasSuffix { //%s
		return strings.HasSuffix(vStr, sub)
	}
	if !hasPrefix && hasSuffix { //s%
		return strings.HasPrefix(vStr, sub)
	}
	return false
}
