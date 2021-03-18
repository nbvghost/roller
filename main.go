package main

import (
	"github.com/nbvghost/glog"
	"github.com/nbvghost/gweb/tool"
	"github.com/nbvghost/roller/app"
	"github.com/nbvghost/roller/iface"
	"github.com/nbvghost/roller/mold"
)

type LoginBridgeService struct {
}

/*func (service *LoginBridgeService) GetClusterInfos() []map[string]interface{} {

	return Global.TimeTask.GetClusterInfos()
}*/
func (service *LoginBridgeService) Offline(UserID uint64, cluster iface.ICluster) {

}
func (service *LoginBridgeService) CheckToken(Token string) string {

	return tool.CipherDecrypterData(Token)
}
func (service *LoginBridgeService) Online(UserID uint64, cluster iface.ICluster) (iface.IUser, error) {

	return nil, nil
}

func main() {

	application := app.App
	application.LoadAppConfig("AppConfig.json")

	application.GetItemType = func() map[mold.ItemID]map[mold.ItemType]mold.ItemIDItemType {

		return map[mold.ItemID]map[mold.ItemType]mold.ItemIDItemType{}

	}

	application.OverStart = func() {

	}

	err := application.Start(0, nil, &LoginBridgeService{})
	glog.Error(err)
}
