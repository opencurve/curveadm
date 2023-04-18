package scripts

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	TMPL_ECHO = `echo "hello, {{.name}}"`
)

func TestGetScript(t *testing.T) {
	assert := assert.New(t)

	script, err := GetScript(TMPL_ECHO,
		map[string]interface{}{
			"name": "curveadm",
		})
	assert.Nil(err)
	assert.Equal(script, `echo "hello, curveadm"`)
}
