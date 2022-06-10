package worker

import (
	"context"
	"github.com/coreos/etcd/clientv3"
	"github.com/qiuqiu1999/crontab/common"
	"net"
	"time"
)

// 注册节点到etcd： /cron/workers/IP地址
type Register struct {
	etcd    *common.Etcd
	localIP string // 本机IP
}

var G_register *Register

// 获取本机网卡IP
func getLocalIP() (ipv4 string, err error) {
	var (
		addrs   []net.Addr
		addr    net.Addr
		ipNet   *net.IPNet // IP地址
		isIpNet bool
	)
	// 获取所有网卡
	if addrs, err = net.InterfaceAddrs(); err != nil {
		return
	}
	// 取第一个非lo的网卡IP
	for _, addr = range addrs {
		// 这个网络地址是IP地址: ipv4, ipv6
		if ipNet, isIpNet = addr.(*net.IPNet); isIpNet && !ipNet.IP.IsLoopback() {
			// 跳过IPV6
			if ipNet.IP.To4() != nil {
				ipv4 = ipNet.IP.String() // 192.168.1.1
				return
			}
		}
	}

	err = common.ERR_NO_LOCAL_IP_FOUND
	return
}

// 注册到/cron/workers/IP, 并自动续租
func (register *Register) keepOnline() {
	var (
		leaseGrantResp *clientv3.LeaseGrantResponse
		err            error
		keepAliveChan  <-chan *clientv3.LeaseKeepAliveResponse
		keepAliveResp  *clientv3.LeaseKeepAliveResponse
		cancelCtx      context.Context
		cancelFunc     context.CancelFunc
	)
	// 注册路径
	regKey := common.JOB_WORKER_DIR + register.localIP

	for {

		cancelFunc = nil

		// 创建租约
		if leaseGrantResp, err = register.etcd.Lease.Grant(context.TODO(), 10); err != nil {
			time.Sleep(1 * time.Second)
			continue
		}

		// 自动续租
		if keepAliveChan, err = register.etcd.Lease.KeepAlive(context.TODO(), leaseGrantResp.ID); err != nil {
			time.Sleep(1 * time.Second)
			continue
		}

		cancelCtx, cancelFunc = context.WithCancel(context.TODO())

		// 注册到etcd
		if _, err = register.etcd.Kv.Put(cancelCtx, regKey, "", clientv3.WithLease(leaseGrantResp.ID)); err != nil {
			time.Sleep(1 * time.Second)
			cancelFunc()
			continue
		}

		// 处理续租应答
		for {
			select {
			case keepAliveResp = <-keepAliveChan:
				if keepAliveResp == nil { // 续租失败
					time.Sleep(1 * time.Second)
					continue
				}
			}
		}
	}
}

func InitRegister() error {
	var (
		etcd    *common.Etcd
		err     error
		localIp string
	)
	if etcd, err = common.NewEtcd(&common.EtcdConfig{
		EtcdEndpoints:   G_config.EtcdEndpoints,
		EtcdDialTimeout: time.Duration(G_config.EtcdDialTimeout),
	}); err != nil {
		return err
	}

	// 本机IP
	if localIp, err = getLocalIP(); err != nil {
		return err
	}

	G_register = &Register{
		etcd:    etcd,
		localIP: localIp,
	}

	// 服务注册
	go G_register.keepOnline()

	return nil
}
