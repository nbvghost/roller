package socket

import (
	"bytes"
	"encoding/binary"
	"github.com/golang/protobuf/proto"
	"github.com/nbvghost/glog"
	"sync"
)

var _lock sync.RWMutex
var _msgID int32

func _getMsgID() int32 {
	_lock.Lock()
	_msgID++
	_lock.Unlock()
	//now, _ := strconv.ParseUint(time.Now().Format("20060102150405"), 10, 64)
	//msgID := int32(now)
	return _msgID
}

/*func GetPackageType(message proto.Message) (Package []byte) {
	messagePackageName := proto.MessageName(message)
	messageName := strings.Split(messagePackageName, ".")[1]
}*/
func MakePackage(packageType int32, message proto.Message) (Package []byte) {
	//msgID := int32(1000 + _msgID)
	//msgID++
	//messagePackageName := proto.MessageName(message)
	//messageName := strings.Split(messagePackageName, ".")[1] //结构名字
	//packageType := pb.MessageName_MsgType_value[messageName]

	msgBuffer := bytes.NewBuffer([]byte{})
	b, err := proto.Marshal(message)
	if glog.Error(err) {
		return []byte{}
	}
	msgBuffer.Write(b)

	bodyBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bodyBuffer, binary.BigEndian, &packageType)
	binary.Write(bodyBuffer, binary.BigEndian, msgBuffer.Bytes())

	body := bodyBuffer.Bytes()

	bodyLen := int32(len(body))

	packBuffer := bytes.NewBuffer([]byte{})

	msgID := _getMsgID()
	binary.Write(packBuffer, binary.BigEndian, &msgID)
	binary.Write(packBuffer, binary.BigEndian, &bodyLen)
	binary.Write(packBuffer, binary.BigEndian, body)

	return packBuffer.Bytes()
}
