package taillog

import (
	"context"
	"time"

	"code.kawai.com/wdfky/logagent/kafka"
	"github.com/Shopify/sarama"
	"github.com/hpcloud/tail"
	"github.com/sirupsen/logrus"
)

type tailTask struct {
	path   string
	topic  string
	tails  *tail.Tail
	ctx    context.Context
	cancel context.CancelFunc
}

func newTailTask(path, topic string) *tailTask {
	ctx, cance := context.WithCancel(context.Background())
	tt := &tailTask{
		path:   path,
		topic:  topic,
		ctx:    ctx,
		cancel: cance,
	}
	return tt
}
func (t *tailTask) Init() (err error) {
	cfg := tail.Config{
		ReOpen:    true,                                 //重新打开
		Follow:    true,                                 //是否跟随
		Location:  &tail.SeekInfo{Offset: 0, Whence: 2}, //从文件那个地方读
		MustExist: false,                                //文件不存在不报错
		Poll:      true,                                 //使用Linux的Poll函数，poll的作用是把当前的文件指针挂到等待队列
	}
	t.tails, err = tail.TailFile(t.path, cfg)
	if err != nil {
		logrus.Errorln("tail file faliled, err", err)
		return
	}
	return
}
func (t *tailTask) run() {
	//读取日志发往kafka
	for {
		select {
		case <-t.ctx.Done():
			logrus.Infof("path:%s is stopping...", t.path)
			return

		case line, ok := <-t.tails.Lines:
			if !ok {
				logrus.Warn("tail flle close reopen, path:%s\n", t.path)
				time.Sleep(time.Second)
				continue
			}
			msg := &sarama.ProducerMessage{}
			msg.Topic = t.topic
			msg.Value = sarama.StringEncoder(line.Text)
			kafka.MsgChan <- msg
		}
	}
}
