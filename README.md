## 分布式crontab管理后台
分布式管理crontab， 提供WEB UI进行管理，
使用etcd做服务注册与发现, MongoDB存储执行日志

## 优势
- 可部署多台服务器，可靠的服务注册与发现，解决单点故障问题。
- 提供友好的UI管理界面，方便配置修改与查看
- 支持秒级定时任务

## 编译方式
```shell
# 编译
make build
```

## 使用方式
```
# 启动后台管理服务
./bin/master

# 启动worker服务
./bin/worker
```

## 开源协议

本项目源码采用 [The MIT License](https://opensource.org/licenses/MIT) 开源协议。
<details>
<summary>关于 MIT 协议</summary>

> MIT 协议可能是几大开源协议中最宽松的一个，核心条款是：
>
> 该软件及其相关文档对所有人免费，可以任意处置，包括使用，复制，修改，合并，发表，分发，再授权，或者销售。唯一的限制是，软件中必须包含上述版 权和许可提示。
>
> 这意味着：
> - 你可以自由使用，复制，修改，可以用于自己的项目。
> - 可以免费分发或用来盈利。
> - 唯一的限制是必须包含许可声明。
>
> MIT 协议是所有开源许可中最宽松的一个，除了必须包含许可声明外，再无任何限制。
>
> *以上文字来自 [五种开源协议GPL,LGPL,BSD,MIT,Apache](https://www.oschina.net/question/54100_9455) 。*
</details> 