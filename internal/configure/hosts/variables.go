package hosts

import (
	"fmt"
	"github.com/opencurve/curveadm/internal/errno"
	"github.com/opencurve/curveadm/pkg/variable"
)

type Var struct {
	name     string
	resolved bool
}

var (
	hostVars = []Var{
		{name: "instances_sequence"},
	}
)

func addVariables(hcs []*HostConfig, idx int, vars []Var) error {
	hc := hcs[idx]
	for _, v := range vars {
		err := hc.GetVariables().Register(variable.Variable{
			Name:  v.name,
			Value: getValue(v.name, hcs, idx),
		})
		if err != nil {
			return errno.ERR_REGISTER_VARIABLE_FAILED.E(err)
		}
	}

	return nil
}

func AddHostVariables(hcs []*HostConfig, idx int) error {
	return addVariables(hcs, idx, hostVars)
}

func getValue(name string, hcs []*HostConfig, idx int) string {
	hc := hcs[idx]
	switch name {
	case "instances_sequence":
		return fmt.Sprintf("%02d", hc.GetInstancesSequence())
	}
	return ""
}
