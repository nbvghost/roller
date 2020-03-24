package iface

import "time"

type ISQLModels interface {
	TableName() string
	GetID() uint64
}

type SQLModels struct {
	ID        uint64    `gorm:"column:ID;primary_key;unique" json:",omitempty"`           //条目TID
	CreatedAt time.Time `gorm:"column:CreatedAt;index:CreatedAt_Index" json:",omitempty"` //登陆日期
	UpdatedAt time.Time `gorm:"column:UpdatedAt" json:",omitempty"`                       //修改日期
	//DeletedAt *time.Time `gorm:"column:DeletedAt" json:",omitempty"` //删除日期
	//Delete    int        `gorm:"column:Delete"`                //0=无，1=删除，
}

func (b *SQLModels) GetID() uint64 {
	return b.ID
}
