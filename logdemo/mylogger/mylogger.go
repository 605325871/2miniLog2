package mylogger

import (
	"fmt"
	"path"
	"runtime"
	"strings"
)

func getInfo(n int) (funcName, fileName string, linnum int) {
	pc, file, linnum, ok := runtime.Caller(n)
	if !ok {
		fmt.Println("runtime.Caller() failed")
		return
	}
	funcName = runtime.FuncForPC(pc).Name()
	fileName = path.Base(file)
	funcName = strings.Split(funcName, ".")[1]
	return funcName, fileName, linnum
}
