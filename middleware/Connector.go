package middleware

import (
	"github.com/jinzhu/gorm"
	"github.com/nbvghost/roller/entity"
	"github.com/nbvghost/roller/iface"
	"github.com/nbvghost/roller/mold"
)

var Connector = &connector{
	MySQL:  &MySQLConnector{models: make([]iface.ISQLModels, 0)},
	Redis:  &RedisConnector{},
	isInit: false,
}

type connector struct {
	MySQL  *MySQLConnector
	Redis  *RedisConnector
	isInit bool
}

func (c *connector) Init(logicIndex int) {
	if c.isInit {
		panic("connector is init")
	}

	c.MySQL.createMySQL(logicIndex)
	item := &mold.Item{}
	if c.MySQL.GetMySQL().HasTable(item) == false {
		c.MySQL.GetMySQL().Set("gorm:table_options", "AUTO_INCREMENT=10000").CreateTable(item)
	}
	kv := &mold.KV{}
	if c.MySQL.GetMySQL().HasTable(kv) == false {
		c.MySQL.GetMySQL().Set("gorm:table_options", "AUTO_INCREMENT=10000").CreateTable(kv)
	}
	mm := &mold.MassMail{}
	if c.MySQL.GetMySQL().HasTable(mm) == false {
		c.MySQL.GetMySQL().Set("gorm:table_options", "AUTO_INCREMENT=10000").CreateTable(mm)
	}

	Notice := entity.Notice             //&mold.KV{K: "Notice", V: ""}
	ForceOffline := entity.ForceOffline //&mold.KV{K: "ForceOffline", V: ""}
	LoginLimit := entity.LoginLimit     //&mold.KV{K: "LoginLimit", V: "false"}

	err := c.MySQL.GetMySQL().Table((&mold.KV{}).TableName()).Where("K=?", Notice.K).First(&mold.KV{}).Error
	if gorm.IsRecordNotFoundError(err) {
		c.MySQL.GetMySQL().Save(&mold.KV{K: Notice.K, V: ""})
	}
	err = c.MySQL.GetMySQL().Table((&mold.KV{}).TableName()).Where("K=?", ForceOffline.K).First(&mold.KV{}).Error
	if gorm.IsRecordNotFoundError(err) {
		c.MySQL.GetMySQL().Save(&mold.KV{K: ForceOffline.K, V: ""})
	}
	err = c.MySQL.GetMySQL().Table((&mold.KV{}).TableName()).Where("K=?", LoginLimit.K).First(&mold.KV{}).Error
	if gorm.IsRecordNotFoundError(err) {
		c.MySQL.GetMySQL().Save(&mold.KV{K: LoginLimit.K, V: "false"})
	}

	for index := range c.MySQL.models {
		if c.MySQL.GetMySQL().HasTable(c.MySQL.models[index]) == false {
			c.MySQL.GetMySQL().Set("gorm:table_options", "AUTO_INCREMENT=2000").CreateTable(c.MySQL.models[index])
		}
		c.MySQL.GetMySQL().AutoMigrate(c.MySQL.models[index])
	}
	c.Redis.createRedis()
	c.isInit = true
}
