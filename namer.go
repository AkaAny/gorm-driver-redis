package gorm_driver_redis

import (
	"gorm.io/gorm/schema"
)

type RedisNamer struct {
}

func (rn *RedisNamer) TableName(table string) string {
	//TODO implement me
	panic("implement me")
}

func (rn *RedisNamer) SchemaName(table string) string {
	//TODO implement me
	panic("implement me")
}

func (rn *RedisNamer) ColumnName(table, column string) string {
	//TODO implement me
	panic("implement me")
}

func (rn *RedisNamer) JoinTableName(joinTable string) string {
	//TODO implement me
	panic("implement me")
}

func (rn *RedisNamer) RelationshipFKName(relationship schema.Relationship) string {
	//TODO implement me
	panic("implement me")
}

func (rn *RedisNamer) CheckerName(table, column string) string {
	//TODO implement me
	panic("implement me")
}

func (rn *RedisNamer) IndexName(table, column string) string {
	//TODO implement me
	panic("implement me")
}
