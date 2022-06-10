
目录
---
* [配置样例](#配置样例)
* [配置层级](#配置层级)
* [变量](#变量)
* [配置项](#配置项)
  * [user](#user)
  * [ssh_port](#ssh_port)
  * [private_key_file](#private_key_file)
  * [report_usage](#report_usage)
  * [log_dir](#log_dir)
  * [data_dir](#data_dir)
  * [container_image](#container_image)
  * [s3.ak](#s3ak)
  * [s3.sk](#s3sk)
  * [s3.endpoint](#s3endpoint)
  * [s3.bucket_name](#s3bucket_name)
  * [variable](#variable)
  * [其他](#其他)

配置样例
===

```shell
global:
  user: curve
  ssh_port: 22
  private_key_file: /home/curve/.ssh/id_rsa
  report_usage: true
  data_dir: /home/${user}/curvefs/data/${service_role}${service_sequence}
  log_dir: /home/${user}/curvefs/logs/${service_role}${service_sequence}
  container_image: opencurvedocker/curvefs:latest
  variable:
    target: 10.0.1.1

etcd_services:
  config:
    listen.ip: ${target}
    listen.port: ${service_sequence}2380  # 12380,22380,32380
    listen.client_port: ${service_sequence}2379  # 12379,22379,32379
  deploy:
    - host: ${target}
    - host: ${target}
    - host: ${target}

mds_services:
  config:
    listen.ip: ${target}
    listen.port: ${service_sequence}6700  # 16700,26700,36700
    listen.dummy_port: ${service_sequence}7700  # 17700,27700,37700
  deploy:
    - host: ${target}
    - host: ${target}
    - host: ${target}

metaserver_services:
  config:
    listen.ip: ${target}
    listen.port: ${service_sequence}6701
    metaserver.loglevel: 0
  deploy:
    - host: ${target}
    - host: ${target}
    - host: ${target}
      config:
        metaserver.loglevel: 3

```

[返回目录](#目录)

配置层级
===

topology 的配置项分为 3 个层级：

* 按照优先级从低到高依次为:
    * 全局配置：`global`
    * 同角色服务配置：`service.config`
    * 指定服务配置：`deploy.config`
* 配置项会根据层级依次做合并操作，如 `全局配置` 会合并到 `同角色服务配置`
* 针对相同配置项，优先采用优先级高的配置值

[返回目录](#目录)

变量
===

`curveadm` 内置变量系统，这些变量可用于配置 `topology` 以增加配置的灵活性，变量分为两类：

* 用户自定义变量：用户可以自定义变量，以减少 `topology` 中的重复配置
* 内嵌变量：`curveadm` 内置了近 20 个变量，详见 [topology_variables](https://github.com/opencurve/curveadm/blob/master/internal/configure/topology_variables.go#L32)

[返回目录](#目录)

配置项
===

user
---

> 默认值：
> 
> 示例：curve
> 
> 说明：连接远端主机 `SSH` 服务的用户 <br>
> 该用户同样用于执行部署操作，请确保该用户有 `sudo` 权限，因为该用户将用于操作 `docker cli`

[返回目录](#目录)
 
ssh_port
---

> 默认值：22
> 
> 示例：1180
> 
> 说明：连接远端主机 `SSH` 服务的端口

[返回目录](#目录)

private_key_file
---

> 默认值：
> 
> 示例：/home/curve/.ssh/id_rsa
> 
> 说明：用于连接远端主机 `SSH` 服务的私钥路径

[返回目录](#目录)

report_usage
---

> 默认值：true
>
> 示例：false
>
> 说明：是否匿名上报用户集群使用量 <br>
> 开启该选项后，curveadm 会匿名上报用户集群 `UUID` 以及集群使用量，来帮助 curve 团队更好的了解用户及改进服务

[返回目录](#目录)

log_dir
---

> 默认值：
>
> 示例：/mnt/logs
>
> 说明：保存服务日志的目录 <br>
> 如果不配置该选项，日志默认保存在容器内的指定目录，一旦容器被清理，日志将会随之删除

[返回目录](#目录)

data_dir
---

> 默认值：
> 
> 示例：/mnt/data
> 
> 说明：保存服务数据的目录 <br>
> 如果不配置该选项，数据默认保存在容器内的指定目录，一旦容器被清理，数据将会随之丢失

[返回目录](#目录)

container_image
---

> 默认值：
> 
> 示例：opencurvedocker/curvefs:latest
> 
> 说明：服务镜像

[返回目录](#目录)

s3.ak
---

> 默认值：
> 
> 示例：minioadmin
> 
> 说明：访问 `S3` 服务的 `Access Key`

[返回目录](#目录)

s3.sk
---

> 默认值：
>
> 示例：minioadmin
>
> 说明：访问 `S3` 服务的 `Secret Key`

[返回目录](#目录)

s3.endpoint
---

> 默认值：
>
> 示例：http://127.0.0.1:9000
>
> 说明：`S3` 服务地址
 
[返回目录](#目录)

s3.bucket_name
---

> 默认值：
>
> 示例：curvefs-test
>
> 说明：`S3` 服务中的桶名

[返回目录](#目录)

variable
---

> 默认值：
> 
> 示例：target: 10.0.1.1
> 
> 说明：用户自定义变量

[返回目录](#目录)

其他
---

* `topology` 中的其余配置项与 `curvefs` 项目中的配置项保持一致，包括默认值，详见 [curvefs/conf](https://github.com/opencurve/curve/tree/fs/curvefs/conf)
* 若想修改相关配置，在 `topology` 中修改即可，如修改 metaserver 中的日志等级，你可以在 `topology` 中增加以下配置项：
```shell
metaserver.loglevel: 9
```

[返回目录](#目录)
