package taillog

import (
	"code.kawai.com/wdfky/logagent/common"
	"github.com/sirupsen/logrus"
)

type tailTaskMgr struct {
	tailTaskMap      map[string]*tailTask
	collectEntryList []common.CollectEntry
	confChan         chan []common.CollectEntry
}

var (
	ttMgr *tailTaskMgr
)

func Init(allConf []common.CollectEntry) (err error) {
	//all里面有几个日志的收集信息
	//针对每一个信息造一个tails
	ttMgr = &tailTaskMgr{
		tailTaskMap:      make(map[string]*tailTask, 20),
		collectEntryList: allConf,
		confChan:         make(chan []common.CollectEntry),
	}
	for _, acf := range allConf {
		tt := newTailTask(acf.Path, acf.Topic)
		err = tt.Init()
		if err != nil {
			logrus.Errorf("creat tail failed path:%v, err:%v", acf.Path, err)
			continue
		}
		logrus.Infof("create a tail task for path:%s success", acf.Path)
		ttMgr.tailTaskMap[tt.path] = tt
		go tt.run()
	}
	//等新配置
	go ttMgr.watch()
	return
}
func (t *tailTaskMgr) watch() {
	for {
		newConf := <-t.confChan //新配置来了
		logrus.Infof("get new conf from etcd, conf:%v, start manage tailTask...", newConf)
		for _, conf := range newConf {
			// 1. 原来已经存在的任务就不用动
			if t.isExist(conf) {
				continue
			}
			// 2. 原来没有的我要新创建一个taiTask任务
			tt := newTailTask(conf.Path, conf.Topic)
			err := tt.Init()
			if err != nil {
				logrus.Errorf("creat tail failed path:%v, err:%v", conf.Path, err)
				continue
			}
			logrus.Infof("create a tail task for path:%s success", conf.Path)
			ttMgr.tailTaskMap[tt.path] = tt
			go tt.run()
		}
		for path, task := range t.tailTaskMap {
			found := false
			for _, conf := range newConf {
				if conf.Path == path {
					found = true
					break
				}
			}
			if !found {
				logrus.Infof("the task collect path:%s need to stop.", task.path)
				delete(t.tailTaskMap, path) // 从管理类中删
				task.cancel()
			}
		}
		// 3. 原来有的现在没有的要tailTask停掉
		// 找出tailTaskMap中存在,但是newConf不存在的那些tailTask,把它们都停掉
	}
}
func (t *tailTaskMgr) isExist(conf common.CollectEntry) bool {
	_, ok := t.tailTaskMap[conf.Path]
	return ok
}
func GetNewConf(newConf []common.CollectEntry) {
	ttMgr.confChan <- newConf
}
