package mold

import (
	"github.com/jinzhu/gorm"
	"github.com/nbvghost/glog"
	"github.com/nbvghost/roller/entity/pb"
	"github.com/nbvghost/roller/iface"
)

type ItemIDItemType struct {
	ItemID   ItemID
	ItemType ItemType
	Name     string
}

type ItemID uint64
type ItemType uint64
type ItemOperationType string

/*type BaseModel struct {
	ID        uint64    `gorm:"column:ID;primary_key;unique" json:",omitempty"`           //条目TID
	CreatedAt time.Time `gorm:"column:CreatedAt;index:CreatedAt_Index" json:",omitempty"` //登陆日期
	UpdatedAt time.Time `gorm:"column:UpdatedAt" json:",omitempty"`                       //修改日期
	//DeletedAt *time.Time `gorm:"column:DeletedAt" json:",omitempty"` //删除日期
	//Delete    int        `gorm:"column:Delete"`                //0=无，1=删除，
}*/
type Item struct {
	iface.SQLModels
	UserID   uint64   `gorm:"column:UserID;index:UserID"`                      //userid
	ItemID   ItemID   `gorm:"column:ItemID;index:ItemID"`                      //DataTableID
	ItemType ItemType `gorm:"column:ItemType;index:ItemType"`                  //DataTableID
	WorldKey string   `gorm:"column:WorldKey;index:WorldKeyAreaID;default:''"` //DataTableID
	AreaID   int64    `gorm:"column:AreaID;index:WorldKeyAreaID;default:'0'"`  //DataTableID
	Quantity int64    `gorm:"column:Quantity;default:'0'"`                     //
}

func (fi *Item) TableName() string {
	return "Item"
}
func (fi *Item) GetID() uint64 {
	return fi.ID
}

type KV struct {
	K string `gorm:"column:K;primary_key;unique"` //
	V string `gorm:"column:V"`                    //
}

func (kv *KV) ProtoMessage() *pb.ActionState {
	as := &pb.ActionState{}
	as.Message = kv.V
	as.Action = kv.K
	return as
}
func (KV) TableName() string {
	return "KV"
}

/*func (fi *Item) ProtoMessage() *pb.Item {
	return &pb.Item{
		ItemID:uint64(fi.ItemID),
		ItemType:uint64(fi.ItemType),
		Quantity:fi.Quantity,
	}
}*/

func (fi *Item) Persistence(DB *gorm.DB) error {

	return DB.Model(&Item{}).Where("ID=?", fi.ID).Updates(map[string]interface{}{"Quantity": fi.Quantity}).Error

}
func (fi *Item) ObtainCache(DB *gorm.DB, UserID uint64, ItemID ItemID, ItemType ItemType) error {

	err := DB.Where("UserID=? and ItemID=? and ItemType=?", UserID, ItemID, ItemType).First(fi).Error //SelectOne(user, "select * from User where Email=?", Email)
	if gorm.IsRecordNotFoundError(err) {
		fi.UserID = UserID
		fi.ItemID = ItemID
		fi.ItemType = ItemType
		err := DB.Create(fi).Error
		if glog.Error(err) {
			return err
		}
	}

	return err
}

type MassMail struct {
	iface.SQLModels
	Awards    string `gorm:"column:Awards;type:JSON"` //
	Title     string `gorm:"column:Title"`
	Content   string `gorm:"column:Content"`
	WorldKey  string `gorm:"column:WorldKey"`  //为空时，所有的世界
	AreaID    int64  `gorm:"column:AreaID"`    //-1时，所有区服
	UserID    uint64 `gorm:"column:UserID"`    //0为所部的用户
	ExpiryDay uint64 `gorm:"column:ExpiryDay"` //用效期天
}

func (*MassMail) TableName() string {
	return "MassMail"
}
