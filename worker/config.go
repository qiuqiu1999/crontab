package worker

import (
	"encoding/json"
	"io/ioutil"
)

// 程序配置
type Config struct {
	EtcdEndpoints         []string `json:"etcdEndpoints"`
	EtcdDialTimeout       int      `json:"etcdDialTimeout"`
	MongodbUri            string   `json:"mongodbUri"`
	MongodbConnectTimeout int      `json:"mongodbConnectTimeout"`
	JobLogBatchSize       int      `json:"jobLogBatchSize"`
	JobLogCommitTimeout   int      `json"jobLogCommitTimeout"`
}

var (
	// 单例
	G_config *Config
)

// 加载配置
func InitConfig(filename string) error {
	var (
		content []byte
		conf    Config
		err     error
	)

	// 1, 把配置文件读进来
	if content, err = ioutil.ReadFile(filename); err != nil {
		return err
	}

	// 2, 做JSON反序列化
	if err = json.Unmarshal(content, &conf); err != nil {
		return err
	}

	// 3, 赋值单例
	G_config = &conf

	return nil
}
