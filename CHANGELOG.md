v0.0.9
---

* bugfix(*): fixed download and upload file conflict for deploy in local machine

v0.0.8
---


v0.0.7
---

v0.0.6
---

* feature(fs): now we will report CurveFS cluster usage to curve center (issue #18)

---

v0.0.5
---

* improve(fs): synchronize tools config to its default path
* imporve(fs): trim ending slash of mountpoint when mount/umount/check
* bugfix(fs): added volume for log dir and data dir for client container
* bugfix(fs): specify host network for client container
* bugfix(fs): create the missing configure directory when synchronize tools config
* bugfix(*): use empty string for default binary option of reload command (issue #11)

---

v0.0.4
---

* feature(*): now we can get support from curve team (issue #6)
* feature(*): support replace service binary without re-deploy (issue #2)

---

v0.0.3
---
* feature(*): support export and import cluster database
* improve(fs): wait mds leader election success before create curvefs topology
* improve(*): change the current working directory when enter service container
* bugfix(fs): use fusermount to umount filesystem instead of stop fuse client (issue #3)
 
---

v0.0.2 
--- 
* feature(*): support upgrade curveadm to latest version
