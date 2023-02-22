v0.2.0
---
  * Improve: add CI (build and test action), thanks to Tsong Lew.
  * Improve: update go version since Go 1.18 is required to build rivo/uniseg, thanks to Tsong Lew.
  * Feature(exec): support execute command in specified container, thanks to Wangpan.
  * Feature(target): support specify target block size, thanks to mfordjody.
  * Feature(format): support stop formmating, thanks to DemoLiang.
  * Feature(mount): support setting environment variable for client container.
  * Feature(hosts): support setting SSH address which only for SSH connect.
  * Feature(playbook): add playbook which user can run any scripts in any hosts.
  * Feature(playbook): support deploy memcache by playbook, thanks to SiKu.
  * Feature(playbook): support setting host environment variable.
  * Feature(playbook): support pass arguments to run scripts.
  * Feature(playbook): support exclude and intersection pattern for playbook label.
  * Feature(playground): now we can run playground by specified container image.
  * Feature: add curvefs-fuse-bt bpftrace tool.
  * Fix: set environment variable failed while executing command.
  * Fix: its no need to become user when execute command in local.
  * Fix(map): map a volume which name contain underscore symbol.
  * Fix(format): wrong sed expression in become_user modle.

v0.0.10
---
* bugfix(*): fixed read & install file error

v0.0.9
---
* bugfix(*): fixed download and upload file conflict for deploy in local machine

v0.0.6
---
* feature(fs): now we will report CurveFS cluster usage to curve center (issue #18)

v0.0.5
---
* improve(fs): synchronize tools config to its default path
* imporve(fs): trim ending slash of mountpoint when mount/umount/check
* bugfix(fs): added volume for log dir and data dir for client container
* bugfix(fs): specify host network for client container
* bugfix(fs): create the missing configure directory when synchronize tools config
* bugfix(*): use empty string for default binary option of reload command (issue #11)


v0.0.4
---
* feature(*): now we can get support from curve team (issue #6)
* feature(*): support replace service binary without re-deploy (issue #2)

v0.0.3
---
* feature(*): support export and import cluster database
* improve(fs): wait mds leader election success before create curvefs topology
* improve(*): change the current working directory when enter service container
* bugfix(fs): use fusermount to umount filesystem instead of stop fuse client (issue #3)

v0.0.2
---
* feature(*): support upgrade curveadm to latest version
