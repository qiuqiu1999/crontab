### 分布式crontab管理后台
使用etcd做服务注册与发现, MongoDB存储执行日志
```
# 编译
make build

# 启动后台管理服务
./bin/master

# 启动worker服务
./bin/worker
```