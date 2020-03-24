package translation

import (
	"github.com/jinzhu/gorm"
	"github.com/nbvghost/gweb/tool"
	"github.com/nbvghost/roller/mold"
)

type MassMailTranslation struct {
}

func (service *MassMailTranslation) AddMassMail(DB *gorm.DB, Awards map[mold.ItemID]map[mold.ItemType]uint64, WorldKey string, AreaID int64, UserID uint64, Title, Content string, ExpiryDay uint64) error {
	item := &mold.MassMail{}
	AwardsBytes, _ := tool.JsonMarshal(&Awards)
	item.Awards = string(AwardsBytes)
	item.Title = Title
	item.Content = Content
	item.WorldKey = WorldKey
	item.AreaID = AreaID
	item.UserID = UserID
	item.ExpiryDay = ExpiryDay
	err := DB.Create(&item).Error
	return err
}
