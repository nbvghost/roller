package socket

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/nbvghost/roller/db"
	"github.com/nbvghost/roller/entity"
	"github.com/nbvghost/roller/entity/pb"
	"github.com/nbvghost/roller/iface"
	"github.com/nbvghost/roller/translation"

	"github.com/nbvghost/gweb"
	"github.com/nbvghost/gweb/conf"
	"runtime/debug"
	"sync"

	"log"

	"net/http"

	"github.com/nbvghost/glog"
	"github.com/nbvghost/gweb/tool"

	"strconv"
	"time"

	"strings"
)

//var NotLogin                                =  map[string]interface{}{"Action":"NotLogin","Message":"您还没有登陆,请刷新再试"}//KV{K: "NOT_LOGIN", V: "您还没有登陆,请刷新再试"}
//var InvalidCookie                           = map[string]interface{}{"Action":"InvalidCookie","Message":"cookie无效"}
//var ForcedOffline = map[string]interface{}{"Action":"ForcedOffline","Message":"您在其它登陆的账号已经被迫下线"}
//var KeepLive = map[string]interface{}{}
//AccountLoggedDifferentLocation           = KV{K: "AccountLoggedDifferentLocation", V: "您的账号在异地登陆了"}                  //您的账号在异地登陆了
//ForcedOffline                            = KV{K: "ForcedOffline", V: "您在其它登陆的账号已经被迫下线"}                              //您在其它登陆的账号已经被迫下线
var DefaultWebSocket = &GameWebSocket{GlobalUser: &UserMap{}}

type GameWebSocket struct {
	//dao.BaseDao
	//MsgID int32
	//sync.RWMutex
	GlobalUser *UserMap
	Server     *http.Server
	//MySQL *db.MySQLConnector
	//Redis *db.RedisConnector
	//MessageProcessor *MessageProcessor
	ILoginBridge iface.ILoginBridge
	config       *entity.AppConfig
	protoMessage iface.IProtoMessage
	//SupportMessageType int
	//ItemAll map[mold.ItemID]map[mold.ItemType]mold.ItemIDItemType
}

func (ws *GameWebSocket) GetProtoMessage() iface.IProtoMessage {
	return ws.protoMessage
}
func (ws *GameWebSocket) SendBytesToWorldKeyUser(Package []byte, WorldKey string) {
	ws.GlobalUser.Range(func(UserID uint64, value *Cluster) bool {
		if value != nil {
			if strings.EqualFold(value.WorldKey, WorldKey) {
				err := value.SendSelfBytesPackage(Package)
				//err := service.SendActionMessage(action, value)
				glog.Error(err)
			}
		}
		return true
	})
}
func (ws *GameWebSocket) SendBytesToAreaUser(Package []byte, WorldKey string, AreaID int64) {
	ws.GlobalUser.Range(func(UserID uint64, value *Cluster) bool {
		if value != nil {

			if strings.EqualFold(value.WorldKey, WorldKey) && value.AreaID == AreaID {
				err := value.SendSelfBytesPackage(Package)
				//err := service.SendActionMessage(action, value)
				glog.Error(err)
			}
		}
		return true
	})
}
func (ws *GameWebSocket) SendBytesAllUser(Package []byte) {
	ws.GlobalUser.Range(func(UserID uint64, value *Cluster) bool {
		if value != nil {
			err := value.SendSelfBytesPackage(Package)
			//err := service.SendActionMessage(action, value)
			glog.Error(err)

		}
		return true
	})
}
func (ws *GameWebSocket) SendActionMessageAllUser(action *entity.Action) {
	ws.GlobalUser.Range(func(UserID uint64, value *Cluster) bool {
		if value != nil {

			err := ws.SendActionMessage(action, value)
			glog.Error(err)

		}
		return true
	})
}

func (ws *GameWebSocket) OfflineAll() {
	ws.GlobalUser.Range(func(UserID uint64, value *Cluster) bool {
		if value != nil {
			value.Close()
		}
		return true
	})
}

func (ws *GameWebSocket) SendActionMessage(action *entity.Action, cluster *Cluster) error {

	packageType := int32(-1)

	//var id int32 = 45 //4
	msgID := _getMsgID()

	var bodyLen int32 = 0

	bodyBuffer := bytes.NewBuffer([]byte{})

	//-----msg------
	msgBuffer := bytes.NewBuffer([]byte{})
	b, err := tool.JsonMarshal(action)
	if glog.Error(err) {
		return err
	}
	msgBuffer.Write(b)
	binary.Write(bodyBuffer, binary.BigEndian, &packageType)
	binary.Write(bodyBuffer, binary.BigEndian, msgBuffer.Bytes())

	bodyLen = int32(len(bodyBuffer.Bytes()))

	packBuffer := bytes.NewBuffer([]byte{})
	binary.Write(packBuffer, binary.BigEndian, &msgID)
	binary.Write(packBuffer, binary.BigEndian, &bodyLen)

	binary.Write(packBuffer, binary.BigEndian, bodyBuffer.Bytes())

	if cluster.ClientConn != nil {
		err := cluster.ClientConn.WriteMessage(websocket.BinaryMessage, packBuffer.Bytes())
		//err = websocket.Message.Send(cluster.ClientConn, packBuffer.Bytes())
		if glog.Error(err) {
			return err
		}
	}

	return nil

}
func (ws *GameWebSocket) StopSocketService() {
	glog.Debug("开始清理工作")

	clean := 0
	ws.GlobalUser.Range(func(UserID uint64, value *Cluster) bool {
		glog.Debug("清理：", value)
		value.Close()
		clean++
		return true
	})

	//持久化的是整个redis
	db.Connector.Redis.Persistence(db.Connector.MySQL.GetMySQL(), 0)

	if ws.Server != nil {
		ws.Server.Close()
	}
	glog.Trace("清理了：" + strconv.Itoa(clean) + "个用户")

}

func (ws *GameWebSocket) gameInfoAction(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("Access-Control-Allow-Origin", "*")

	result := make(map[string]interface{})
	clusters := make([]*Cluster, 0)
	ws.GlobalUser.Range(func(UserID uint64, value *Cluster) bool {

		clusters = append(clusters, value)
		return true

	})
	result["ClusterInfos"] = clusters
	b, _ := json.Marshal(&result)
	response.WriteHeader(http.StatusOK)
	response.Write(b)

}
func (ws *GameWebSocket) noticeSendEmailAction(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("Access-Control-Allow-Origin", "*")

	Title := request.FormValue("Title")
	Content := request.FormValue("Content")

	ItemID, _ := strconv.ParseUint(request.FormValue("ItemID"), 10, 64)
	ItemType, _ := strconv.ParseUint(request.FormValue("ItemType"), 10, 64)
	ItemQuantity, _ := strconv.ParseUint(request.FormValue("ItemQuantity"), 10, 64)

	fmt.Println(Title)
	fmt.Println(Content)
	fmt.Println(ItemID)
	fmt.Println(ItemType)
	fmt.Println(ItemQuantity)

	_UserID := request.FormValue("UserID")
	UserID, _ := strconv.ParseUint(_UserID, 10, 64)
	if UserID == 0 || len(_UserID) == 0 || strings.EqualFold(_UserID, "0") {

		//err := service.Global.MassMail.AddMassMail(service.GetOrm(), [][]uint64{{FunctionType, Quantity}}, Title, Content)
		//return &gweb.JsonResult{Data: (&dao.ActionStatus{}).SmartError(err, "操作成功", nil)}

	} else {
		//err := service.Global.MassMail.AddCustomUserMail(service.GetOrm(), service.GetRedis(), UserID, Title, Content, [][]uint64{{FunctionType, Quantity}}, 0)
		//return &gweb.JsonResult{Data: (&dao.ActionStatus{}).SmartError(err, "操作成功", nil)}
	}

}
func (ws *GameWebSocket) noticeSendNoticeAction(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("Access-Control-Allow-Origin", "*")

	result := make(map[string]interface{})
	clusters := make([]*Cluster, 0)
	ws.GlobalUser.Range(func(UserID uint64, value *Cluster) bool {

		clusters = append(clusters, value)
		return true

	})
	result["ClusterInfos"] = clusters
	b, _ := json.Marshal(&result)
	response.WriteHeader(http.StatusOK)
	response.Write(b)

}
func (ws *GameWebSocket) noticeSendForceOfflineAction(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("Access-Control-Allow-Origin", "*")

	result := make(map[string]interface{})
	clusters := make([]*Cluster, 0)
	ws.GlobalUser.Range(func(UserID uint64, value *Cluster) bool {

		clusters = append(clusters, value)
		return true

	})
	result["ClusterInfos"] = clusters
	b, _ := json.Marshal(&result)
	response.WriteHeader(http.StatusOK)
	response.Write(b)

}
func (ws *GameWebSocket) StartWebSocket(protoMessage iface.IProtoMessage, config *entity.AppConfig, logicIndex int) error {
	//logic.IP, logic.TcpPort
	logic := config.LogicServer[logicIndex]

	var defaultHandler = http.DefaultServeMux
	ws.config = config
	ws.protoMessage = protoMessage
	//http.Handle("/game/server", websocket.Handler(ws.WebSocketHandler))
	http.HandleFunc("/game/server", ws.WebSocketHandler)
	http.HandleFunc("/game/info", ws.gameInfoAction)

	http.HandleFunc("/game/notice/send/email", ws.noticeSendEmailAction)
	http.HandleFunc("/game/notice/send/notice", ws.noticeSendNoticeAction)
	http.HandleFunc("/game/notice/send/force_offline", ws.noticeSendForceOfflineAction)

	ws.Server = &http.Server{
		Addr:         logic.IP + ":" + logic.TcpPort,
		Handler:      defaultHandler,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
		//MaxHeaderBytes: 1 << 20,
	}
	ws.Server.SetKeepAlivesEnabled(true)
	glog.Trace("start websocket at：" + ws.Server.Addr)
	var err error

	if ws.config.EnableHttps {
		conf.Config.TLSCertFile = config.CertFile
		conf.Config.TLSKeyFile = config.KeyFile
		//err = ws.Server.ListenAndServeTLS(ws.config.CertFile, ws.config.KeyFile)
		gweb.StartServer(defaultHandler, nil, ws.Server)
	} else {
		//err = ws.Server.ListenAndServe()
		gweb.StartServer(defaultHandler, ws.Server, nil)
	}

	return err

}

var upgrader = websocket.Upgrader{}

func (ws *GameWebSocket) WebSocketHandler(response http.ResponseWriter, request *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool {

		glog.Trace(r.Header)

		return true
	}
	conn, err := upgrader.Upgrade(response, request, nil)
	if glog.Error(err) {
		log.Print("upgrade:", err)
		return
	}

	glog.Trace("WebSocket", "start")

	UserID := uint64(0)

	WsKey := request.Header.Get("Sec-WebSocket-Key")

	currentWebsocketUser := &Cluster{}
	currentWebsocketUser.Init(conn, WsKey)

	//version := ws.Request().URL.Query().Get("version")
	//log.Println(version)
	token := request.URL.Query().Get("token")

	if strings.EqualFold(token, "") {
		err := ws.SendActionMessage(entity.LoginFailedInvalidToken, currentWebsocketUser)
		if glog.Error(err) {

		}
		return

	} else {

		UserID, _ = strconv.ParseUint(ws.ILoginBridge.CheckToke(token), 10, 64)
		if UserID == 0 {
			err := ws.SendActionMessage(entity.LoginFailedInvalidToken, currentWebsocketUser)
			if glog.Error(err) {

			}
			return
		}
	}

	//ws.GlobalUser.Get(UserID)
	glog.Trace("WebSocket", "setClusterVar")
	t := time.Now().UnixNano()
	status := ws.setClusterVar(currentWebsocketUser, UserID)
	glog.Trace("loginTime", float64(time.Now().UnixNano()-t)/1000/1000)
	if status != nil {
		err := currentWebsocketUser.SendSelfPackage(status)
		glog.Error(err)
		return
	}

	//cluster.LastOperationTime = time.Now().Unix()
	defer func() {

		//currentWebsocketUser.Close()
		//service.CloseWebSocket(false, websocketUser)
	}()

	ws.execute(currentWebsocketUser)
}

func (ws *GameWebSocket) execute(currentWebsocketUser *Cluster) {

	defer func() {
		if r := recover(); r != nil {
			//_, file, line, _ := runtime.Caller(1)
			//log.Println(file, line, r)
			glog.Trace(r)
			debug.PrintStack()
		}
		currentWebsocketUser.Close()
	}()

	for {

		messageType, Package, err := currentWebsocketUser.ClientConn.ReadMessage()
		//n, err := currentWebsocketUser.ClientConn.Conn.Read(buffer[0:])
		if glog.Error(err) {
			//glog.Debug(err)
			return
		}

		currentWebsocketUser.PushBuffer(MessagePackage{MessageType: messageType, Package: Package})

	}
}
func (ws *GameWebSocket) setClusterVar(currentWebsocketUser *Cluster, UserID uint64) (status *pb.ActionState) {
	currentWebsocketUser.UserID = UserID
	wsUsers, ok := ws.GlobalUser.Get(UserID)
	if ok {

		if strings.EqualFold(currentWebsocketUser.WebSocketKey, wsUsers.WebSocketKey) == false {

			err := ws.SendActionMessage(entity.AccountLoggedDifferentLocation, wsUsers)
			glog.Error(err)

			wsUsers.Close()
			//wsUsers.WebSocketKey = currentWebsocketUser.WebSocketKey
			//wsUsers.ClientConn = currentWebsocketUser.ClientConn
			//currentWebsocketUser.SetSession(wsUsers.GetSession())

			currentWebsocketUser.UserID = wsUsers.UserID
			currentWebsocketUser.WorldKey = wsUsers.WorldKey
			currentWebsocketUser.AreaID = wsUsers.AreaID
			currentWebsocketUser.InputPackageNum = wsUsers.InputPackageNum
			currentWebsocketUser.OutPackageNum = wsUsers.OutPackageNum
			currentWebsocketUser.InputBytes = wsUsers.InputBytes
			currentWebsocketUser.OutBytes = wsUsers.OutBytes
			currentWebsocketUser.ExecuteUnixNano = wsUsers.ExecuteUnixNano
			currentWebsocketUser.Session = wsUsers.Session

			ws.GlobalUser.Put(UserID, currentWebsocketUser)

		}
		return nil
	} else {
		user, err := ws.ILoginBridge.Online(UserID, currentWebsocketUser)
		if glog.Error(err) {
			return entity.GetKV(entity.Error, err.Error()).ProtoMessage()
		}

		glog.Trace("Online Session:", currentWebsocketUser.Session)

		if currentWebsocketUser.Session == nil {
			panic(errors.New("NotFound Session"))
		}

		glog.Trace("Online User:", user)

		currentWebsocketUser.UserID = user.GetUserID()
		currentWebsocketUser.WorldKey = user.GetWorldKey()
		currentWebsocketUser.AreaID = user.GetAreaID()

		//map[mold.ItemID]map[mold.ItemType]mold.ItemIDItemType

		itemAll := translation.Translation.Item.GetItemAll()
		for akey := range itemAll {
			for bkey := range itemAll[akey] {
				types := itemAll[akey][bkey]
				err, _ = currentWebsocketUser.Translation.Item.Cache(UserID, currentWebsocketUser.WorldKey, currentWebsocketUser.AreaID, types.ItemID, types.ItemType)
				if glog.Error(err) {
					return entity.GetKV(entity.Error, err.Error()).ProtoMessage()
				}
			}
		}

		//items:=&pb.Items{}
		//items.Items=currentWebsocketUser.Translation.Item.ProtoMessage(currentWebsocketUser.Translation.Item.ObtainCache(currentWebsocketUser.UserID))
		//currentWebsocketUser.SendSelfPackage(items)

		//currentWebsocketUser.session=session
		//currentWebsocketUser.SetSession(session)
		ws.GlobalUser.Put(UserID, currentWebsocketUser)
		//currentWebsocketUser.GetSession().GetUser().OnLineTime = time.Now()
		return nil
	}

}

type UserMap struct {
	//Map map[uint64]WebSocketUser
	_map sync.Map
}

func (ssm *UserMap) Put(UserID uint64, value *Cluster) {

	ssm._map.Store(UserID, value) // Map[key] = value

}
func (ssm *UserMap) Range(f func(UserID uint64, value *Cluster) bool) {
	ssm._map.Range(func(key, value interface{}) bool {
		return f(key.(uint64), value.(*Cluster))
	})
}
func (ssm *UserMap) Del(UserID uint64) {

	//delete(ssm.Map, k)

	ssm._map.Delete(UserID)

	//db.NotifyAll(&db.Message{db.Socket_Type_2_STC,k})
}

func (ssm *UserMap) Get(UserID uint64) (*Cluster, bool) {
	v, ok := ssm._map.Load(UserID)
	if v == nil {
		return nil, false
	}
	if ok == false {
		return nil, false
	}
	return v.(*Cluster), ok
}
