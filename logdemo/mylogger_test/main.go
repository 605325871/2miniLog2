package main

import (
	"logdemo/mylogger"
	"time"
)

func main() {
	log := mylogger.NewFailLogger("info", "./", "logdem.log", 10*1024)
	id := 100
	name := "四阿"
	for {

		log.Debug("这是一条debug日志")
		log.Error("这是一条error日志,id:%d,name:%s", id, name)
		log.Info("这是一条info日志")
		log.Fatal("这是一条fatal日志")
		time.Sleep(time.Second)
	}

}
