/**
 * @Author: 
 * @Description:
 * @File:  miaerror
 * @Version: 1.0.0
 * @Date: 2020/11/20 01:38
 */
package MiaError

import (
	"MiaGame/Library/MiaLog"
	"errors"
)

func CheckError(err error) bool {
	if err!=nil {
		MiaLog.CError("error is :",err.Error())
		return true
	}
	return false
}
func NewError(err string) error{
	return errors.New(err)
}

