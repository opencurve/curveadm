package fs

import (
	"strings"
	"testing"

	"github.com/opencurve/curveadm/internal/configure"
	"github.com/stretchr/testify/assert"
)

var (
	KEY_KIND  = strings.ToLower(configure.KEY_KIND)
	KEY_ADDRS = strings.ToLower(configure.KEY_CURVEFS_LISTEN_MDS_ADDRS)
	KEY_ENV   = strings.ToLower(configure.KEY_ENVIRONMENT)
)

func run(t *testing.T, config map[string]interface{}, envs []string) {
	assert := assert.New(t)
	cc, err := configure.NewClientConfig(config)
	assert.Nil(err)
	assert.Equal(envs, getEnvironments(cc))
}

func TestConfigureEnv_Basic(t *testing.T) {
	run(t, map[string]interface{}{
		KEY_KIND:  "curvefs",
		KEY_ADDRS: "1.1.1.1",
	}, []string{
		"LD_PRELOAD=/usr/local/lib/libjemalloc.so",
	})

	run(t, map[string]interface{}{
		KEY_KIND:  "curvefs",
		KEY_ADDRS: "1.1.1.1",
		KEY_ENV:   "MALLOC_CONF=prof:true,lg_prof_interval:26,prof_prefix:/curvefs/client/logs/jeprof.out",
	}, []string{
		"LD_PRELOAD=/usr/local/lib/libjemalloc.so",
		"MALLOC_CONF=prof:true,lg_prof_interval:26,prof_prefix:/curvefs/client/logs/jeprof.out",
	})

	run(t, map[string]interface{}{
		KEY_KIND:  "curvefs",
		KEY_ADDRS: "1.1.1.1",
		KEY_ENV:   "NAME=jack AGE=18 FROM=china",
	}, []string{
		"LD_PRELOAD=/usr/local/lib/libjemalloc.so",
		"NAME=jack",
		"AGE=18",
		"FROM=china",
	})

	run(t, map[string]interface{}{
		KEY_KIND:  "curvefs",
		KEY_ADDRS: "1.1.1.1",
		KEY_ENV:   "LD_PRELOAD=",
	}, []string{
		"LD_PRELOAD=/usr/local/lib/libjemalloc.so",
		"LD_PRELOAD=",
	})
}
