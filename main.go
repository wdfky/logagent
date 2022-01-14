package main

import (
	"fmt"

	"code.kawai.com/wdfky/logagent/common"
	"code.kawai.com/wdfky/logagent/etcd"
	"code.kawai.com/wdfky/logagent/kafka"
	"code.kawai.com/wdfky/logagent/taillog"
	"github.com/go-ini/ini"
	"github.com/sirupsen/logrus"
)

type Config struct {
	KafkaConf   `ini:"kafka"`
	EtcdConf    `ini:"etcd"`
	CollectConf `ini:"collect"`
}
type KafkaConf struct {
	Address  string `ini:"address"`
	Chansize int    `ini:"chan_size"`
}
type EtcdConf struct {
	Address    string `ini:"address"`
	Collectkey string `ini:"collect_key"`
}
type CollectConf struct {
	LogFilePath string `ini:"logfile_path"`
}

var (
	cfg = new(Config)
)

func run() {
	select {}
}

//读取日志发送给kafka
func main() {
	ip, err := common.GetOutboundIP()
	if err != nil {
		logrus.Errorf("get ip failed, err: %v", err)
		return
	}
	//0.初始化配置
	err = ini.MapTo(cfg, "./conf/config.ini")
	if err != nil {
		logrus.Error("config load failed, err:", err)
		return
	}
	//1.初始化kafka链接
	err = kafka.Init([]string{cfg.KafkaConf.Address}, cfg.KafkaConf.Chansize)
	if err != nil {
		logrus.Error("init kafka failed, err:", err)
		return
	}
	logrus.Infoln("init to kafka sucess")
	err = etcd.Init(cfg.EtcdConf.Address)
	if err != nil {
		logrus.Error("init etcd failed, err:", err)
	}
	logrus.Infoln("init to etcd sucess")
	collectKey := fmt.Sprintf(cfg.EtcdConf.Collectkey, ip)
	allConf, err := etcd.GetConf(collectKey)
	if err != nil {
		logrus.Error("get conf failed , err:", err)
	}
	//2.打开日志文件准备输出日志
	go etcd.WatchConf(collectKey)
	err = taillog.Init(allConf)
	if err != nil {
		logrus.Errorf("open log file failed, err:", err)
		return
	}
	logrus.Info("init tailfile success")
	run()
}
