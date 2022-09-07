module github.com/opencurve/curveadm

go 1.16

require (
	github.com/Microsoft/go-winio v0.5.1 // indirect
	github.com/Microsoft/hcsshim v0.9.0 // indirect
	github.com/containerd/continuity v0.2.1 // indirect
	github.com/docker/cli v20.10.9+incompatible
	github.com/docker/docker v20.10.9+incompatible // indirect
	github.com/docker/docker-credential-helpers v0.6.4 // indirect
	github.com/fatih/color v1.13.0
	github.com/fvbommel/sortorder v1.0.2 // indirect
	github.com/go-resty/resty/v2 v2.7.0
	github.com/google/uuid v1.2.0
	github.com/jpillora/longestcommon v0.0.0-20161227235612-adb9d91ee629
	github.com/kpango/glg v1.6.11
	github.com/mattn/go-sqlite3 v1.6.0
	github.com/melbahja/goph v1.3.0
	github.com/mitchellh/hashstructure/v2 v2.0.2
	github.com/moby/sys/mount v0.2.0 // indirect
	github.com/moby/term v0.0.0-20210619224110-3f7ff695adc6
	github.com/pingcap/log v0.0.0-20210906054005-afc726e70354
	github.com/sergi/go-diff v1.2.0
	github.com/spf13/cobra v1.3.0
	github.com/spf13/viper v1.10.0
	github.com/stretchr/testify v1.7.0
	github.com/theupdateframework/notary v0.7.0 // indirect
	github.com/vbauerster/mpb/v7 v7.1.5
	go.uber.org/zap v1.21.0
	golang.org/x/crypto v0.0.0-20220829220503-c86fa9a7ed90
)

replace github.com/melbahja/goph v1.3.0 => github.com/Wine93/goph v0.0.0-20220907030102-3564d9035b54
