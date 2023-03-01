package mylogger

import (
	"fmt"
	"os"
	"path"
	"sync"
	"time"
)

type FileLogger struct {
	level      loglevel
	filpath    string
	filName    string
	filobj     *os.File
	errFileobj *os.File
	maxFile    int64
	logCHan    chan *logMsg
}

type logMsg struct {
	lv        loglevel
	msg       string
	funcName  string
	fileName  string
	timestamp string
	line      int
}

func NewFailLogger(lv, fp, fn string, maxsize int64) *FileLogger {
	logevel, err := parseLogevel(lv)
	if err != nil {
		panic(err)
	}
	f1 := &FileLogger{
		level:   logevel,
		filpath: fp,
		filName: fn,
		maxFile: maxsize,
		logCHan: make(chan *logMsg, 50000),
	}
	f1.initFile()
	if err != nil {
		panic(err)
	}
	return f1
}

func (f *FileLogger) initFile() error {
	fullFileNmae := path.Join(f.filpath, f.filName)
	filobj, err := os.OpenFile(fullFileNmae, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		fmt.Println("open log fill err ", err)
		return err
	}
	errfilobj, err := os.OpenFile(fullFileNmae+".err", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		fmt.Println("open log fill err ", err)
		return err
	}
	//日志文件都打开了
	f.filobj = filobj
	f.errFileobj = errfilobj

	//开启后台的gor去写日志
	for i := 0; i < 5; i++ {
		go f.writebackground()
	}
	return nil
}

var once sync.Once

func (f *FileLogger) spilFile(fileobj *os.File) (*os.File, error) {
	//需要切割日志
	nowStr := time.Now().Format("200601021504000")

	filInfo, err := fileobj.Stat()
	if err != nil {
		fmt.Println(" f.filobj.Stat() err", err)
		return nil, err
	}
	logName := path.Join(f.filpath, filInfo.Name()) //拿到当前日志的完整名字
	newlogName := fmt.Sprintf("%s.bak%s", logName, nowStr)
	//1.关闭当前文件
	once.Do(func() {
		fileobj.Close()
	})

	//2.备份一下rename
	os.Rename(logName, newlogName)
	//3.打开一个新的文件
	Newfilobj, err := os.OpenFile(logName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("open new filelog err", err)
		return nil, err
	}
	//4.将新打开的日志文件对象赋值给filobj
	return Newfilobj, err
}

func (f *FileLogger) writebackground() {

	for {
		if f.checkSize(f.filobj) {
			newfile, err := f.spilFile(f.filobj)
			if err != nil {
				fmt.Println("f. spilFlie err", err)
				return
			}
			f.filobj = newfile
		}

		select {
		case logtmp := <-f.logCHan:
			//把日志拼出来
			logInfo := fmt.Sprintf("[%s][%s][%s:%s:%d]%s\n", logtmp.timestamp, getlogstring(logtmp.lv), logtmp.funcName, logtmp.fileName, logtmp.line, logtmp.msg)

			fmt.Fprintf(f.filobj, "[%s][%s][%s:%s:%d]%s\n", logtmp.timestamp, getlogstring(logtmp.lv), logtmp.funcName, logtmp.fileName, logtmp.line, logtmp.msg)
			if logtmp.lv >= ERROR {
				if f.checkSize(f.errFileobj) {
					newfile, err := f.spilFile(f.errFileobj)
					if err != nil {
						fmt.Println("f. spilFlie err", err)
						return
					}
					f.errFileobj = newfile
				}
				fmt.Fprintf(f.errFileobj, logInfo)
			}
		default:
			//取不到休息500毫秒
			time.Sleep(time.Millisecond * 500)
		}

	}

}
func (f *FileLogger) log(lv loglevel, format string, a ...interface{}) {

	if f.enable(lv) {
		msg := fmt.Sprintf(format, a...)
		funcName, fileName, lineno := getInfo(3)
		now := time.Now()
		//先把日志发送到通道中去
		//1.先造一个发送的结构体
		logTmp := &logMsg{
			lv:        lv,
			msg:       msg,
			funcName:  funcName,
			fileName:  fileName,
			timestamp: now.Format("2006-01-02 03:04:05"),
			line:      lineno,
		}

		select {
		case f.logCHan <- logTmp:
		default:
			//把日志丢掉保证不出现阻塞
		}
	}

}

func (f *FileLogger) enable(level loglevel) bool {
	return f.level <= level
}
func (f *FileLogger) checkSize(fil *os.File) bool {
	filInfo, err := fil.Stat()
	if err != nil {
		fmt.Printf("get file failed ,err:%v\n", err)
		return false
	}
	return filInfo.Size() >= f.maxFile
}

func (f *FileLogger) Debug(msg string, arg ...interface{}) {
	f.log(DEBUG, msg, arg...)
}
func (f *FileLogger) Info(msg string, arg ...interface{}) {
	f.log(INFO, msg, arg...)
}
func (f *FileLogger) Waring(msg string, arg ...interface{}) {
	f.log(WARNING, msg, arg...)
}
func (f *FileLogger) Error(msg string, arg ...interface{}) {
	f.log(ERROR, msg, arg...)
}
func (f *FileLogger) Fatal(msg string, arg ...interface{}) {
	f.log(FATAL, msg, arg...)
}
func (f *FileLogger) Close() {
	f.filobj.Close()
	f.errFileobj.Close()
}
