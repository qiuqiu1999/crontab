package worker

import (
	"context"
	"fmt"
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

// 监听任务变化
func (jobMgr *JobMgr) watchJobs() error {
	var (
		getResp            *clientv3.GetResponse
		kvpair             *mvccpb.KeyValue
		job                *common.Job
		watchStartRevision int64
		watchChan          clientv3.WatchChan
		watchResp          clientv3.WatchResponse
		watchEvent         *clientv3.Event
		jobName            string
		jobEvent           *common.JobEvent
		err                error
	)

	// 1, get一下/cron/jobs/目录下的所有任务，并且获知当前集群的revision
	if getResp, err = jobMgr.etcd.Kv.Get(context.TODO(), common.JOB_SAVE_DIR, clientv3.WithPrefix()); err != nil {
		return err
	}

	// 当前有哪些任务
	for _, kvpair = range getResp.Kvs {
		// 反序列化json得到Job
		if job, err = common.UnpackJob(kvpair.Value); err == nil {
			jobEvent = common.BuildJobEvent(common.JOB_EVENT_SAVE, job)
			// 同步给scheduler(调度协程)
			G_scheduler.PushJobEvent(jobEvent)
		}
	}

	// 2, 从该revision向后监听变化事件
	go func() { // 监听协程
		// 从GET时刻的后续版本开始监听变化
		watchStartRevision = getResp.Header.Revision + 1
		// 监听/cron/jobs/目录的后续变化
		watchChan = jobMgr.etcd.Watcher.Watch(context.TODO(), common.JOB_SAVE_DIR, clientv3.WithRev(watchStartRevision), clientv3.WithPrefix())
		// 处理监听事件
		for watchResp = range watchChan {
			for _, watchEvent = range watchResp.Events {
				switch watchEvent.Type {
				case mvccpb.PUT: // 任务保存事件
					if job, err = common.UnpackJob(watchEvent.Kv.Value); err != nil {
						continue
					}
					// 构建一个更新Event
					jobEvent = common.BuildJobEvent(common.JOB_EVENT_SAVE, job)
				case mvccpb.DELETE: // 任务被删除了
					// Delete /cron/jobs/job10
					jobName = common.ExtractJobName(string(watchEvent.Kv.Key))

					job = &common.Job{Name: jobName}

					// 构建一个删除Event
					jobEvent = common.BuildJobEvent(common.JOB_EVENT_DELETE, job)
				}
				// 变化推给scheduler
				G_scheduler.PushJobEvent(jobEvent)
			}
		}
	}()
	return nil
}

// 监听强杀任务通知
func (jobMgr *JobMgr) watchKiller() {
	var (
		watchChan  clientv3.WatchChan
		watchResp  clientv3.WatchResponse
		watchEvent *clientv3.Event
		jobEvent   *common.JobEvent
		jobName    string
		job        *common.Job
	)
	// 监听/cron/killer目录
	go func() { // 监听协程
		// 监听/cron/killer/目录的变化
		watchChan = jobMgr.etcd.Watcher.Watch(context.TODO(), common.JOB_KILLER_DIR, clientv3.WithPrefix())
		// 处理监听事件
		for watchResp = range watchChan {
			for _, watchEvent = range watchResp.Events {
				switch watchEvent.Type {
				case mvccpb.PUT: // 杀死任务事件
					jobName = common.ExtractKillerName(string(watchEvent.Kv.Key))
					job = &common.Job{Name: jobName}
					jobEvent = common.BuildJobEvent(common.JOB_EVENT_KILL, job)
					// 事件推给scheduler
					G_scheduler.PushJobEvent(jobEvent)
				case mvccpb.DELETE: // killer标记过期, 被自动删除
				}
			}
		}
	}()
}

// 初始化管理器
func InitJobMgr() error {

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

	timeoutCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, err = etcd.Client.Status(timeoutCtx, G_config.EtcdEndpoints[0])
	if err != nil {
		return fmt.Errorf("error checking etcd status: %v", err)
	}

	// 赋值单例
	G_jobMgr = &JobMgr{
		etcd: etcd,
	}

	// 启动任务监听
	G_jobMgr.watchJobs()

	// 启动监听killer
	G_jobMgr.watchKiller()

	return nil
}

// 创建任务执行锁
func (jobMgr *JobMgr) CreateJobLock(jobName string) *JobLock {
	return InitJobLock(jobName, jobMgr.etcd.Kv, jobMgr.etcd.Lease)
}
