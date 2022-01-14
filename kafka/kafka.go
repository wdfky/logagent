package kafka

import (
	"github.com/Shopify/sarama"
	"github.com/sirupsen/logrus"
)

var (
	client  sarama.SyncProducer //声明一个全局的kafka生产者client
	MsgChan chan *sarama.ProducerMessage
)

//初始化client
func Init(addr []string, chan_size int) (err error) {
	config := sarama.NewConfig()
	//tailf包使用
	config.Producer.RequiredAcks = sarama.WaitForAll          //发完数据需要leader和follow都确认
	config.Producer.Partitioner = sarama.NewRandomPartitioner //新选出一个partition
	config.Producer.Return.Successes = true                   //成功交付的消息将在seccess channel返回
	//连接kafka
	client, err = sarama.NewSyncProducer(addr, config)
	if err != nil {
		logrus.Error("producer closed, err:", err)
		return
	}
	MsgChan = make(chan *sarama.ProducerMessage, chan_size)
	go SendToKafka()
	return
}
func SendToKafka() {
	for {
		select {
		case msg := <-MsgChan:
			pid, offset, err := client.SendMessage(msg)
			if err != nil {
				logrus.Warning("send massage failed, err:", err)
				return
			}
			logrus.Infof("pid:%v offset:%v\n", pid, offset)
		}
	}
}
