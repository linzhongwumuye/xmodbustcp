package main

import (
	"flag"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"runtime/debug"
	"strconv"
	"syscall"
	"xlib/log"
	"xmodbustcp/business"
	"xmodbustcp/define"
)

var (
	confpath = flag.String("confpath", "./etc/conf.json", "--svrConf=path")
	APP_NAME = ""
	APP_Version = ""
	Build_Data = ""
)

func Version() {
	log.Infof("APP_NAME:%-20t%s%t",APP_NAME)
	log.Infof("APP_Version:%-20t%s%t",APP_Version)
	log.Infof("Build_Data:%-20t%s%t",Build_Data)
}


func main() {
	flag.Parse()
	defer func() {
		Version()
		if err := recover(); err != nil {
			log.Fatal(""+
				"||||||||||||||||||||\n"+
				"||||||||||||||||||||\n"+
				"||||||||||||||||||||\n"+
				"||||||||||||||||||||\n"+
				"||||||||||||||||||||\n", err,
				"\r\n"+string(debug.Stack())+"\r\n")
		}
		os.Exit(1)
	}()

	var svrconf define.SvrConfInterface
	svrconf = new(define.Svrconf)
	var absSvrConfFile string
	if filepath.IsAbs(*confpath) {
		absSvrConfFile = *confpath
	} else {
		dir, _ := os.Getwd()
		absSvrConfFile = dir + string(filepath.Separator) + *confpath
	}
	if err := define.ReadSvrConf(absSvrConfFile, svrconf); err != nil {
		return
	}

	// 配置日志
	if err := log.SetLogger(svrconf.GetLogRollType(), svrconf.GetLogDir(), svrconf.GetLogFile(), svrconf.GetLogCount(), svrconf.GetLogSize(), svrconf.GetLogUnit(), svrconf.GetLogLevel(), svrconf.GetLogCompress()); err != nil {
		return
	}
	go log.HandleSignalChangeLogLevel()

	// 更新pid文件
	if pid := os.Getpid(); pid != 1 {
		if err := ioutil.WriteFile(svrconf.GetPid(), []byte(strconv.Itoa(pid)), 0777); err != nil {
			log.Error("Create pid file", svrconf.GetPid(), err.Error())
			return
		} else {
			log.Info("Create pid file", svrconf.GetPid(), "success")
		}
	}
	defer os.Remove(svrconf.GetPid())


	//监听signal，使用Signal停掉服务
	signalChan := make(chan os.Signal)
	signal.Ignore(syscall.SIGPIPE, syscall.SIGALRM)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		signalRecv := <-signalChan
		log.Info("Recieve signal", signalRecv.String())
		business.StopXSvrer()
	}()

	if err := business.StartXSvrer(svrconf); err != nil {
		log.Error("Program Abort", err)
	}
}
