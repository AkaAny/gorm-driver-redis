package gorm_driver_redis

import (
	"database/sql"
	"fmt"
	"gorm-driver-redis/mirror"
	"gorm.io/gorm/schema"
	"reflect"
	"unsafe"
)

type RedisRows struct {
	sch    *schema.Schema
	rows   []map[string]interface{}
	cursor int
}

func NewRedisRows(sch *schema.Schema) *RedisRows {
	return &RedisRows{
		sch:    sch,
		cursor: 0,
	}
}

func (rr *RedisRows) AppendRow(rowMap map[string]interface{}) {
	if rr.cursor != 0 {
		panic("cannot add row after next is called")
	}
	rr.rows = append(rr.rows, rowMap)
}

func (rr *RedisRows) Columns() ([]string, error) {
	var columns = make([]string, 0)
	for _, fieldObj := range rr.sch.Fields {
		var columnName = fieldObj.DBName
		columns = append(columns, columnName)
	}
	return columns, nil
}

func (rr *RedisRows) ColumnTypes() ([]*sql.ColumnType, error) {
	var columnTypes = make([]*sql.ColumnType, 0)
	for _, fieldObj := range rr.sch.Fields {
		var fieldLength = fieldObj.IndirectFieldType.Len()
		var columnType = &mirror.ColumnType{
			Name:              fieldObj.Name,
			HasNullable:       !fieldObj.NotNull,
			HasLength:         false,
			HasPrecisionScale: fieldObj.Scale != 0,
			Nullable:          !fieldObj.NotNull,
			Length:            int64(fieldLength),
			DatabaseType:      string(fieldObj.DataType),
			Precision:         int64(fieldObj.Precision),
			Scale:             int64(fieldObj.Scale),
			ScanType:          fieldObj.IndirectFieldType,
		}
		var sqlColumnType = (*sql.ColumnType)(unsafe.Pointer(columnType))
		columnTypes = append(columnTypes, sqlColumnType)
	}
	return columnTypes, nil
}

func (rr *RedisRows) Next() bool {
	if rr.cursor == len(rr.rows) {
		return false
	}
	return true
}

func (rr *RedisRows) getCurAndAddCursor() map[string]interface{} {
	var rowMap = rr.rows[rr.cursor]
	rr.cursor++
	return rowMap
}

func (rr *RedisRows) Scan(dests ...interface{}) error {
	var rowMap = rr.getCurAndAddCursor()
	for columnIndex, _ := range dests {
		//var destValue = reflect.ValueOf(dest) //scanIntoStruct us *<indirect field type>
		var k = rr.sch.Fields[columnIndex].DBName
		var v = rowMap[k]
		fmt.Println("scan column:", k, v)
		//var vValue = reflect.ValueOf(v)
		//just take addr and copy as dests[i] is **<indirect field type>
		dests[columnIndex] = &v
		//destElemValue.Set(vValue)
	}
	return nil
}

func getElem(destValue reflect.Value) reflect.Value {
	var result = destValue
	for result.Kind() == reflect.Ptr {
		result = result.Elem()
	}
	return result
}

func (rr *RedisRows) Err() error {
	return nil
}

func (rr *RedisRows) Close() error {
	return nil
}
