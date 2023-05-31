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

// __SIGN_BY_WINE93__

package os

const (
	PATH_FSTAB      = "/etc/fstab"
	PATH_OS_RELEASE = "/etc/os-release"
	MAX_PORT        = 65535
)

func GetFSTabPath() string     { return PATH_FSTAB }
func GetMaxPortNum() int       { return MAX_PORT }
func GetOSReleasePath() string { return PATH_OS_RELEASE }
