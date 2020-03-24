package app

import (
	"errors"
	"github.com/nbvghost/roller/db"
	"github.com/nbvghost/roller/entity"
	"github.com/nbvghost/roller/iface"
	"github.com/nbvghost/roller/mold"
	"github.com/nbvghost/roller/translation"
	"strconv"
	"sync"
	"time"

	"github.com/nbvghost/glog"
	"github.com/nbvghost/gweb/thread"
	"github.com/nbvghost/gweb/tool"
	"github.com/nbvghost/roller/socket"
	"io/ioutil"
	"net/http"
	"strings"
)

var App = &Application{LogicIndex: -1, GameSocket: socket.DefaultWebSocket}
var tt = &timeTask{}

type Application struct {
	//dataTables map[string]interface{}
	Config      *entity.AppConfig
	_cleaning   chan int
	LogicIndex  int
	OverStart   func()
	GetItemType func() map[mold.ItemID]map[mold.ItemType]mold.ItemIDItemType
	GameSocket  *socket.GameWebSocket
}

func (app *Application) initListeningExit() {
	app._cleaning = make(chan int)
	thread.NewCoroutine(func() {
		sn := thread.ListeningSignal()

		for {
			select {
			case v := <-sn:
				glog.Debug("读到退出信息号")
				glog.Trace("收到退出信号：", v.String())
				glog.Debug("服务准备退出")
				socket.DefaultWebSocket.StopSocketService()
				glog.Trace("清理结束，退出服务程序")
				//signal.Stop(v)
				//time.Sleep(10 * time.Second)
				app.overCleanStatus()
			}
		}

	}, func(v interface{}, stack []byte) {
		glog.Trace(v)
		glog.Trace(string(stack))
	})
}
func (app *Application) getCleanStatus() {
	<-app._cleaning
	close(app._cleaning)
}
func (app *Application) overCleanStatus() {
	app._cleaning <- 1
}
func (app *Application) GetClusterInfos() []map[string]interface{} {

	return tt.GetClusterInfos()
}
func (app *Application) RegisterSQLModels(model iface.ISQLModels) {
	db.Connector.MySQL.RegisterSQLModels(model)
}
func (app *Application) LoadAppConfig(appConfigFileName string) error {
	if app.Config == nil {
		app.Config = &entity.AppConfig{}
		err := app.Config.Load(appConfigFileName)
		if glog.Error(err) {
			return err
		}
	}
	return nil
}
func (app *Application) Start(logicIndex int, protoMessage iface.IProtoMessage, loginBridge iface.ILoginBridge) error {
	if app.Config == nil {
		return errors.New("NotFound AppConfig.Please Call Application.LoadAppConfig Method")
	}
	if logicIndex > len(app.Config.LogicServer)-1 {
		logicIndex = len(app.Config.LogicServer) - 1
	}
	logic := app.Config.LogicServer[logicIndex]
	glog.Trace("启动Socket:", logicIndex, logic.IP, logic.TcpPort)

	app.LogicIndex = logicIndex

	app.initListeningExit()

	if len(db.Connector.MySQL.SQLModels()) <= 0 {
		glog.Trace("NotFound SQL Models To Create.Please Call Application.RegisterSQLModels Method")
	}
	db.Connector.Init(app.Config, logicIndex)

	socket.DefaultWebSocket.ILoginBridge = loginBridge

	if app.GetItemType == nil {
		return errors.New("NotFound Item Information.Please Overwrite Application.GetItemType Method")
		//GetItemType
	}
	allItemType := app.GetItemType()

	if app.OverStart != nil {
		app.OverStart()
	}

	translation.Translation.Item.SetItemAll(allItemType)

	ttList := tt.Start()
	err := socket.DefaultWebSocket.StartWebSocket(protoMessage, app.Config, logicIndex)
	tt.Stop(ttList)
	app.getCleanStatus()

	return err

}
func (app *Application) ReadDataTables(dataTables map[string]interface{}) error {
	/*if app.dataTables==nil{
		app.dataTables =make(map[string]interface{})
	}*/

	for key := range dataTables {
		err := app.ReadDataTable(key, dataTables[key])
		if err != nil {
			return err
		}
	}
	/*thread.NewCoroutine(func() {
		ticker := time.NewTicker(10 * time.Minute)
		//ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				readDataTable()
			}
		}

	}, func(v interface{}, stack []byte) {
		glog.Trace(v)
		glog.Trace(string(stack))
	})*/
	return nil
}
func (app *Application) ReadDataTable(dataTablePath string, dataTable interface{}) error {
	/*if app.dataTables==nil{
		app.dataTables =make(map[string]interface{})
	}*/

	if strings.EqualFold(dataTablePath, "") {
		err := errors.New("dataTablePath 不能为空")
		glog.Error(err)
		return err
	}

	if strings.Contains(dataTablePath, "http://") || strings.Contains(dataTablePath, "https://") {

		response, err := http.Get(dataTablePath)
		if glog.Error(err) {
			glog.Debugf("data table %v 加载出错", dataTablePath)
			return err
		}
		b, err := ioutil.ReadAll(response.Body)
		if glog.Error(err) {
			glog.Debugf("data table %v 加载出错", dataTablePath)
			return err
		}
		response.Body.Close()
		//Global.Lock()
		err = tool.JsonUnmarshal(b, dataTable)
		if glog.Error(err) {
			return err
		}
		return err
	} else {
		b, err := ioutil.ReadFile(dataTablePath)
		if glog.Error(err) == false {
			err = tool.JsonUnmarshal(b, dataTable)
			if glog.Error(err) {
				return err
			}
		}

		return err

		//Global.Unlock()
	}

}

type timeTask struct {
	ClusterInfos []map[string]interface{}
	sync.RWMutex
	//SCoinRankList []dao.SCoinRank
}

func (ts *timeTask) Stop(_ts []*time.Ticker) {

	for index := range _ts {

		_ts[index].Stop()

	}

}
func (ts *timeTask) GetClusterInfos() []map[string]interface{} {
	ts.RLock()
	defer ts.RUnlock()
	return ts.ClusterInfos
}
func (ts *timeTask) Start() []*time.Ticker {
	results := make([]*time.Ticker, 0)

	a, _ := time.ParseDuration(strconv.FormatInt(App.Config.PersistenceTime, 10) + "s")
	at := time.NewTicker(a)

	b, _ := time.ParseDuration("1s")
	bt := time.NewTicker(b)

	results = append(results, at)
	results = append(results, bt)
	thread.NewCoroutine(func() {

		for range at.C {
			ts.persistence()
		}

	}, func(v interface{}, stack []byte) {
		glog.Trace(v)
		glog.Trace(string(stack))
	})

	thread.NewCoroutine(func() {

		for range bt.C {
			ts.printSession()
		}

	}, func(v interface{}, stack []byte) {
		glog.Trace(v)
		glog.Trace(string(stack))
	})

	return results
}
func (ts *timeTask) persistence() {
	db.Connector.Redis.Persistence(db.Connector.MySQL.GetMySQL(), 0)

	socket.DefaultWebSocket.GlobalUser.Range(func(UserID uint64, wsUsers *socket.Cluster) bool {
		//socket.DefaultWebSocket.GlobalUser.PersistenceFunctionItem(wsUsers.GetSession())
		//err:=wsUsers.Session.Persistence(db.Connector.MySQL.GetMySQL())
		//glog.Error(err)
		/*if wsUsers.IsClose() == false {
			Global.FunctionItem.UpdateSCoinRank(wsUsers)
		}*/
		return true
	})
}
func (ts *timeTask) printSession() {

	_ClusterInfos := make([]map[string]interface{}, 0)
	//service.ClusterCount = 0
	st := time.Now().UnixNano()
	socket.DefaultWebSocket.GlobalUser.Range(func(UserID uint64, wsUsers *socket.Cluster) bool {

		info := map[string]interface{}{
			"UserID": wsUsers.UserID,
			//"MessageQueueLength": wsUsers.GetMessageBufferLen(),
			"IsClose":         wsUsers.IsClose(),
			"InputPackageNum": wsUsers.InputPackageNum,
			"OutPackageNum":   wsUsers.OutPackageNum,
			"InputBytes":      wsUsers.InputBytes,
			"OutBytes":        wsUsers.OutBytes,
			"ExecuteUnixNano": wsUsers.ExecuteUnixNano,
			"CreateAt":        time.Now().Format("2006-01-02 15:04:05"),
		}
		_ClusterInfos = append(_ClusterInfos, info)
		return true
	})
	st = time.Now().UnixNano() - st
	glog.Debug("在线人数:"+strconv.Itoa(len(_ClusterInfos)), "用时(s)：", float64(st)/float64(1000)/float64(1000)/float64(1000))

	ts.Lock()
	ts.ClusterInfos = _ClusterInfos
	ts.Unlock()
	//log.Println(Global.User.GetUserByRandAndThree())
}
