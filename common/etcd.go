package common

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"time"
)

type Etcd struct {
	Client  *clientv3.Client
	Kv      clientv3.KV
	Lease   clientv3.Lease
	Watcher clientv3.Watcher
}

type EtcdConfig struct {
	EtcdEndpoints   []string
	EtcdDialTimeout time.Duration
}

func NewEtcd(etcdConfig *EtcdConfig) (*Etcd, error) {
	config := clientv3.Config{
		Endpoints:   etcdConfig.EtcdEndpoints,
		DialTimeout: etcdConfig.EtcdDialTimeout * time.Millisecond,
	}

	cli, err := clientv3.New(config)
	if err != nil {
		return nil, err
	}

	timeoutCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, err = cli.Status(timeoutCtx, config.Endpoints[0])
	if err != nil {
		return nil, fmt.Errorf("error checking etcd status: %v", err)
	}

	kv := clientv3.NewKV(cli)
	lease := clientv3.NewLease(cli)
	watcher := clientv3.NewWatcher(cli)

	etcd := &Etcd{
		Client:  cli,
		Kv:      kv,
		Lease:   lease,
		Watcher: watcher,
	}
	return etcd, nil
}

//
func (etcd *Etcd) GetKeyValue(key string) (*clientv3.GetResponse, error) {
	getResp, err := etcd.Kv.Get(context.TODO(), key, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	return getResp, err
}

func (etcd *Etcd) PutKeyValue(key, value string) (*clientv3.PutResponse, error) {
	// 保存到etcd
	putResp, err := etcd.Kv.Put(context.TODO(), key, value, clientv3.WithPrevKV())
	if err != nil {
		return nil, err
	}
	return putResp, nil
}

func (etcd *Etcd) DeleteKey(key string) (*clientv3.DeleteResponse, error) {
	// 从etcd中删除它
	delResp, err := etcd.Kv.Delete(context.TODO(), key, clientv3.WithPrevKV())
	if err != nil {
		return nil, err
	}
	return delResp, nil
}

func (etcd *Etcd) PutKeyWithLease(key string, timeout int64) error {

	// 创建一个租约让其稍后自动过期即可
	leaseGrantResp, err := etcd.Lease.Grant(context.TODO(), timeout)
	if err != nil {
		return err
	}

	// 租约ID
	leaseId := leaseGrantResp.ID
	// 设置killer标记
	_, err = etcd.Kv.Put(context.TODO(), key, "", clientv3.WithLease(leaseId))
	if err != nil {
		return err
	}
	return nil
}
