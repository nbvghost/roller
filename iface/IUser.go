package iface

type IUser interface {
	GetUserID() uint64
	GetWorldKey() string
	GetAreaID() int64
}
