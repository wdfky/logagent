package etcd

import (
	"context"
	"encoding/json"
	"time"

	"code.kawai.com/wdfky/logagent/common"
	"code.kawai.com/wdfky/logagent/taillog"
	"github.com/sirupsen/logrus"
	clientv3 "go.etcd.io/etcd/client/v3"
)

var (
	cli *clientv3.Client
)

func Init(addr string) (err error) {
	cli, err = clientv3.New(clientv3.Config{
		Endpoints:   []string{addr},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		logrus.Error("connect to ectd failed, err:", err)
		return
	}
	return
}
func GetConf(key string) (collectEntryList []common.CollectEntry, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()
	resp, err := cli.Get(ctx, key)
	if err != nil {
		logrus.Error("get from etcd failed, err:", err)
		return
	}
	if len(resp.Kvs) == 0 {
		logrus.Warningf("get len:0 from etcd by key:%s", key)
	}
	ret := resp.Kvs[0]
	//ret.Value json格式字符串
	err = json.Unmarshal(ret.Value, &collectEntryList)
	if err != nil {
		logrus.Errorf("json unmarshal failed, err: %v", err)
		return
	}
	return
}

//监控etcd中配置变化
func WatchConf(key string) {

	//监视
	for {
		wc := cli.Watch(context.Background(), key)
		logrus.Info(key)
		//尝试从通道取值
		var newConf []common.CollectEntry
		for wresp := range wc {
			logrus.Info("get new conf from etcd")
			for _, evt := range wresp.Events {
				if evt.Type == clientv3.EventTypeDelete {
					// 如果是删除
					logrus.Warning("etcd delete the key!!!")
					taillog.GetNewConf(newConf) // 没有任何接收就是阻塞的
					continue
				}
				err := json.Unmarshal(evt.Kv.Value, &newConf)
				if err != nil {
					logrus.Errorf("json unmarshal new conf failed, err:%v", err)
					continue
				}
				logrus.Info("新配置来了")
				taillog.GetNewConf(newConf)
			}
		}
	}

}
