# 简单的日志库

>实现日志的开关，分级别输出,选择输出到终端或者文件，实现日志文件大小的切割

## 输出到终端 

### 定义日志的级别
```go

type loglevel uint64

const (
	UNKNOWE loglevel = iota
	DEBUG
	TRACE
	INFO
	WARNING
	ERROR
	FATAL
)

```

>定义日志的级别

### 向终端输出的结构体
```go
type ConsoleLogger struct {
	level loglevel
}

```

### 统一解析传来的日志级别定义
```go
func parseLogevel(s string) (loglevel, error) {
	s = strings.ToLower(s)
	switch s {
	case "debug":
		return DEBUG, nil
	case "tarce":
		return TRACE, nil
	case "info":
		return INFO, nil
	case "warning":
		return WARNING, nil
	case "error":
		return ERROR, nil
	case "fatal":
		return FATAL, nil
	default:
		err := errors.New("日志级别错误")
		return UNKNOWE, err
	}
}
```

### 对外的接口返回所构造的结构体
```go
func Newlog(logstr string) ConsoleLogger {
	loevel, err := parseLogevel(logstr)
	if err != nil {
		panic(err)
	}

	return ConsoleLogger{
		level: loevel}
}
```

### 控制日志输出级别
```go
func (c ConsoleLogger) enable(level loglevel) bool {
	return c.level <= level
}
```
>如果日志本身定义输出级别为info,则debug级别的详细信息不会输出

### 日志格式化输出封装
```go
func (c ConsoleLogger) log(lv loglevel, format string, a ...interface{}) {
	if c.enable(lv) { //如果所有输出的级别大于等于日志结构体本身级别才输出，级别越大越重要
		msg := fmt.Sprintf(format, a...) //...利用可变参数可以将一些信息传入，用sprintf将其格式化为一个字符串
		funcName, fileName, lineno := getInfo(3)// runtime.caller的函数可以根据函数栈zhen调用得到信息
		now := time.Now()
		fmt.Printf("[%s][%s][%s:%s:%d]%s\n", now.Format("2006-01-02 03:04:05"), getlogstring(lv), funcName, fileName, lineno, msg)
	}

}

```
### 不同级别的输出
```go
func (c ConsoleLogger) Debug(msg string, arg ...interface{}) {
	c.log(DEBUG, msg, arg...)
}
func (c ConsoleLogger) Info(msg string, arg ...interface{}) {
	c.log(INFO, msg, arg...)
}
func (c ConsoleLogger) Waring(msg string, arg ...interface{}) {
	c.log(WARNING, msg, arg...)
}
func (c ConsoleLogger) Error(msg string, arg ...interface{}) {
	c.log(ERROR, msg, arg...)
}
func (c ConsoleLogger) Fatal(msg string, arg ...interface{}) {
	c.log(FATAL, msg, arg...)
}

```
### 得到运行时信息
```go
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

```




## 输出到文件中

### 定义文件结构体
```go
type FileLogger struct {
	level      loglevel
	filpath    string //文件路径
	filName    string//文件名字
	filobj     *os.File
	errFileobj *os.File
	maxFile    int64
}

```
>将日志输出到文件中，对于错误级别以上的日志特别输出到一个文件中

### 构造新的文件日志结构体
```go
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
	}
	f1.initFile()
	if err != nil {
		panic(err)
	}
	return f1
}

```

### 初始化打开文件对象
```go
func (f *FileLogger) initFile() error {
	fullFileNmae := path.Join(f.filpath, f.filName)
	filobj, err := os.OpenFile(fullFileNmae, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("open log fill err ", err)
		return err
	}
	errfilobj, err := os.OpenFile(fullFileNmae+".err", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("open log fill err ", err)
		return err
	}
	f.filobj = filobj
	f.errFileobj = errfilobj
	return nil
}
```

### 分割文件大小
```go
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
	fileobj.Close()
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
```

### 格式化输出日志
```go
func (f *FileLogger) log(lv loglevel, format string, a ...interface{}) {

	msg := fmt.Sprintf(format, a...)
	funcName, fileName, lineno := getInfo(3)
	now := time.Now()
	if f.checkSize(f.filobj) {
		newfile, err := f.spilFile(f.filobj)
		if err != nil {
			fmt.Println("f. spilFlie err", err)
			return
		}
		f.filobj = newfile
	}

	fmt.Fprintf(f.filobj, "[%s][%s][%s:%s:%d]%s\n", now.Format("2006-01-02 03:04:05"), getlogstring(lv), funcName, fileName, lineno, msg)
	if lv >= ERROR {
		if f.checkSize(f.errFileobj) {
			newfile, err := f.spilFile(f.errFileobj)
			if err != nil {
				fmt.Println("f. spilFlie err", err)
				return
			}
			f.errFileobj = newfile
		}
		fmt.Fprintf(f.errFileobj, "[%s][%s][%s:%s:%d]%s\n", now.Format("2006-01-02 03:04:05"), getlogstring(lv), funcName, fileName, lineno, msg)
	}

}
```


### 检查日志级别，与文件大小
```go
func (f *FileLogger) enable(level loglevel) bool {
	return f.level <= level
}
func (f *FileLogger) checkSize(fil *os.File) bool {
	filInfo, err := fil.Stat()
	if err != nil {
		fmt.Printf("get file failed ,err:%v", err)
		return false
	}
	return filInfo.Size() >= f.maxFile
}
```
### 日志输出
```go
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
```