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
 * Created Date: 2022-07-08
 * Author: Jingli Chen (Wine93)
 */

package configure

import "fmt"

const (
	PATH_FSTAB = "/etc/fstab"

	FORMAT_PATH_DEVICE_SCHEDULER  = "/sys/block/%s/queue/scheduler"
	FORMAT_PATH_DEVICE_ROTATIONAL = "/sys/block/%s/queue/rotational"
)

func GetFSTabPath() string { return PATH_FSTAB }

func GetDeviceShedulerPath(device string) string {
	return fmt.Sprintf(FORMAT_PATH_DEVICE_SCHEDULER, device)
}

func GetDeviceRotationalPath(device string) string {
	return fmt.Sprintf(FORMAT_PATH_DEVICE_ROTATIONAL, device)
}
