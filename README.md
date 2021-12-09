# CurveAdm

Deploy and manage Curve cluster

> NOTE: CurveAdm only support [CurveFS](https://github.com/opencurve/curve/tree/fs) now, the [CurveBS](https://github.com/opencurve/curve) will be supported later

Table of Contents
===

* [Requirement](#requirement)
    * [Docker](#docker)
    * [Minio (optional)](#minio)
* [Installation](#installation)
* [Usage](#usage)
    * [Deploy Cluster](#deploy-cluster) 
    * [Mount FileSystem](#mount-filesystem)
    * [Umount FileSystem](#umount-filesystem)
* [Devops](#devops)
* [Ask for help](#ask-for-help)

Requirement
---

### Docker

CurveAdm depends on docker, please [install docker](https://docs.docker.com/engine/install/) first.

Please make sure the docker daemon has running, you can run the below command to verify:

```shell
sudo docker run hello-world
```

This command downloads a test image and runs it in a container. When the container runs, it prints a message and exits.

### Minio

Current version CurveFS only supports S3 storage backend, so you need deploy a S3 storage service or use object storage of public cloud provider, eg. AWS S3, Alibaba cloud OSS, Tencent cloud COS and so on. Here we use MinIO for S3 storage backend to deploy a CurveFS cluster:

```shell
# deploy a standalone MinIO S3 storage service
sudo docker run -d --name minio \
    -p 9000:9000 \
    -p 9900:9900 \
    -v minio-data:/data \     # minio-data is a local path, you should create this directory before runing the minio container
    --restart unless-stopped \
    minio/minio server /data --console-address ":9900"

```
The default Access Key and Secret Key of the root user are both `minioadmin`, the endpoint of S3 service is `http://$IP:9000`, you should create a bucket on browser with URL `http://$IP:9900`, then you can use these infomation to deploy a CurveFS cluster. You can get more reference at [deploy-minio-standalone](https://docs.min.io/minio/baremetal/installation/deploy-minio-standalone.html).

[Back to TOC](#table-of-contents)

Installation
---

```shell
sh -c "$(curl -fsSL https://curveadm.nos-eastchina1.126.net/script/install.sh)"
```

[Back to TOC](#table-of-contents)

Usage
---

### Deploy Cluster

Prepare cluster topology, you can refer to the sample configuration:

* [cluster](examples/cluster/topology.yaml)
* [stand-alone](examples/stand-alone/topology.yaml)

```shell
vi topology.yaml
```

Add cluster with specified topology:

```shell
curveadm cluster add c1 -f topology.yaml
```

Switch cluster:

```shell
curveadm cluster checkout c1
```

Deploy cluster:
```shell
curveadm deploy
```

Show cluster status:
```shell
curveadm status
```

Show cluster status with detail(show log dir and data dir used by service):
```shell
curveadm status -v
```


### Mount FileSystem

Prepare client config, you can refer to the sample configuration:

* [cluster](examples/cluster/client.yaml)
* [stand-alone](examples/stand-alone/client.yaml)

```shell
vi client.yaml
```

Mount filesystem:

```shell
sudo curveadm mount NAME-OF-CURVEFS MOUNTPONT -c client.yaml
```

### Umount FileSystem

```shell
sudo curveadm umount MOUNTPOINT
```

[Back to TOC](#table-of-contents)

Devops
---

Run `curveadm -h` for more information.

[Back to TOC](#table-of-contents)

# Ask for help

If you encounter an unsolvable problem during deployment or using Curve, you can use the `support` command to seek help from the curve team. After executing this command, all Curve service logs and configuration files will be packaged and collected, and encrypted and uploaded to On our log collection server, so that the Curve team can analyze and solve problems:

```shell
curveadm support
```
After the `support` command is executed, the logs and config files will be packaged and uploaded. If the upload is successful, a secret key will be returned. You can get help by telling the secret key to the Curve team. You can contact the Curve team by adding the WeChat account `opencurve`.

[Back to TOC](#table-of-contents)
