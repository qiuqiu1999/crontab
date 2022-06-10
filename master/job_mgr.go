package master

import (
	"encoding/json"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/qiuqiu1999/crontab/common"
	"time"
)

// 任务管理器
type JobMgr struct {
	etcd *common.Etcd
}

var G_jobMgr *JobMgr

// 初始化管理器
func InitJobMgr() error {
	etcd, err := common.NewEtcd(&common.EtcdConfig{
		EtcdEndpoints:   G_config.EtcdEndpoints,
		EtcdDialTimeout: time.Duration(G_config.EtcdDialTimeout),
	})
	if err != nil {
		return err
	}

	G_jobMgr = &JobMgr{
		etcd: etcd,
	}
	return nil
}

// 保存任务
func (jobMgr *JobMgr) SaveJob(job *common.Job) (*common.Job, error) {
	// 把任务保存到/cron/jobs/任务名 -> json
	var (
		jobKey    string
		jobValue  []byte
		putResp   *clientv3.PutResponse
		oldJobObj common.Job
		err       error
	)

	// etcd的保存key
	jobKey = common.JOB_SAVE_DIR + job.Name
	// 任务信息json
	if jobValue, err = json.Marshal(job); err != nil {
		return nil, err
	}
	// 保存到etcd
	if putResp, err = jobMgr.etcd.PutKeyValue(jobKey, string(jobValue)); err != nil {
		return nil, err
	}
	// 如果是更新, 那么返回旧值
	if putResp.PrevKv != nil {
		// 对旧值做一个反序列化
		if err = json.Unmarshal(putResp.PrevKv.Value, &oldJobObj); err != nil {
			return nil, err
		}
	}
	return &oldJobObj, nil
}

// 删除任务
func (jobMgr *JobMgr) DeleteJob(name string) (oldJob *common.Job, err error) {
	var (
		jobKey    string
		delResp   *clientv3.DeleteResponse
		oldJobObj common.Job
	)

	// etcd中保存任务的key
	jobKey = common.JOB_SAVE_DIR + name

	// 从etcd中删除它
	if delResp, err = jobMgr.etcd.DeleteKey(jobKey); err != nil {
		return
	}

	// 返回被删除的任务信息
	if len(delResp.PrevKvs) != 0 {
		// 解析一下旧值, 返回它
		if err = json.Unmarshal(delResp.PrevKvs[0].Value, &oldJobObj); err != nil {
			err = nil
			return
		}
		oldJob = &oldJobObj
	}
	return
}

// 列举任务
func (jobMgr *JobMgr) ListJobs() ([]*common.Job, error) {
	var (
		dirKey  string
		getResp *clientv3.GetResponse
		kvPair  *mvccpb.KeyValue
		job     *common.Job
		err     error
	)

	// 任务保存的目录
	dirKey = common.JOB_SAVE_DIR

	// 获取目录下所有任务信息
	if getResp, err = jobMgr.etcd.GetKeyValue(dirKey); err != nil {
		return nil, err
	}

	// 初始化数组空间
	jobList := make([]*common.Job, 0)
	// len(jobList) == 0

	// 遍历所有任务, 进行反序列化
	for _, kvPair = range getResp.Kvs {
		job = &common.Job{}
		if err = json.Unmarshal(kvPair.Value, job); err != nil {
			err = nil
			continue
		}
		jobList = append(jobList, job)
	}
	return jobList, err
}

// 杀死任务
func (jobMgr *JobMgr) KillJob(name string) error {
	// 更新一下key=/cron/killer/任务名
	var (
		killerKey string
		err       error
	)

	// 通知worker杀死对应任务
	killerKey = common.JOB_KILLER_DIR + name

	// 让worker监听到一次put操作, 创建一个租约让其稍后自动过期即可
	if err = jobMgr.etcd.PutKeyWithLease(killerKey, 1); err != nil {
		return err
	}
	return nil
}
