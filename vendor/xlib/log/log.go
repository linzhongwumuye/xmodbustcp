package log

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

type level int32

//日志级别，从低到高
const (
	all_level level = iota
	debug_level
	info_level
	warn_level
	error_level
	fatal_level
	off_level
)

type unit int64

//文件大小的单位
const (
	_            = iota
	kb_unit unit = 1 << (iota * 10)
	mb_unit
	gb_unit
	tb_unit
)

//全局变量
var logLevel level = 1       //日志级别
var maxFileSize int64        //日志文件大小
var maxFileCount int32       //日志文件数量
var dailyRolling bool = true //是否按日期滚动
var rollingFile bool = false //是否按文件大小滚动
var logObj *_FILE            //文件结构对象
var compressType int64

const dateformat = "2006-01-02" //日期格式

const (
	CompressTypeNo int64 = iota
	CompressTypeGzip
)

//文件结构
type _FILE struct {
	dir      string        //目录
	filename string        //文件名
	_suffix  int           //文件名后缀
	_date    *time.Time    //当前日期
	mu       *sync.RWMutex //对logObj锁控制
	logfile  *os.File      //文件描述符
	lg       *log.Logger   //标准库log对象
}

type Logger struct{}

func DefaultLogger() *Logger {
	return &Logger{}
}

func (l *Logger) Printf(format string, v ...interface{}) {
	Infof(format, v...)
}

//配置日志
func SetLogger(RollType, Dir, File string, Count int32, Size int64, Unit, Level string, compressT int64) error {
	compressType = compressT
	switch RollType {
	case "RollingFile":
		setLogRollingFile(Dir, File, Count, Size, getUnit(Unit))
		setLogLevel(getLevel(Level))
	case "RollingDaily":
		setLogRollingDaily(Dir, File)
		setLogLevel(getLevel(Level))
	default:
		return errors.New("Wrong LogRollType " + RollType)
	}
	return nil
}

func SetLogLevelAll() {
	setLogLevel(all_level)
}

func SetLogLevelError() {
	setLogLevel(error_level)
}

//根据单位字符串返回单位
func getUnit(Unit string) unit {
	switch Unit {
	case "KB":
		return kb_unit
	case "MB":
		return mb_unit
	case "GB":
		return gb_unit
	case "TB":
		return tb_unit
	default:
		return 0
	}
}

//根据日志级别字符串返回级别
func getLevel(Level string) level {
	switch Level {
	case "all":
		return all_level
	case "debug":
		return debug_level
	case "info":
		return info_level
	case "warn":
		return warn_level
	case "error":
		return error_level
	case "fatal":
		return fatal_level
	case "off":
		return off_level
	default:
		return 0
	}
}

//设置日志级别
func setLogLevel(_level level) {
	logLevel = _level
}

//设置日志文件按文件大小方式滚动
//第一个参数为日志文件存放目录
//第二个参数为日志文件命名
//第三个参数为备份文件最大数量
//第四个参数为备份文件大小
//第五个参数为文件大小的单位KB|MB|GB|TB
func setLogRollingFile(fileDir, fileName string, maxNumber int32, maxSize int64, _unit unit) {
	maxFileCount = maxNumber
	maxFileSize = maxSize * int64(_unit)
	rollingFile = true
	dailyRolling = false
	mkdirlog(fileDir)
	logObj = &_FILE{dir: fileDir, filename: fileName, mu: new(sync.RWMutex)}
	logObj.mu.Lock()
	defer logObj.mu.Unlock()
	for i := 1; i <= int(maxNumber); i++ {
		if isExist(fileDir + string(filepath.Separator) + fileName + "." + strconv.Itoa(i)) {
			logObj._suffix = i
		} else {
			break
		}
	}
	if !logObj.isMustRename() {
		logObj.logfile, _ = os.OpenFile(fileDir+string(filepath.Separator)+fileName, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
		logObj.lg = log.New(logObj.logfile, "", log.Ldate|log.Ltime|log.Lshortfile)
	} else {
		logObj.rename()
	}
	go fileMonitor()
}

//设置日志文件备按日期滚动
//第一个参数为日志文件存放目录
//第二个参数为日志文件命名
func setLogRollingDaily(fileDir, fileName string) {
	rollingFile = false
	dailyRolling = true
	t, _ := time.Parse(dateformat, time.Now().Format(dateformat))
	mkdirlog(fileDir)
	logObj = &_FILE{dir: fileDir, filename: fileName, _date: &t, mu: new(sync.RWMutex)}
	logObj.mu.Lock()
	defer logObj.mu.Unlock()
	if !logObj.isMustRename() {
		logObj.logfile, _ = os.OpenFile(fileDir+string(filepath.Separator)+fileName, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
		logObj.lg = log.New(logObj.logfile, "", log.Ldate|log.Ltime|log.Lshortfile)
	} else {
		logObj.rename()
	}
}

//创建目录
func mkdirlog(dir string) (e error) {
	_, er := os.Stat(dir)
	b := er == nil || os.IsExist(er)
	if !b {
		if err := os.MkdirAll(dir, 0666); err != nil {
			if os.IsPermission(err) {
				fmt.Println("create dir error:", err.Error())
				e = err
			}
		}
	}
	return
}

//捕获异常
func catchError() {
	if err := recover(); err != nil {
		log.Println("err", err)
	}
}

func Debug(v ...interface{}) {
	if dailyRolling {
		fileCheck()
	}
	defer catchError()
	if logObj != nil {
		logObj.mu.RLock()
		defer logObj.mu.RUnlock()
	}

	if logLevel <= debug_level {
		if logObj != nil {
			logObj.lg.Output(2, fmt.Sprintln(color(levelColor[debug_level], "[DEBUG]"), v))
		}
	}
}

func Info(v ...interface{}) {
	if dailyRolling {
		fileCheck()
	}
	defer catchError()
	if logObj != nil {
		logObj.mu.RLock()
		defer logObj.mu.RUnlock()
	}
	if logLevel <= info_level {
		if logObj != nil {
			logObj.lg.Output(2, fmt.Sprintln(color(levelColor[info_level], "[INFO]"), v))
		}
	}
}

func Warn(v ...interface{}) {
	if dailyRolling {
		fileCheck()
	}
	defer catchError()
	if logObj != nil {
		logObj.mu.RLock()
		defer logObj.mu.RUnlock()
	}

	if logLevel <= warn_level {
		if logObj != nil {
			logObj.lg.Output(2, fmt.Sprintln(color(levelColor[warn_level], "[WARN]"), v))
		}
	}
}

func Error(v ...interface{}) {
	if dailyRolling {
		fileCheck()
	}
	defer catchError()
	if logObj != nil {
		logObj.mu.RLock()
		defer logObj.mu.RUnlock()
	}
	if logLevel <= error_level {
		if logObj != nil {
			logObj.lg.Output(2, fmt.Sprintln(color(levelColor[error_level], "[ERROR]"), v))
		}
	}
}

func Fatal(v ...interface{}) {
	if dailyRolling {
		fileCheck()
	}
	defer catchError()
	if logObj != nil {
		logObj.mu.RLock()
		defer logObj.mu.RUnlock()
	}
	if logLevel <= fatal_level {
		if logObj != nil {
			logObj.lg.Output(2, fmt.Sprintln(color(levelColor[fatal_level], "[FATAL]"), v))
		}
	}
}

func Debugf(format string, v ...interface{}) {
	if dailyRolling {
		fileCheck()
	}
	defer catchError()
	if logObj != nil {
		logObj.mu.RLock()
		defer logObj.mu.RUnlock()
	}

	if logLevel <= debug_level {
		if logObj != nil {
			logObj.lg.Output(2, fmt.Sprintln(color(levelColor[debug_level], "[DEBUG]"), fmt.Sprintf(format, v...)))
		}
	}
}

func Infof(format string, v ...interface{}) {
	if dailyRolling {
		fileCheck()
	}
	defer catchError()
	if logObj != nil {
		logObj.mu.RLock()
		defer logObj.mu.RUnlock()
	}

	if logLevel <= info_level {
		if logObj != nil {
			logObj.lg.Output(2, fmt.Sprintln(color(levelColor[info_level], "[INFO]"), fmt.Sprintf(format, v...)))
		}
	}
}

func Warnf(format string, v ...interface{}) {
	if dailyRolling {
		fileCheck()
	}
	defer catchError()
	if logObj != nil {
		logObj.mu.RLock()
		defer logObj.mu.RUnlock()
	}

	if logLevel <= warn_level {
		if logObj != nil {
			logObj.lg.Output(2, fmt.Sprintln(color(levelColor[warn_level], "[WARN]"), fmt.Sprintf(format, v...)))
		}
	}
}

func Errorf(format string, v ...interface{}) {
	if dailyRolling {
		fileCheck()
	}
	defer catchError()
	if logObj != nil {
		logObj.mu.RLock()
		defer logObj.mu.RUnlock()
	}

	if logLevel <= error_level {
		if logObj != nil {
			logObj.lg.Output(2, fmt.Sprintln(color(levelColor[error_level], "[ERROR]"), fmt.Sprintf(format, v...)))
		}
	}
}

func Fatalf(format string, v ...interface{}) {
	if dailyRolling {
		fileCheck()
	}
	defer catchError()
	if logObj != nil {
		logObj.mu.RLock()
		defer logObj.mu.RUnlock()
	}

	if logLevel <= fatal_level {
		if logObj != nil {
			logObj.lg.Output(2, fmt.Sprintln(color(levelColor[fatal_level], "[FATAL]"), fmt.Sprintf(format, v...)))
		}
	}
}

//是否要滚动到下一个文件
func (f *_FILE) isMustRename() bool {
	if dailyRolling {
		t, _ := time.Parse(dateformat, time.Now().Format(dateformat))
		if t.After(*f._date) {
			return true
		}
	} else {
		if maxFileCount > 1 {
			if fileSize(f.dir+string(filepath.Separator)+f.filename) >= maxFileSize {
				return true
			}
		}
	}
	return false
}

//滚动到下一个文件
func (f *_FILE) rename() {
	if dailyRolling {
		fn := f.dir + string(filepath.Separator) + f.filename + "." + f._date.Format(dateformat)
		if !isExist(fn) && f.isMustRename() {
			if f.logfile != nil {
				f.logfile.Close()
			}
			err := os.Rename(f.dir+string(filepath.Separator)+f.filename, fn)
			if err != nil {
				f.lg.Println("rename err", err.Error())
			}
			t, _ := time.Parse(dateformat, time.Now().Format(dateformat))
			f._date = &t
			f.logfile, _ = os.Create(f.dir + string(filepath.Separator) + f.filename)
			f.lg = log.New(logObj.logfile, "\n", log.Ldate|log.Ltime|log.Lshortfile)
		}
	} else {
		f.coverNextOne()
	}
}

//滚动到下一个文件的后缀
func (f *_FILE) nextSuffix() int {
	return int(f._suffix%int(maxFileCount) + 1)
}

func (f *_FILE) genFileNameBySuffix(suffix int) string {
	if suffix == 0 {
		return fmt.Sprintf("%s%s%s", f.dir, string(filepath.Separator), f.filename)
	}
	return fmt.Sprintf("%s%s%s.%d", f.dir, string(filepath.Separator), f.filename, suffix)
}

func (f *_FILE) genCompressFileNameBySuffix(suffix int) string {
	fileName := f.genFileNameBySuffix(suffix)
	switch compressType {
	case CompressTypeGzip:
		return fmt.Sprintf("%s.gz", fileName)
	default:
		return fileName
	}
}

//滚动到下一个文件
func (f *_FILE) coverNextOne() {
	f._suffix = f.nextSuffix()
	if f.logfile != nil {
		f.logfile.Close()
	}
	for fileSuffix := int(maxFileCount) - 1; fileSuffix >= 0; fileSuffix-- {
		if isExist(f.genCompressFileNameBySuffix(fileSuffix)) {
			os.Rename(f.genCompressFileNameBySuffix(fileSuffix), f.genCompressFileNameBySuffix(fileSuffix+1))
		}
	}
	if maxFileCount > 1 {
		os.Rename(f.genFileNameBySuffix(0), f.genFileNameBySuffix(1))
		go compressFile(f.genFileNameBySuffix(1))
	}
	f.logfile, _ = os.Create(f.dir + string(filepath.Separator) + f.filename)
	f.lg = log.New(logObj.logfile, "\n", log.Ldate|log.Ltime|log.Lshortfile)
}

func compressFile(fileName string) {
	switch compressType {
	case CompressTypeGzip:
		dstFileName := fmt.Sprintf("%s.gz", fileName)
		bs, err := ioutil.ReadFile(fileName)
		if err != nil {
			Fatalf("压缩文件打开出错,file name: %s,err：%s", fileName, err.Error())
			return
		}
		var b bytes.Buffer
		w := gzip.NewWriter(&b)
		w.Write(bs)
		w.Close()
		err = ioutil.WriteFile(dstFileName, b.Bytes(), 0666)
		if err != nil {
			Fatalf("压缩文件写入出错,file name: %s,err: %s", dstFileName, err.Error())
			return
		}
		os.Remove(fileName)
	}
}

func color(col, s string) string {
	if col == "" {
		return s
	}
	return "\x1b[0;" + col + "m" + s + "\x1b[0m"
}

var levelColor = map[level]string{
	all_level:   "32",
	debug_level: "32",
	info_level:  "33",
	warn_level:  "31",
	error_level: "31",
	fatal_level: "35",
	off_level:   "35",
}

//获取文件大小
func fileSize(file string) int64 {
	f, e := os.Stat(file)
	if e != nil {
		fmt.Println(e.Error())
		return 0
	}
	return f.Size()
}

//检查文件是否存在
func isExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}

//文件监控线程
func fileMonitor() {
	timer := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-timer.C:
			fileCheck()
		}
	}
}

//检查是否需要滚动并做相应滚动
func fileCheck() {
	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
		}
	}()
	if logObj != nil && logObj.isMustRename() {
		logObj.mu.Lock()
		defer logObj.mu.Unlock()
		logObj.rename()
	}
}
