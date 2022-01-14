package conf

import "time"

type AppConf struct {
	KafkaConf `ini:"kafka"`
	EtcdConf  `ini:"etcd"`
}
type KafkaConf struct {
	Address  string `ini:"address"`
	Chansize int    `ini:"chan_size"`
}
type EtcdConf struct {
	Address string        `ini:"address"`
	Timeout time.Duration `ini:"timeout"`
}
