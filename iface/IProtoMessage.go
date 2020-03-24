package iface

import (
	"github.com/golang/protobuf/proto"
	"reflect"
)

type IProtoMessage interface {
	RegisterHandler(packageType int32, x interface{})
	GetHandler(packageType int32) reflect.Type
	GetMessageName(packageType int32) string
	GetMessageType(message proto.Message) int32
}
