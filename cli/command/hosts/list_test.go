package hosts

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	data = `
hosts:
  - host: host1
    hostname: 1.1.1.1
    labels:
      - all
      - host1
      - group1
  - host: host2
    hostname: 2.2.2.2
    labels:
      - all
      - host2
      - group1
      - group2
  - host: host3
    hostname: 3.3.3.3
    labels:
      - all
      - host3
      - group2
`
)

func run(t *testing.T, data string, labels []string, out []string) {
	assert := assert.New(t)
	hcs, err := filter(data, labels)
	assert.Nil(err)
	assert.Equal(len(hcs), len(out))
	for i, hc := range hcs {
		assert.Equal(hc.GetHost(), out[i])
	}
}

func TestPlaybookLabel_ParsePattern(t *testing.T) {
	assert := assert.New(t)
	include, exclude, intersect := parsePattern([]string{"group1", "!group2", "&group3"})
	assert.Len(include, 1)
	assert.Len(exclude, 1)
	assert.Len(intersect, 1)
	assert.True(include["group1"])
	assert.True(exclude["group2"])
	assert.True(intersect["group3"])
}

func TestPlaybookLabel_Basic(t *testing.T) {
	run(t, data, []string{"host1"}, []string{"host1"})
	run(t, data, []string{"host2"}, []string{"host2"})
	run(t, data, []string{"host3"}, []string{"host3"})
	run(t, data, []string{"group1"}, []string{"host1", "host2"})
	run(t, data, []string{"group2"}, []string{"host2", "host3"})
	run(t, data, []string{"all"}, []string{"host1", "host2", "host3"})
}

func TestPlaybookLabel_EmptyLabel(t *testing.T) {
	run(t, data, []string{}, []string{"host1", "host2", "host3"})
	run(t, data, []string{"", ""}, []string{"host1", "host2", "host3"})
}

func TestPlaybookLabel_MultiPattern(t *testing.T) {
	run(t, data, []string{"host1", "host2"}, []string{"host1", "host2"})
	run(t, data, []string{"host1", "host3"}, []string{"host1", "host3"})
	run(t, data, []string{"group1", "group2"}, []string{"host1", "host2", "host3"})
}

func TestPlaybookLabel_ExcludePattern(t *testing.T) {
	run(t, data, []string{"group1", "!host1"}, []string{"host2"})
	run(t, data, []string{"group1", "!host2"}, []string{"host1"})
	run(t, data, []string{"group2", "!host1"}, []string{"host2", "host3"})
	run(t, data, []string{"group2", "!host2"}, []string{"host3"})
	run(t, data, []string{"group2", "!host3"}, []string{"host2"})
	run(t, data, []string{"!host1"}, []string{"host2", "host3"})
	run(t, data, []string{"!group1"}, []string{"host3"})
	run(t, data, []string{"!all"}, []string{})
}

func TestPlaybookLabel_IntersectPattern(t *testing.T) {
	run(t, data, []string{"group1", "&host1"}, []string{"host1"})
	run(t, data, []string{"group1", "&host2"}, []string{"host2"})
	run(t, data, []string{"group1", "&host3"}, []string{})
	run(t, data, []string{"all", "&group1"}, []string{"host1", "host2"})
	run(t, data, []string{"all", "&group2"}, []string{"host2", "host3"})
	run(t, data, []string{"all", "&group1", "&group2"}, []string{"host2"})
	run(t, data, []string{"&group1", "&group2"}, []string{"host2"})
	run(t, data, []string{"host2", "&group1", "host3"}, []string{"host2"})
}

func TestPlaybookLabel_MixPattern(t *testing.T) {
	run(t, data, []string{"&group1", "&group2"}, []string{"host2"})
	run(t, data, []string{"&group1", "&group2", "!host2"}, []string{})
}
