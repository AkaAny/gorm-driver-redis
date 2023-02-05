package gorm_driver_redis

import (
	"fmt"
	"gorm.io/gorm"
	"testing"
)

type TestModel struct {
	RedisEntity
	ID   uint `gorm:"primaryKey"`
	Name string
}

func InitDB() *gorm.DB {
	var dialector = Open("tcp://localhost:6379?userName=&password=&db=0")
	db, err := gorm.Open(dialector)
	if err != nil {
		panic(err)
	}
	return db
}

func TestSaveCreate(t *testing.T) {
	var db = InitDB()
	var dst = &TestModel{
		ID:   0,
		Name: "name1",
	}
	if err := db.Save(dst).Error; err != nil {
		panic(err)
	}
}

func TestSaveCreateWithExpire(t *testing.T) {
	var db = InitDB()
	var dst = &TestModel{
		RedisEntity: RedisEntity{ExpireInSecond: 60},
		ID:          0,
		Name:        "name2",
	}
	if err := db.Save(dst).Error; err != nil {
		panic(err)
	}
}

func TestSaveUpdate(t *testing.T) {
	var db = InitDB()
	var dst = &TestModel{
		ID:   1, //if any of primary field is zero value, gorm will call create
		Name: "name1",
	}
	if err := db.Save(dst).Error; err != nil {
		panic(err)
	}
}

func TestDelete(t *testing.T) {
	var db = InitDB()
	if err := db.Where(TestModel{ID: 1}).Delete(&TestModel{}).Error; err != nil {
		panic(err)
	}
}

func TestFind(t *testing.T) {
	var db = InitDB()
	//fast path: one primary key only
	var result = make([]*TestModel, 0)
	db.Where(TestModel{
		ID: 1,
	}).Find(&result)
	fmt.Println(result)
}

func TestFindWithFullTableScan(t *testing.T) {
	var db = InitDB()
	var result = make([]*TestModel, 0)
	db.Where(TestModel{
		Name: "name1",
	}).Find(&result)
	fmt.Println(result)
}

func TestFirstWithFullTableScan(t *testing.T) {
	var db = InitDB()
	//fast path: one primary key only
	var result = new(TestModel)
	db.Where(TestModel{
		Name: "name1",
	}).First(result)
	fmt.Println(result)
}

func TestFindWithLikeClause(t *testing.T) {
	var db = InitDB()
	var result = make([]*TestModel, 0)
	db.Where("name LIKE ?", "name%").Find(&result)
	fmt.Println(result)
}
