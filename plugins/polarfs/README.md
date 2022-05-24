# PolarFS plugin

This plugin can help you install [PolarFS](PolarFS) to multiple host machines.

NOTE: now we only support deploy in debian system.

## Usage

* Install PolarFS on targets
 
```shell
$ curveadm plugin run polarfs --hosts 'host1:host2:host3' --arg mds_listen_addr='10.0.1.1:6700,10.0.1.2:6700:10.0.1.3:6700'
```

[PolarPFS]: https://github.com/opencurve/PolarDB-FileSystem
