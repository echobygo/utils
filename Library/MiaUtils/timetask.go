package MiaUtils

import (
	"context"
	"fmt"
	"time"
)


/**
 * @Description
 * @Author 岁岁年年码不同
 * @Date 2022/2/24 4:19 下午
 **/
type cbfun  func(Grpccontext context.Context)
type cbFunLogOut func(string)
func TaskCronInterFace(Grpccontext context.Context ,timeOutmillisecond int64 ,taskfun cbfun ,outPutLog cbFunLogOut) {
	defer func() {
		err := recover()
		if err != nil {
			fmt.Println("err=", err)
		}
	}()
	if taskfun == nil {
		return
	}
	var timed = time.Duration(timeOutmillisecond) * time.Millisecond
	tiker := time.NewTicker(timed)
	go func() {
		for {
			select {
			case <-Grpccontext.Done():
				tiker.Stop()
				return
			case <-tiker.C:
				if outPutLog !=nil {
					outPutLog("current Running "+GetFunctionName(taskfun))
				}
				if taskfun!=nil  {
					taskfun(Grpccontext)
				}
				if outPutLog !=nil  {
					outPutLog("current Running "+GetFunctionName(taskfun))
				}
				
			}
		}
	}()
}