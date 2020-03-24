package translation

import (
	"github.com/jinzhu/gorm"
	"github.com/nbvghost/glog"
	"github.com/nbvghost/roller/mold"
	"strconv"
	"strings"
)

type KVTranslation struct {
}

func (service *KVTranslation) Save(DB *gorm.DB, kv *mold.KV) error {
	item := service.GetByKey(DB, kv.K)
	return DB.Model(&mold.KV{}).Where("K=?", item.K).Updates(map[string]interface{}{"V": kv.V}).Error
}
func (service *KVTranslation) GetByKey(DB *gorm.DB, Key string) *mold.KV {
	item := mold.KV{}
	err := DB.Where("K=?", Key).First(&item).Error //SelectOne(user, "select * from User where Email=?", Email)
	if gorm.IsRecordNotFoundError(err) {
		item.K = Key
		DB.Save(&item)
	}
	return &item

}
func (service *KVTranslation) SetByKey(DB *gorm.DB, Key string, Value string) error {
	err := DB.Model(&mold.KV{}).Where("K=?", Key).Updates(map[string]interface{}{"K": Key, "V": Value}).Error //SelectOne(user, "select * from User where Email=?", Email)
	return err

}
func (service *KVTranslation) ListByKey(DB *gorm.DB, Keys []string) []mold.KV {
	var item []mold.KV
	err := DB.Where("K in (?)", Keys).Find(&item).Error //SelectOne(user, "select * from User where Email=?", Email)
	glog.Error(err)
	return item
}

/*func (service *_KV) ChangeByKey(DB *gorm.DB, Key string, Value string) error {
	//err := service.ChangeMap(DB) //SelectOne(user, "select * from User where Email=?", Email
	err := DB.Model(dao.KV{}).Where("K=?", Key).Updates(map[string]string{"V": Value}).Error
	glog.Error(err)
	return err
}*/
func (service *KVTranslation) GetByBool(DB *gorm.DB, Key string) bool {
	item := service.GetByKey(DB, Key)
	if strings.EqualFold(item.V, "true") {
		return true
	} else {
		return false
	}
}
func (service *KVTranslation) SetByBool(DB *gorm.DB, Key string, Value bool) error {
	err := DB.Model(&mold.KV{}).Where("K=?", Key).Updates(map[string]interface{}{"K": Key, "V": strconv.FormatBool(Value)}).Error //SelectOne(user, "select * from User where Email=?", Email)
	return err
}
func (service *KVTranslation) GetByInt(DB *gorm.DB, Key string) int64 {
	item := service.GetByKey(DB, Key)
	n, _ := strconv.ParseInt(item.V, 10, 64)
	return n
}
func (service *KVTranslation) SetByInt(DB *gorm.DB, Key string, Value int64) error {
	err := DB.Model(&mold.KV{}).Where("K=?", Key).Updates(map[string]interface{}{"K": Key, "V": strconv.FormatInt(Value, 10)}).Error //SelectOne(user, "select * from User where Email=?", Email)
	return err
}
