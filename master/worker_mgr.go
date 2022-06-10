package master

import (
	"github.com/qiuqiu1999/crontab/common"
	"time"
)

// /cron/workers/
type WorkerMgr struct {
	etcd *common.Etcd
}

var G_workerMgr *WorkerMgr

func InitWorkerMgr() error {
	var (
		etcd *common.Etcd
		err  error
	)
	if etcd, err = common.NewEtcd(&common.EtcdConfig{
		EtcdEndpoints:   G_config.EtcdEndpoints,
		EtcdDialTimeout: time.Duration(G_config.EtcdDialTimeout),
	}); err != nil {
		return err
	}

	G_workerMgr = &WorkerMgr{
		etcd: etcd,
	}
	return nil
}

// 获取在线worker列表
func (workerMgr *WorkerMgr) ListWorkers() ([]string, error) {
	workerArr := make([]string, 0)

	// 获取目录下所有Kv
	getResp, err := workerMgr.etcd.GetKeyValue(common.JOB_WORKER_DIR)
	if err != nil {
		return nil, err
	}
	// 解析每个节点的IP
	for _, kv := range getResp.Kvs {
		// kv.Key : /cron/workers/xxx.xxx.xxx.xxx
		workerIP := common.ExtractWorkerIP(string(kv.Key))
		workerArr = append(workerArr, workerIP)
	}
	return workerArr, nil
}
