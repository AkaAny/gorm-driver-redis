package gorm_driver_redis

import (
	"context"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/migrator"
	"gorm.io/gorm/schema"
	"sync"
)

type RedisMigrator struct {
	migrator.Migrator
	Dialector  *RedisDialector
	schemaMap  map[string]*schema.Schema
	cacheStore *sync.Map
}

func (m *RedisMigrator) getAsRedisDialector() *RedisDialector {
	//return m.Dialector.(*RedisDialector)
	return m.Dialector
}

func (m *RedisMigrator) AutoMigrate(dsts ...interface{}) error {
	for _, dst := range dsts {
		sch, err := m.checkCanBeRegisteredInRedis(dst)
		if err != nil {
			return err
		}
		m.schemaMap[sch.Name] = sch
	}
	return nil
}

func (m *RedisMigrator) checkCanBeRegisteredInRedis(dst interface{}) (*schema.Schema, error) {
	//这个是封装的hmap，主要是检查一下有没有非Valuer类型
	sch, err := schema.Parse(dst, m.cacheStore, schema.NamingStrategy{})
	if err != nil {
		return nil, err
	}

	return sch, nil
}

func (m *RedisMigrator) CurrentDatabase() string {
	var rd = m.getAsRedisDialector()
	return fmt.Sprintf("%d", rd.DB)
}

// RunWithValue run migration with statement value
func (m *RedisMigrator) RunWithValue(value interface{}, fc func(*gorm.Statement) error) error {
	stmt := &gorm.Statement{}
	//if m.DB.Statement != nil {
	//	stmt.Table = m.DB.Statement.Table
	//	stmt.TableExpr = m.DB.Statement.TableExpr
	//}
	if table, ok := value.(string); ok {
		stmt.Table = table
	} else if err := stmt.ParseWithSpecialTableName(value, stmt.Table); err != nil {
		return err
	}

	return fc(stmt)
}

func (m *RedisMigrator) FullDataTypeOf(field *schema.Field) clause.Expr {
	return m.Migrator.FullDataTypeOf(field)
}

func (m *RedisMigrator) GetTypeAliases(databaseTypeName string) []string {
	//TODO implement me
	panic("implement me")
}

func (m *RedisMigrator) CreateTable(dst ...interface{}) error {
	//因为gorm默认的migrator会直接尝试执行sql语句，但是对于redis而言每个hmap的结构实际上没什么关系，所以不如直接拦截掉
	fmt.Println("[redis migrator] create table")
	return nil
}

func (m *RedisMigrator) DropTable(dst ...interface{}) error {
	fmt.Println("[redis migrator] drop table")
	return nil
}

func (m *RedisMigrator) HasTable(dst interface{}) bool {
	_, err := m.checkCanBeRegisteredInRedis(dst)
	return err == nil
}

func (m *RedisMigrator) RenameTable(oldName, newName interface{}) error {
	var rd = m.getAsRedisDialector()
	var pattern = fmt.Sprintf("%s_", oldName)
	var cmdHandle = rd.RedisClient.Keys(context.Background(), pattern)
	if err := cmdHandle.Err(); err != nil {
		return fmt.Errorf("redis get cmd handle err:%w", err)
	}
	var keys = make([]string, 0)
	if err := cmdHandle.ScanSlice(&keys); err != nil {
		return fmt.Errorf("redis scan result slice err:%w", err)
	}
	//var newKeys = make([]string, 0)
	//for _, key := range keys {
	//	var newKey = pattern + key[len(pattern):]
	//	//get value
	//	m.Dialector.RedisClient.HGetAll(context.Background(), key)
	//}
	fmt.Println("[redis migrator] rename table")
	return nil
}

func (m *RedisMigrator) GetTables() (tableList []string, err error) {
	for _, sch := range m.schemaMap {
		tableList = append(tableList, sch.Table)
	}
	return tableList, nil
}

func (m *RedisMigrator) AddColumn(dst interface{}, field string) error {
	//TODO implement me
	panic("implement me")
}

func (m *RedisMigrator) DropColumn(dst interface{}, field string) error {
	//TODO implement me
	panic("implement me")
}

func (m *RedisMigrator) AlterColumn(dst interface{}, field string) error {
	//TODO implement me
	panic("implement me")
}

func (m *RedisMigrator) MigrateColumn(dst interface{}, field *schema.Field, columnType gorm.ColumnType) error {
	return nil
}

func (m *RedisMigrator) HasColumn(dst interface{}, field string) bool {
	var hasField = false
	err := m.RunWithValue(dst, func(stmt *gorm.Statement) error {
		var fieldObj = stmt.Schema.LookUpField(field)
		hasField = fieldObj != nil
		return nil
	})
	if err != nil {
		panic(err)
	}
	return hasField
}

func (m *RedisMigrator) RenameColumn(dst interface{}, oldName, field string) error {
	//TODO implement me
	panic("implement me")
}

func (m *RedisMigrator) ColumnTypes(dst interface{}) ([]gorm.ColumnType, error) {
	//TODO implement me
	panic("implement me")
}

func (m *RedisMigrator) CreateView(name string, option gorm.ViewOption) error {
	//TODO implement me
	panic("implement me")
}

func (m *RedisMigrator) DropView(name string) error {
	//TODO implement me
	panic("implement me")
}

func (m *RedisMigrator) CreateConstraint(dst interface{}, name string) error {
	//TODO implement me
	panic("implement me")
}

func (m *RedisMigrator) DropConstraint(dst interface{}, name string) error {
	//TODO implement me
	panic("implement me")
}

func (m *RedisMigrator) HasConstraint(dst interface{}, name string) bool {
	//TODO implement me
	panic("implement me")
}

func (m *RedisMigrator) CreateIndex(dst interface{}, name string) error {
	return nil
}

func (m *RedisMigrator) DropIndex(dst interface{}, name string) error {
	//TODO implement me
	panic("implement me")
}

func (m *RedisMigrator) HasIndex(dst interface{}, name string) bool {
	var exist = false
	err := m.RunWithValue(dst, func(stmt *gorm.Statement) error {
		var index = stmt.Schema.LookIndex(name)
		exist = index != nil
		return nil
	})
	if err != nil {
		panic(err)
	}
	return exist
}

func (m *RedisMigrator) RenameIndex(dst interface{}, oldName, newName string) error {
	return nil
}

func (m *RedisMigrator) GetIndexes(dst interface{}) ([]gorm.Index, error) {
	return nil, errors.New("unsupported")
}
