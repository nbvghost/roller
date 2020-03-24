package iface

import "github.com/jinzhu/gorm"

type ICache interface {
	ObtainCache(DB *gorm.DB, UserID uint64) error
	Persistence(DB *gorm.DB) error
	//FlushRedis() error
	//ObtainCache(DB *gorm.DB, UserID uint64) error
	//Persistence(DB *gorm.DB) error
}
