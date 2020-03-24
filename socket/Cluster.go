package socket

import (
	"bytes"
	"encoding/binary"
	"github.com/gorilla/websocket"
	"github.com/nbvghost/roller/db"
	"github.com/nbvghost/roller/entity"
	"github.com/nbvghost/roller/iface"
	"github.com/nbvghost/roller/translation"
	"log"
	"strings"

	"reflect"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"

	"github.com/nbvghost/glog"
)

//var UNKNOWN_TYPE = pb.ActionStatus{Success: false, Message: fmt.Sprintf(ActionCode.V, args...)} //service.Global.ActionStatus.GetAS(ac.UNKNOWN_TYPE, PackageName)
type MessagePackage struct {
	MessageType int
	Package     []byte
}

//as := pb.ActionStatus{}
//	as.Success = false
//	as.Message = fmt.Sprintf(ActionCode.V, args...)
//	as.Action = ActionCode.K
//	return &as
type Cluster struct {
	WebSocketKey    string                          `json:"-"`
	ClientConn      *websocket.Conn                 `json:"-"`
	_IsClose        bool                            `json:"-"`
	Translation     *translation.PackageTranslation `json:"-"`
	Session         iface.ICache                    `json:"-"`
	UserID          uint64
	WorldKey        string
	AreaID          int64
	InputPackageNum uint64
	OutPackageNum   uint64
	InputBytes      uint64
	OutBytes        uint64
	ExecuteUnixNano int64
	sync.RWMutex    `json:"-"`
}

func (cluster *Cluster) Init(Conn *websocket.Conn, WsKey string) {
	if cluster.ClientConn != nil && cluster.WebSocketKey != "" {
		glog.Trace("Cluster Are Duplicate Init")
		return
	}
	cluster.ClientConn = Conn
	cluster.WebSocketKey = WsKey

	cluster.Translation = translation.Translation

	//cluster.msgBuffer = make(chan MessagePackage, 1000)
	//cluster.closeChan = make(chan bool,1)

	/*thread.NewCoroutine(func() {
		cluster.runningOpenPackage()
	}, func(v interface{}, stack []byte) {
		glog.Trace(v)
		glog.Trace(string(stack))
	})*/

}
func (cluster *Cluster) SetSession(session iface.ICache) {
	cluster.Session = session
}

func (cluster *Cluster) handler(messagePackage MessagePackage) {
	if cluster.IsClose() == true {
		glog.Trace("客户端已经关闭")
		return
	}

	cluster.InputPackageNum++
	cluster.InputBytes += uint64(len(messagePackage.Package))

	ht := time.Now().UnixNano()

	var id int32
	var bodyLen int32
	var packageType int32

	readBuffer := bytes.NewBuffer(messagePackage.Package)
	binary.Read(readBuffer, binary.BigEndian, &id)
	binary.Read(readBuffer, binary.BigEndian, &bodyLen)
	binary.Read(readBuffer, binary.BigEndian, &packageType)

	packageName := DefaultWebSocket.GetProtoMessage().GetMessageName(packageType) //pb.MessageName_MsgType_name[packageType]
	if strings.EqualFold(packageName, "") {
		glog.Trace("无效的包", "packageName", packageName, "packageType", packageType)
		return
	}

	if bodyLen < 4 {
		log.Println("发送的了一个空的包")
		err := cluster.SendSelfPackage(entity.GetKV(entity.Empty, packageName).ProtoMessage())
		glog.Error(err)
		return
	}

	t := DefaultWebSocket.GetProtoMessage().GetHandler(packageType) //iface.GetHandlerType(packageName)
	if t == nil {
		//log.Panic("没有找到消息包：" + PackageName)
		err := cluster.SendSelfPackage(entity.GetKV(entity.UnknownType, packageName).ProtoMessage())
		glog.Error(err)
		return
	}

	packageBody := make([]byte, bodyLen-4)
	binary.Read(readBuffer, binary.BigEndian, &packageBody)

	var msg = reflect.New(proto.MessageType("pb." + packageName).Elem()).Interface().(proto.Message)
	//log.Println(msg)
	err := proto.Unmarshal(packageBody, msg)
	if glog.Error(err) {
		err = cluster.SendSelfPackage(entity.GetKV(entity.UnknownError, packageName).ProtoMessage()) ///iface.GetAS(ac.UNKNOWN_ERROR, err.Error()))
		glog.Error(err)
		return
	}

	v := reflect.New(t.Elem())
	method := v.MethodByName("Handler")
	//glog.Trace(method.IsValid())
	//glog.Trace(method.IsZero())
	//glog.Trace(method.IsNil())
	//glog.Trace("---------------------")
	if method.IsValid() == false {
		cluster.SendSelfPackage(entity.GetKV(entity.UnknownError, t.String()).ProtoMessage())
		return
	}
	_as := method.Call([]reflect.Value{reflect.ValueOf(msg), reflect.ValueOf(cluster)})[0]
	as := _as.Interface().(*entity.MessageAction)

	if as.Error != nil {
		as.Data = entity.GetKV(entity.Error, as.Error.Error()).ProtoMessage()
	}
	err = cluster.SendSelfPackage(as.Data)
	if glog.Error(err) {

	}

	glog.Trace(map[string]interface{}{"MessageData": proto.CompactTextString(msg), "UserID": cluster.UserID, "Type": "HandleMessage", "MessageName": packageName, "ExecuteUnixNano": ht})
	ht = time.Now().UnixNano() - ht
	cluster.ExecuteUnixNano += ht
	return
}

/*func (cluster *Cluster) runningOpenPackage() {
	for {
		select {
		case v := <-cluster.msgBuffer:
			cluster.handler(v)
		case v := <-cluster.closeChan:
			if v {
				close(cluster.msgBuffer)
				close(cluster.closeChan)
				return
			}
		case <-time.After(time.Second * time.Duration(DefaultWebSocket.config.SessionExpires)):
			cluster.Close()

		}
	}
	glog.Debug("chan is close")

}*/
/*func (cluster *Cluster) GetMessageBufferLen() int {
	return len(cluster.msgBuffer)
}*/
/*func (cluster *Cluster) PushBuffer(messagePackage MessagePackage) {
	if cluster.IsClose(){
		return
	}
	if len(cluster.msgBuffer) < cap(cluster.msgBuffer) {
		cluster.msgBuffer <- messagePackage
	} else {
		err := cluster.SendSelfPackage(entity.GetAS(entity.LogicBusy))
		glog.Error(err)
	}
}*/
func (cluster *Cluster) PushBuffer(messagePackage MessagePackage) {
	if cluster.IsClose() {
		return
	}
	cluster.handler(messagePackage)
	/*if len(cluster.msgBuffer) < cap(cluster.msgBuffer) {
		cluster.msgBuffer <- messagePackage
	} else {
		err := cluster.SendSelfPackage(entity.GetAS(entity.LogicBusy))
		glog.Error(err)
	}*/
}
func (cluster *Cluster) SetClose(isClose bool) {
	cluster.Lock()
	defer cluster.Unlock()
	if cluster._IsClose == false {
		//cluster.closeChan <- true
	}
	cluster._IsClose = isClose
}
func (cluster *Cluster) IsClose() bool {
	cluster.RLock()
	defer cluster.RUnlock()
	return cluster._IsClose
}

func (cluster *Cluster) Close() {
	tc := time.Now().UnixNano()
	glog.Debug("gateway 已经关闭了连接", cluster)
	//glog.Debug("gateway 已经关闭了连接", cluster)
	//cluster.Lock()
	if cluster.IsClose() == false {
		cluster.SetClose(true)
		if cluster.ClientConn != nil {
			err := cluster.ClientConn.Close()
			glog.Error(err)
		}
		//cluster.closeChan <- true
		if cluster.UserID != 0 {
			DefaultWebSocket.ILoginBridge.Offline(cluster.UserID, cluster)
			cluster.Session.Persistence(db.Connector.MySQL.GetMySQL())
			db.Connector.Redis.Persistence(db.Connector.MySQL.GetMySQL(), cluster.UserID)
			DefaultWebSocket.GlobalUser.Del(cluster.UserID)
		}

		tc = time.Now().UnixNano() - tc
		glog.Trace(cluster.UserID, "CloseClient", "用时", float64(tc)/float64(1000)/float64(1000), "ms")
	}
}

func (cluster *Cluster) SendSelfPackages(Data []proto.Message) error {
	for index := range Data {
		smsg := Data[index].(proto.Message)
		if reflect.ValueOf(smsg).IsNil() == false {
			err := cluster.SendSelfPackage(smsg)
			if glog.Error(err) {
				return err
			}
		}
	}
	return nil
}
func (cluster *Cluster) SendSelfPackage(message proto.Message) error {

	//return service.SendBytes(sendtime, message, websocketUser.UserID, websocketUser.TCPConn)
	Package := MakePackage(DefaultWebSocket.protoMessage.GetMessageType(message), message)
	if len(Package) <= 0 {
		//log.Println("发送了一个空的信息")
		return nil
	}

	err := cluster.SendSelfBytesPackage(Package)
	if err != nil {
		return err
	}

	if DefaultWebSocket.config.Debug == false {
		glog.Trace(map[string]interface{}{"MessageData": proto.CompactTextString(message), "UserID": cluster.UserID, "Type": "ExecuteMessage", "MessageName": proto.MessageName(message)})
	}
	return nil
}
func (cluster *Cluster) SendSelfBytesPackage(Package []byte) error {
	/*packBuffer := bytes.NewBuffer([]byte{})
	binary.Write(packBuffer, binary.BigEndian, &cluster.UserID)
	binary.Write(packBuffer, binary.BigEndian, &sendtime)
	binary.Write(packBuffer, binary.BigEndian, Package)*/

	cluster.OutPackageNum++
	cluster.OutBytes += uint64(len(Package))

	glog.Trace("write len", len(Package))

	err := cluster.ClientConn.WriteMessage(websocket.BinaryMessage, Package)
	//err := websocket.Message.Send(cluster.ClientConn, Package)
	return err

}

/*func (session *Session) GetCache() *dao.SessionCache {
	//Cache *dao.SessionCache //

}*/

/*func (session *Session) GetProductLines() []*dao.ProductData {
	return session._ProductLines
}*/

/*func (session *Session) GetStealFunctionItemEffect() *dao.FunctionItemEffect {
	return session._StealFunctionItemEffect
}*/

//set
/*func (session *Session) SetProductLines(data []*dao.ProductData) {
	session._ProductLines = data
}*/

/*func (session *Session) SetStealFunctionItemEffect(data *dao.FunctionItemEffect) {
	session._StealFunctionItemEffect = data
}*/
