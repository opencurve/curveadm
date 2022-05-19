# Shell plugin

Execute shell commands on targets

## Usage

* execute shell command in localhost

```shell
$ curveadm plugin run shell --arg cmd=hostname --arg local=true
```

* execute shell command in remote host (host1, host2, host3)

```shell
$ curveadm plugin run shell --hosts 'host1:host2:host3' --arg cmd=hostname
```
