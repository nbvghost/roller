package entity

import (
	"fmt"
	"github.com/golang/protobuf/proto"

	"github.com/nbvghost/roller/mold"
)

type Action struct {
	Action  string
	Message string
}

var SessionExpires = &Action{Action: "SessionExpires", Message: "session过期,请刷新再试"}                            //
var SendWait = &Action{Action: "SendWait", Message: "处理中，请稍候"}                                                //
var ReceiveWait = &Action{Action: "ReceiveWait", Message: "处理中，请稍候"}                                          //
var LoginFailedInvalidToken = &Action{Action: "LoginFailedInvalidToken", Message: "登陆失败，token无效"}             //
var AccountLoggedDifferentLocation = &Action{Action: "AccountLoggedDifferentLocation", Message: "您的账号在异地登陆了"} //

//var LogicBusy = mold.KV{K: "LogicBusy", V: "正在处理，请稍候"}
var UnknownType = mold.KV{K: "UnknownType", V: "未知类型:%v"}   //未知类型:%v
var UnknownError = mold.KV{K: "UnknownError", V: "未知错误:%v"} //未知错误
var Error = mold.KV{K: "Error", V: "错误:%v"}                 //错误
var Empty = mold.KV{K: "Empty", V: "Empty:%v"}              //错误
var SQLError = mold.KV{K: "SQLError", V: "SQLError:%v"}     //错误
var InsufficientItem = mold.KV{K: "InsufficientItem", V: "%v不足"}

var Notice = mold.KV{K: "Notice", V: "%v"}             //检测到新功能，我们对目前版本进行离线更新，界时请做好离线准备
var ForceOffline = mold.KV{K: "ForceOffline", V: "%v"} //离线更新，服务器准备断开
var LoginLimit = mold.KV{K: "LoginLimit", V: "%v"}     //限制登陆
//var NoticeOfflineLoginMessage = mold.KV{K: "NoticeOfflineLoginMessage", V: "%v"}

/*type KV struct {
	K string
	V string
}*/

func GetKV(kv mold.KV, args ...interface{}) *mold.KV {
	as := &mold.KV{}
	as.V = fmt.Sprintf(kv.V, args...)
	as.K = kv.K
	return as
}

type MessageAction struct {
	Error error
	Data  proto.Message
}

/*type ActionStatus struct {
	Code int32
	Message string
	Data    interface{}
}

func (as *ActionStatus) SmartSuccessData(data interface{}) *ActionStatus {
	as.Message = "SUCCESS"
	as.Code = 0
	as.Data = data
	return as
}
func (as *ActionStatus) SmartError(err error, successTxt string, data interface{}) *ActionStatus {

	if err == nil {
		as.Message = successTxt
		as.Code = 0
		as.Data = data
	} else {
		as.Message = err.Error()
		as.Code = -1
		as.Data = data
	}
	return as
}
func (as *ActionStatus) Smart(success int32, s string, f string) *ActionStatus {
	as.Code = success
	if success==0 {
		as.Message = s
	} else {
		as.Message = f
	}
	return as
}
func (as *ActionStatus) SmartData(success int32, s string, f string, data interface{}) *ActionStatus {
	as.Code = success
	if success==0 {
		as.Message = s
		as.Data = data
	} else {
		as.Message = f
	}
	return as
}*/

