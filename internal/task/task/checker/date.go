/*
 *  Copyright (c) 2022 NetEase Inc.
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 */

/*
 * Project: CurveAdm
 * Created Date: 2022-07-14
 * Author: Jingli Chen (Wine93)
 */

package checker

import (
	"fmt"
	"strconv"
	"time"

	"github.com/opencurve/curveadm/cli/cli"
	comm "github.com/opencurve/curveadm/internal/common"
	"github.com/opencurve/curveadm/internal/configure/topology"
	"github.com/opencurve/curveadm/internal/errno"
	"github.com/opencurve/curveadm/internal/task/context"
	"github.com/opencurve/curveadm/internal/task/step"
	"github.com/opencurve/curveadm/internal/task/task"
)

const (
	MAX_TIME_DIFFERENCE = 15
)

type Time struct {
	host string
	time int64
}

func step2Pre(start *int64) step.LambdaType {
	return func(ctx *context.Context) error {
		*start = time.Now().Unix()
		return nil
	}
}

func newIfNil(curveadm *cli.CurveAdm) map[string]Time {
	m := curveadm.MemStorage().Get(comm.KEY_ALL_HOST_DATE)
	if m != nil {
		return m.(map[string]Time)
	}
	return map[string]Time{}
}

func step2Post(curveadm *cli.CurveAdm, dc *topology.DeployConfig, start *int64, out *string) step.LambdaType {
	return func(ctx *context.Context) error {
		if len(*out) == 0 {
			return errno.ERR_INVALID_DATE_FORMAT.
				S("date is empty")
		}

		time, err := strconv.Atoi(*out)
		if err != nil {
			return errno.ERR_INVALID_DATE_FORMAT.
				F("date: %s", *out)
		}

		m := newIfNil(curveadm)
		m[dc.GetHost()] = Time{dc.GetHost(), int64(time)}
		curveadm.MemStorage().Set(comm.KEY_ALL_HOST_DATE, m)
		return nil
	}
}

func NewGetHostDate(curveadm *cli.CurveAdm, dc *topology.DeployConfig) (*task.Task, error) {
	hc, err := curveadm.GetHost(dc.GetHost())
	if err != nil {
		return nil, err
	}

	subname := fmt.Sprintf("host=%s start=%d", dc.GetHost(), time.Now().Unix())
	t := task.NewTask("Get Host Date <date>", subname, hc.GetConnectConfig())

	var start int64
	var out string
	t.AddStep(&step.Lambda{
		Lambda: step2Pre(&start),
	})
	t.AddStep(&step.Date{
		Format:      "+%s",
		Out:         &out,
		ExecOptions: curveadm.ExecOptions(),
	})
	t.AddStep(&step.Lambda{
		Lambda: step2Post(curveadm, dc, &start, &out),
	})

	return t, nil
}

func checkDate(curveadm *cli.CurveAdm) step.LambdaType {
	return func(ctx *context.Context) error {
		var minT, maxT Time
		min, max := int64(0), int64(0)
		m := newIfNil(curveadm)
		for _, t := range m {
			if min == 0 || t.time < min {
				min = t.time
				minT = t
			}
			if max == 0 || t.time > max {
				max = t.time
				maxT = t
			}
		}

		if max-min > MAX_TIME_DIFFERENCE {
			return errno.ERR_HOST_TIME_DIFFERENCE_OVER_30_SECONDS.
				F("difference=%d %s(%d) %s(%d)",
					max-min, maxT.host, maxT.time, minT.host, minT.time)
		}
		return nil
	}
}

func NewCheckDate(curveadm *cli.CurveAdm, c interface{}) (*task.Task, error) {
	t := task.NewTask("Check Host Date <date>", "", nil)
	t.AddStep(&step.Lambda{
		Lambda: checkDate(curveadm),
	})
	return t, nil
}

// TODO(P0): client time < service time
