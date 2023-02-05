package gorm_driver_redis

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"gorm-driver-redis/utils"
	_ "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/migrator"
	"gorm.io/gorm/schema"
	"net/url"
	"reflect"
	"strconv"
	"sync"
)

type RedisDialector struct {
	DSNUrl      *url.URL
	DB          int
	RedisClient *redis.Client
}

func Open(dsn string) gorm.Dialector {
	dsnUrl, err := url.Parse(dsn)
	if err != nil {
		panic(fmt.Errorf("failed to parse dsn with err:%w", err))
	}
	var dbStr = dsnUrl.Query().Get("db")
	db, err := strconv.ParseInt(dbStr, 10, 64)
	if err != nil {
		panic(fmt.Errorf("failed to parse db with err:%w", err))
	}
	client := redis.NewClient(&redis.Options{
		Network:  dsnUrl.Scheme,
		Addr:     dsnUrl.Host,
		Username: dsnUrl.Query().Get("userName"),
		Password: dsnUrl.Query().Get("password"),
		DB:       int(db),
	})
	var dialector = &RedisDialector{
		DSNUrl:      dsnUrl,
		DB:          int(db),
		RedisClient: client,
	}
	return dialector
}
func (rd *RedisDialector) Name() string {
	return "redis"
}

func (rd *RedisDialector) Initialize(db *gorm.DB) error {
	err := db.Callback().Create().Replace("gorm:create", rd.CreateFn)
	if err != nil {
		return fmt.Errorf("failed to replace create hook with err:%w", err)
	}
	err = db.Callback().Update().Replace("gorm:update", rd.UpdateFn)
	if err != nil {
		return fmt.Errorf("failed to replace update hook with err:%w", err)
	}
	err = db.Callback().Query().Replace("gorm:query", rd.QueryFn)
	if err != nil {
		return fmt.Errorf("failed to replace query hook with err:%w", err)
	}
	err = db.Callback().Delete().Replace("gorm:delete", rd.DeleteFn)
	if err != nil {
		return fmt.Errorf("failed to replace update hook with err:%w", err)
	}

	return nil
}

func (rd *RedisDialector) FromHMap(sch *schema.Schema, destValue reflect.Value,
	hmap map[string]interface{}) error {
	for k, v := range hmap {
		var fieldObj = sch.LookUpField(k)
		if fieldObj == nil {
			continue
		}
		if err := fieldObj.Set(context.Background(), destValue, v); err != nil {
			return fmt.Errorf("failed to set hmap key:%s field:%s with err:%w",
				k, fieldObj.Name, err)
		}
	}
	return nil
}

func (rd *RedisDialector) QueryFn(tx *gorm.DB) {
	fmt.Println(tx)
	var reflectValue = tx.Statement.ReflectValue //dest can be struct ptr or slice ptr
	whereClause, hasWhere := GetWhereClause(tx.Statement.Clauses)
	if !hasWhere { //find all
		if err := rd.handleFullTableScan(tx, whereClause, tx.Statement.Schema, reflectValue); err != nil {
			_ = tx.AddError(err)
		}
		return
	}
	whereClause, err := ParseWhereClauses(tx, whereClause)
	if err != nil {
		_ = tx.AddError(err)
		return
	}
	//id first, if entity has only one primary key and it is set in clause, direct use it
	var sch = tx.Statement.Schema
	if len(sch.PrimaryFields) == 1 {
		var primaryKeyField = sch.PrimaryFields[0]
		//TODO: build a new struct with same field but has extra redis tag and its value is dbname
		var filteredClauseExprs = utils.Filter[clause.Expression](whereClause.Exprs, func(item clause.Expression) bool {
			eqClause, ok := item.(clause.Eq)
			if !ok {
				return false
			}
			if clauseColumn, ok := eqClause.Column.(clause.Column); ok {
				return clauseColumn.Name == primaryKeyField.DBName
			}
			return false
		})

		if len(filteredClauseExprs) > 0 { //set in clause
			var primaryKeyClause = filteredClauseExprs[0].(clause.Eq)
			var redisKey = GetDataKeyWithKeyID(sch.Table, primaryKeyClause.Value)
			var cmdHandle = rd.RedisClient.HGetAll(context.Background(), redisKey)
			if err := cmdHandle.Err(); err != nil {
				_ = tx.AddError(fmt.Errorf("failed to call hgetall from redis with err:%w", err))
				return
			}
			//var hmap = make(map[string]interface{})
			var redisRows = NewRedisRows(sch)
			rowMap, err := redisClientScanUsingMakeStruct(sch, cmdHandle)
			if err != nil {
				_ = tx.AddError(err)
				return
			}
			redisRows.AppendRow(rowMap)
			//if err := rd.FromHMap(sch, destItemValue, hmap); err != nil {
			//	_ = tx.AddError(fmt.Errorf("failed to unmarshal from hmap with err:%w", err))
			//}
			//ensure update
			gorm.Scan(redisRows, tx, 0)
			return
		}
	}
	//full table scan
	if err := rd.handleFullTableScan(tx, whereClause, sch, reflectValue); err != nil {
		_ = tx.AddError(err)
		return
	}
	fmt.Println("handle limit")
	//limit
}

func redisClientScanUsingMakeStruct(sch *schema.Schema,
	cmdHandle *redis.StringStringMapCmd) (map[string]interface{}, error) {
	var fields = make([]reflect.StructField, 0)
	for _, fieldObj := range sch.Fields {
		var field = fieldObj.StructField
		var tagStr = fmt.Sprintf(`redis:"%s"`, fieldObj.DBName)
		field.Tag = reflect.StructTag(tagStr)
		fields = append(fields, field)
	}
	var forScanType = reflect.StructOf(fields)
	var forScanValue = reflect.New(forScanType)
	var forScanInterface = forScanValue.Interface()
	if err := cmdHandle.Scan(forScanInterface); err != nil {
		return nil, fmt.Errorf("failed to scan result from handle with err:%w", err)
	}
	//copy to dest value
	var rowMap = map[string]interface{}{}
	for _, fieldObj := range sch.Fields {
		v := forScanValue.Elem().FieldByName(fieldObj.Name).Interface()
		//if err := fieldObj.Set(context.Background(), destValue, v); err != nil {
		//	return fmt.Errorf("failed to set field:%s to dest value with err:%w",
		//		fieldObj.Name, err)
		//}
		rowMap[fieldObj.DBName] = v
	}
	return rowMap, nil
}

func constructSearchSet(sch *schema.Schema, dst interface{}) {
	//所有带index的字段都要单独写一个
	//var searchKeySet = make(map[string]interface{})
	var primaryKeyIndexes = make([]int, len(sch.PrimaryFields))
	for i := 0; i < len(primaryKeyIndexes); i++ {
		primaryKeyIndexes[i] = i
	}
	//var allPossibleSeqs= GenIntSlicePermutationWithDict(primaryKeyIndexes)
	//for _,seq:=range allPossibleSeqs{
	//	for _,index:=range seq{
	//		var field=
	//	}
	//}
	//for _, field := range sch.PrimaryFields {
	//
	//}
}

func (rd *RedisDialector) Migrator(db *gorm.DB) gorm.Migrator {
	return &RedisMigrator{
		Migrator: migrator.Migrator{
			Config: migrator.Config{
				DB:        db,
				Dialector: rd,
			},
		},
		Dialector:  rd,
		cacheStore: &sync.Map{},
	}
}

func (rd *RedisDialector) DataTypeOf(field *schema.Field) string {
	return string(field.DataType)
}

func (rd *RedisDialector) DefaultValueOf(field *schema.Field) clause.Expression {
	return clause.Expr{SQL: "DEFAULT"}
}

func (rd *RedisDialector) BindVarTo(writer clause.Writer, stmt *gorm.Statement, v interface{}) {
	writer.WriteByte('?')
}

func (rd *RedisDialector) QuoteTo(writer clause.Writer, str string) {
	var (
		underQuoted, selfQuoted bool
		continuousBacktick      int8
		shiftDelimiter          int8
	)

	for _, v := range []byte(str) {
		switch v {
		case '`':
			continuousBacktick++
			if continuousBacktick == 2 {
				writer.WriteString("``")
				continuousBacktick = 0
			}
		case '.':
			if continuousBacktick > 0 || !selfQuoted {
				shiftDelimiter = 0
				underQuoted = false
				continuousBacktick = 0
				writer.WriteByte('`')
			}
			writer.WriteByte(v)
			continue
		default:
			if shiftDelimiter-continuousBacktick <= 0 && !underQuoted {
				writer.WriteByte('`')
				underQuoted = true
				if selfQuoted = continuousBacktick > 0; selfQuoted {
					continuousBacktick -= 1
				}
			}

			for ; continuousBacktick > 0; continuousBacktick -= 1 {
				writer.WriteString("``")
			}

			writer.WriteByte(v)
		}
		shiftDelimiter++
	}

	if continuousBacktick > 0 && !selfQuoted {
		writer.WriteString("``")
	}
	writer.WriteByte('`')
}

func (rd *RedisDialector) Explain(sql string, vars ...interface{}) string {
	//TODO implement me
	panic("implement me")
}
