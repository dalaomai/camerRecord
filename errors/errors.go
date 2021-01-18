// 错误封装

package errors

import "fmt"

// CommonError 通用报错
type CommonError struct {
	Msg  string
	Code string
}

func (r *CommonError) Error() string {
	return fmt.Sprintf("谷歌云盘报错：\nmsg:%v \ncode%v", r.Msg, r.Code)
}

// NewError 新建错误
func NewError(msg string) *CommonError {
	return &CommonError{Msg: msg, Code: "0"}
}
