package translation

import (
	"encoding/json"
	"github.com/jinzhu/gorm"
	"github.com/nbvghost/glog"
	"github.com/nbvghost/roller/db"
	"github.com/nbvghost/roller/entity"
	"github.com/nbvghost/roller/mold"
	"strconv"
	"strings"
)

type ItemOperation struct {
	Operation int32 //-1 减  1 加
	ItemID    mold.ItemID
	ItemType  mold.ItemType
	Quantity  uint64
}
type ItemTranslation struct {
	//changes  []ItemOperation
	itemAll map[mold.ItemID]map[mold.ItemType]mold.ItemIDItemType
}

func (service *ItemTranslation) GetItemAll() map[mold.ItemID]map[mold.ItemType]mold.ItemIDItemType {
	return service.itemAll
}
func (service *ItemTranslation) SetItemAll(itemAll map[mold.ItemID]map[mold.ItemType]mold.ItemIDItemType) {
	service.itemAll = itemAll
}

func (service *ItemTranslation) SplitKey(key string) (mold.ItemID, mold.ItemType) {
	keys := strings.Split(key, ":")
	if len(keys) < 2 {
		return 0, 0
	}
	ItemID, _ := strconv.ParseUint(keys[0], 10, 64)
	ItemType, _ := strconv.ParseUint(keys[1], 10, 64)
	return mold.ItemID(ItemID), mold.ItemType(ItemType)
}
func (service *ItemTranslation) MargeKey(ItemID mold.ItemID, ItemType mold.ItemType) string {

	return strconv.FormatUint(uint64(ItemID), 10) + ":" + strconv.FormatUint(uint64(ItemType), 10)
}

/*func (service *ItemService) ProtoMessage(items []mold.Item) []*pb.Item {
	r:=make([]*pb.Item,0)
	for index:=range items{
		r=append(r,items[index].ProtoMessage())
	}
	return r
}*/
func (service *ItemTranslation) ObtainCache(UserID uint64) []mold.Item {
	var items []mold.Item
	var mysql = db.Connector.MySQL.GetMySQL()
	mysql.Where("UserID=?", UserID).Find(&items)

	for index := range items {

		service.Get(UserID, items[index].ItemID, items[index].ItemType)

	}
	return items
}
func (service *ItemTranslation) List(UserID uint64) []mold.Item {
	var redis = db.Connector.Redis

	list := make([]mold.Item, 0)
	maps, err := redis.HGetAll(&mold.Item{}, UserID)
	if glog.Error(err) {
		return list
	}
	for key, _ := range maps {
		item := mold.Item{}
		json.Unmarshal([]byte(maps[key]), &item)
		list = append(list, item)
	}
	return list
}
func (service *ItemTranslation) Sub(UserID uint64, ItemID mold.ItemID, ItemType mold.ItemType, Quantity int64) (*mold.KV, *mold.Item) {

	var itemIDItemType mold.ItemIDItemType
	_, ok := service.itemAll[ItemID]
	if ok {
		itemIDItemType, ok = service.itemAll[ItemID][ItemType]
	}

	err, item := service.Get(UserID, ItemID, ItemType)
	if glog.Error(err) {
		//var kv = entity.KV{K:entity.SQLError.K,V:entity.SQLError.V}
		//kv.V = fmt.Sprintf(kv.V, err)
		return entity.GetKV(entity.SQLError, err.Error()), item
	}
	if item.Quantity < Quantity {
		//var kv = entity.KV{K:entity.InsufficientItem.K,V:entity.InsufficientItem.V}
		return entity.GetKV(entity.InsufficientItem, itemIDItemType.Name), nil
	}
	item.Quantity = item.Quantity - Quantity

	var redis = db.Connector.Redis
	err = redis.HSet(item.UserID, service.MargeKey(ItemID, ItemType), item)
	if glog.Error(err) {
		//var kv = entity.KV{K:entity.SQLError.K,V:entity.SQLError.V}
		//kv.V = fmt.Sprintf(kv.V, err)
		return entity.GetKV(entity.SQLError, err.Error()), item
	}
	return nil, item

}
func (service *ItemTranslation) SetMax(UserID uint64, ItemID mold.ItemID, ItemType mold.ItemType, Quantity int64) (error, *mold.Item, int64) {
	err, item := service.Get(UserID, ItemID, ItemType)
	if glog.Error(err) {
		return err, item, 0
	}
	change := Quantity - item.Quantity
	if change <= 0 {
		return nil, item, 0
	}
	item.Quantity = Quantity

	var redis = db.Connector.Redis
	err = redis.HSet(item.UserID, service.MargeKey(ItemID, ItemType), item)
	if glog.Error(err) {
		return err, item, change
	}
	return nil, item, change
}
func (service *ItemTranslation) Set(UserID uint64, ItemID mold.ItemID, ItemType mold.ItemType, Quantity int64) (error, *mold.Item, int64) {
	err, item := service.Get(UserID, ItemID, ItemType)
	if glog.Error(err) {
		return err, item, 0
	}
	change := Quantity - item.Quantity
	item.Quantity = Quantity

	var redis = db.Connector.Redis
	err = redis.HSet(item.UserID, service.MargeKey(ItemID, ItemType), item)
	if glog.Error(err) {
		return err, item, change
	}
	return nil, item, change
}
func (service *ItemTranslation) Add(UserID uint64, ItemID mold.ItemID, ItemType mold.ItemType, Quantity int64) (error, *mold.Item) {
	err, item := service.Get(UserID, ItemID, ItemType)
	if glog.Error(err) {
		return err, item
	}

	item.Quantity = item.Quantity + Quantity

	var redis = db.Connector.Redis
	err = redis.HSet(item.UserID, service.MargeKey(ItemID, ItemType), item)
	if glog.Error(err) {
		return err, item
	}
	return nil, item
}

//缓存 redis and 创建
func (service *ItemTranslation) Cache(UserID uint64, WorldKey string, AreaID int64, ItemID mold.ItemID, ItemType mold.ItemType) (error, *mold.Item) {

	var redis = db.Connector.Redis

	var mysql = db.Connector.MySQL.GetMySQL()
	var fi mold.Item = mold.Item{}
	err := redis.HGet(UserID, service.MargeKey(ItemID, ItemType), &fi)
	if glog.Error(err) {

		err = mysql.Where("UserID=? and ItemID=? and ItemType=?", UserID, ItemID, ItemType).First(&fi).Error //SelectOne(user, "select * from User where Email=?", Email)
		if gorm.IsRecordNotFoundError(err) {
			fi.UserID = UserID
			fi.ItemID = ItemID
			fi.ItemType = ItemType
			fi.WorldKey = WorldKey
			fi.AreaID = AreaID
			err := mysql.Create(&fi).Error
			if glog.Error(err) {
				//tx.Rollback()
				return err, nil
			}
			//tx.Commit()
		} else {
			//tx.Commit()
		}
		err = redis.HSet(UserID, service.MargeKey(ItemID, ItemType), &fi)
		if glog.Error(err) {
			return err, nil
		}
		return nil, &fi
	} else {
		return nil, &fi
	}
}
func (service *ItemTranslation) Get(UserID uint64, ItemID mold.ItemID, ItemType mold.ItemType) (error, *mold.Item) {
	var redis = db.Connector.Redis
	var mysql = db.Connector.MySQL.GetMySQL()
	var fi mold.Item = mold.Item{}
	err := redis.HGet(UserID, service.MargeKey(ItemID, ItemType), &fi)
	if glog.Error(err) {

		err = mysql.Where("UserID=? and ItemID=? and ItemType=?", UserID, ItemID, ItemType).First(&fi).Error //SelectOne(user, "select * from User where Email=?", Email)
		if gorm.IsRecordNotFoundError(err) {
			return err, nil
		} else {
			//tx.Commit()
		}
		err = redis.HSet(UserID, service.MargeKey(ItemID, ItemType), &fi)
		if glog.Error(err) {
			return err, nil
		}
		return nil, &fi
	} else {
		return nil, &fi
	}
}
