package logger

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"kelei.com/utils/check"
	"kelei.com/utils/frame/command"
)

const (
	defaultLevel     = 5
	defaultCycleTime = 60 * 60 * 60 //日志文件切分周期（秒）
	bCompress        = false        //是否压缩文件
	keepDays         = 3            //日志保存的天数
)

const (
	PanicLevel int = iota
	FatalLevel
	ErrorLevel
	WarnLevel
	InfoLevel
	DebugLevel
)

type LogFile struct {
	level     int `5`
	logTime   int64
	fileName  string
	fileFd    *os.File
	cycleTime int64
	logUser   string
	logShow   bool
}

func (l LogFile) Write(buf []byte) (n int, err error) {
	if l.fileName == "" {
		fmt.Printf("consol: %s", buf)
		return len(buf), nil
	}

	if logFile.logTime+l.cycleTime < time.Now().Unix() {
		createLogFile()
		logFile.logTime = time.Now().Unix()
	}

	if logFile.fileFd == nil {
		return len(buf), nil
	}

	return logFile.fileFd.Write(buf)
}

var logFile LogFile

//新建日志
func Config(logFolder string, level int) bool {
	logFile.fileName = logFolder
	logFile.level = level
	logFile.logShow = true
	logFile.logUser = ""
	SetCycleTime(defaultCycleTime)
	log.SetOutput(logFile)
	log.SetFlags(log.Lmicroseconds | log.Lshortfile)
	return true
}

//设置日志等级
func setLevel(level int) {
	logFile.level = level
}

//获取日志等级
func getLevel() int {
	return logFile.level
}

//获取是否打印日志
func getLogShow() bool {
	return logFile.logShow
}

//设置是否打印日志
func setLogShow(show bool) {
	logFile.logShow = show
}

//打印日志
func showLog(text *string) {
	if logFile.logShow {
		fmt.Fprintln(os.Stdout, *text)
	}
}

//设置日志文件的循环周期
func SetCycleTime(cycleTime int64) {
	logFile.cycleTime = cycleTime
}

func Debugf(format string, args ...interface{}) {
	if logFile.level >= DebugLevel {
		log.SetPrefix("debug ")
		text := format
		if len(args) > 0 {
			text = fmt.Sprintf(format, args...)
		}
		log.Output(2, text)
		showLog(&text)
	}
}

func Infof(format string, args ...interface{}) {
	if logFile.level >= InfoLevel {
		log.SetPrefix("info  ")
		text := format
		if len(args) > 0 {
			text = fmt.Sprintf(format, args...)
		}
		log.Output(2, text)
		showLog(&text)
	}
}

func Warnf(format string, args ...interface{}) {
	if logFile.level >= WarnLevel {
		log.SetPrefix("warn  ")
		text := format
		if len(args) > 0 {
			text = fmt.Sprintf(format, args...)
		}
		log.Output(2, text)
		showLog(&text)
	}
}

func Errorf(format string, args ...interface{}) {
	if logFile.level >= ErrorLevel {
		log.SetPrefix("error ")
		text := format
		if len(args) > 0 {
			text = fmt.Sprintf(format, args...)
		}
		log.Output(2, text)
		showLog(&text)
	}
}

func Fatalf(format string, args ...interface{}) {
	if logFile.level >= FatalLevel {
		log.SetPrefix("fatal ")
		text := format
		if len(args) > 0 {
			text = fmt.Sprintf(format, args...)
		}
		log.Output(2, text)
		panic(text)
	}
}

func CheckError(error error, args ...string) {
	if error != nil {
		info := ""
		if len(args) > 0 {
			info = args[0]
		}
		Errorf(" %s %s", info, error.Error())
	}
}

func CheckFatal(error error, args ...string) {
	if error != nil {
		info := ""
		if len(args) > 0 {
			info = args[0]
		}
		Fatalf(" %s %s", info, error.Error())
	}
}

func createLogFile() bool {
	logdir := "./"
	if index := strings.LastIndex(logFile.fileName, "/"); index != -1 {
		logdir = logFile.fileName[0:index] + "/"
		os.MkdirAll(logFile.fileName[0:index], os.ModePerm)
	}
	now := time.Now()
	filename := fmt.Sprintf("%04d%02d%02d%02d%02d", now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute())
	filename = strings.Replace(logFile.fileName, ".log", filename+".log", 1)
	if err := os.Rename(logFile.fileName, filename); err == nil {
		go func() {
			//压缩文件
			if bCompress {
				tarCmd := exec.Command("tar", "-zcf", filename+".tar.gz", filename, "--remove-files")
				tarCmd.Run()
				rmCmd := exec.Command("rm", "-rf", filename)
				rmCmd.Run()
			}
			//最多保留n天的日志
			rmCmd := exec.Command("/bin/sh", "-c", "find "+logdir+` -mtime +`+strconv.Itoa(keepDays)+` -exec rm {} \;`)
			rmCmd.Run()
		}()
	}
	for index := 0; index < 10; index++ {
		if fd, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666); nil == err {
			logFile.fileFd.Sync()
			logFile.fileFd.Close()
			logFile.fileFd = fd
			break
		}
		logFile.fileFd = nil
	}
	return true
}

var commands = map[string]string{"loglevel": "5debug 4info 3error 2Fatalf", "logshow": "是否将日志打印到桌面"}

func CommandBlock() func(comm string) bool {
	return func(comm string) bool {
		success := true
		commName, commVal := command.ParseCommand(comm)
		switch commName {
		case "hello":
			Infof("world")
		case "loglevel":
			loglevel(commVal)
		case "logshow":
			logshow(commVal)
		case "hh":
			hh()
		default:
			success = false
		}
		return success
	}
}

func hh() {
	buff := bytes.Buffer{}
	for name, des := range commands {
		buff.WriteString(fmt.Sprintf("     %s : %s\n", name, des))
	}
	info := buff.String()
	Infof("{\n%s}", info)
}

func loglevel(commVal string) {
	if commVal == "" {
		Infof("日志等级 : %d", getLevel())
	} else {
		if lvl, err := check.CheckNum(commVal); err == nil {
			setLevel(lvl)
		} else {
			Warnf("参数错误")
		}
	}
}

func logshow(commVal string) {
	if commVal == "" {
		Infof("日志打印到桌面 : %t", getLogShow())
	} else {
		if show, err := check.CheckBool(commVal); err == nil {
			setLogShow(show)
		} else {
			Warnf("参数错误")
		}
	}
}
