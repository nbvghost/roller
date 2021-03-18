package iface

type IRole interface {
    GetRoleID() uint64
    GetWorldKey() string
    GetAreaID() int64
}
