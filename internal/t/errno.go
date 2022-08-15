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
 * Created Date: 2022-07-15
 * Author: Jingli Chen (Wine93)
 */

package errno

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	tui "github.com/opencurve/curveadm/internal/tui/common"
)

const (
	CODE_CANCEL_OPERATION = 900000
)

type ErrorCode struct {
	code        int
	description string
	clue        string
}

var (
	gLogpath string

	elist []*ErrorCode
)

func Init(logpath string) {
	gLogpath = logpath
}

func List() error {
	count := map[int]int{}
	for _, e := range elist {
		fmt.Printf(color.GreenString("%06d ", e.code))
		fmt.Println(color.YellowString("%s", e.description))
		count[e.code]++
	}
	fmt.Println(strings.Repeat("-", 3))
	fmt.Printf("%d error codes\n", len(count))

	for code, num := range count {
		if num > 1 {
			fmt.Println(color.RedString("duplicate code: %d", code))
		}
	}
	return nil
}

func EC(code int, description string) *ErrorCode {
	e := &ErrorCode{
		code:        code,
		description: description,
	}
	elist = append(elist, e)
	return e
}

func (e *ErrorCode) GetCode() int {
	return e.code
}

func (e *ErrorCode) GetClue() string {
	return e.clue
}

// added clue for error code
func (e *ErrorCode) E(err error) *ErrorCode {
	e.clue = err.Error()
	return e
}

func (e *ErrorCode) S(clue string) *ErrorCode {
	e.clue = clue
	return e
}

func (e *ErrorCode) F(format string, a ...interface{}) *ErrorCode {
	e.clue = fmt.Sprintf(format, a...)
	return e
}

func (e *ErrorCode) Error() string {
	if e.code == CODE_CANCEL_OPERATION {
		return ""
	}
	return tui.PromptErrorCode(e.code, e.description, e.clue, gLogpath)
}

/*
 * 0xx: init curveadm
 *
 * 1xx: database/SQL
 *   100: init failed
 *   11*: execute SQL statement
 *     * 110: hosts table
 *     * 111: clusters table
 *     * 112: containers table
 *     * 113: clients table
 *     * 114: plauground table
 *     * 115: audit table
 *
 * 2xx: command options
 *   20*: hosts
 *   21*: cluster
 *   22*: client/bs
 *   23*: client/fs
 *
 * 3xx: configure (curveadm.cfg, hosts.yaml, topology.yaml, format.yaml...)
 *   30*: curvreadm.cfg
 *     * 300: parse failed
 *     * 301: invalid configure value
 *   31*: hosts.yaml
 *     * 310: parse failed
 *     * 311: invalid configure value
 *   32*: topology.yaml
 *     * 320: parse failed
 *     * 321: invalid configure value
 *     * 322: update topology
 *   33*: format.yaml
 *     * 330: parse failed
 *     * 331: invalid configure value
 *   34*: client.yaml
 *   35*: plugin
 *
 * 4xx: common
 *   40*: hosts
 *   41*: services command
 *   42*: curvebs client
 *   43*: curvefs client
 *
 * 5xx: checker
 *   50*: topology
 *     500: s3
 *     501: directory
 *     502: listen address
 *     503: service
 *   51*: ssh
 *   52*: permission
 *   53*: kernel
 *   54*: network
 *   55*: date
 *   56*: service
 *   59*: others
 *
 * 6xx: execute task
 *  60*: ssh
 *  61*: shell command
 *  62*: docker command
 *  63*: file command
 *  69*: others
 *
 * 9xx: others
 */
var (
	// 000: init curveadm
	ERR_GET_USER_HOME_DIR_FAILED            = EC(000001, "get user home dir failed")
	ERR_CREATE_CURVEADM_SUBDIRECTORY_FAILED = EC(000002, "create curveadm subdirectory failed")
	ERR_INIT_LOGGER_FAILED                  = EC(000003, "init logger failed")

	// 100: database/SQL (init failed)
	ERR_INIT_SQL_DATABASE_FAILED = EC(100000, "init SQLite database failed")

	// 110: database/SQL (execute SQL statement: hosts table)
	ERR_GET_HOSTS_FAILED    = EC(110000, "execute SQL failed which get hosts")
	ERR_UPDATE_HOSTS_FAILED = EC(110001, "execute SQL failed which update hosts")
	// 111: database/SQL (execute SQL statement: clusters table)
	ERR_INSERT_CLUSTER_FAILED      = EC(111000, "execute SQL failed which insert cluster")
	ERR_GET_CLUSTER_BY_NAME_FAILED = EC(111001, "execute SQL failed which get cluster by name")
	ERR_CHECKOUT_CLUSTER_FAILED    = EC(111002, "execute SQL failed which checkout cluster")
	ERR_DELETE_CLUSTER_FAILED      = EC(111003, "execute SQL failed which delete cluster")
	ERR_GET_CURRENT_CLUSTER_FAILED = EC(111004, "execute SQL failed which get current cluster")
	ERR_GET_ALL_CLUSTERS_FAILED    = EC(111005, "execute SQL failed which get all clusters")
	ERR_UPDATE_CLUSTER_TOPOLOGY    = EC(111006, "execute SQL failed which update cluster topology")
	// 112: database/SQL (execute SQL statement: containers table)
	ERR_INSERT_SERVICE_CONTAINER_ID_FAILED   = EC(112000, "execute SQL failed which insert service container id")
	ERR_GET_SERVICE_CONTAINER_ID_FAILED      = EC(112001, "execute SQL failed which get service container id")
	ERR_GET_ALL_SERVICES_CONTAINER_ID_FAILED = EC(112002, "execute SQL failed which get all services container id")
	// 113: database/SQL (execute SQL statement: clients table)
	ERR_GET_ALL_CLIENTS_FAILED = EC(113000, "execute SQL failed which get all clients")
	// 114: database/SQL (execute SQL statement: playground table)
	// 115: database/SQL (execute SQL statement: audit table)
	ERR_GET_AUDIT_LOGS_FAILE = EC(115000, "execute SQL failed which get audit logs")

	// 200: command options (hosts)

	// 210: command options (cluster)
	ERR_UNSUPPORT_SKIPPED_SERVICE_ROLE = EC(210000, "unsupport skipped service role")
	ERR_UNSUPPORT_SKIPPED_CHECK_ITEM   = EC(210001, "unsupport skipped check item")
	ERR_UNSUPPORT_CLEAN_ITEM           = EC(210002, "unsupport clean item")
	ERR_NO_SERVICES_MATCHED            = EC(210003, "no services matched")

	// 220: command options (client/bs)
	ERR_INVALID_VOLUME_FORMAT                    = EC(220000, "invalid volume format")
	ERR_ROOT_VOLUME_USER_NOT_ALLOWED             = EC(220001, "root as volume user is not allowed")
	ERR_VOLUME_NAME_MUST_START_WITH_SLASH_PREFIX = EC(220002, "volume name must start with \"/\" prefix")
	ERR_VOLUME_SIZE_MUST_END_WITH_GB_SUFFIX      = EC(220003, "volume size must end with \"GB\" suffix")
	ERR_VOLUME_SIZE_REQUIRES_POSITIVE_INTEGER    = EC(220004, "volume size requires a positive integer")
	ERR_VOLUME_SIZE_MUST_BE_MULTIPLE_OF_10_GB    = EC(220005, "volume size must be a multiple of 10GB, like 10GB, 20GB, 30GB...")
	ERR_CLIENT_CONFIGURE_FILE_NOT_EXIST          = EC(220006, "client configure file not exist")

	// 230: command options (client/fs)

	// 300: configure (curveadm.cfg: parse failed)
	ERR_PARSE_CURVRADM_CONFIGURE_FAILED = EC(300000, "parse curveadm configure failed")
	// 301: configure (curveadm.cfg: invalid configure value)
	ERR_CURVEADM_CONFIGURE_VALUE_REQUIRES_STRING           = EC(301000, "curveadm configure value requires string")
	ERR_CURVEADM_CONFIGURE_VALUE_REQUIRES_INTEGER          = EC(301001, "curveadm configure value requires integer")
	ERR_CURVEADM_CONFIGURE_VALUE_REQUIRES_POSITIVE_INTEGER = EC(301002, "curveadm configure value requires positive integer")
	ERR_UNSUPPORT_LOG_LEVEL                                = EC(301003, "unsupport log level")

	// 310: configure (hosts.yaml: parse failed)
	ERR_HOSTS_FILE_NOT_FOUND   = EC(310000, "hosts file not found")
	ERR_READ_HOSTS_FILE_FAILED = EC(310001, "read hosts file failed")
	ERR_EMPTY_HOSTS            = EC(310002, "hosts is empty")
	ERR_PARSE_HOSTS_FAILED     = EC(310003, "parse hosts failed")
	// 311: configure (hosts.yaml: invalid configure value)
	ERR_HOST_MISSED                                     = EC(311000, "host missed")
	ERR_HOSTNAME_MISSED                                 = EC(311001, "hostname missed")
	ERR_HOSTS_CONFIGURE_VALUE_REQUIRES_STRING           = EC(301002, "hosts configure value requires string")
	ERR_HOSTS_CONFIGURE_VALUE_REQUIRES_INTEGER          = EC(311003, "hosts configure value requires integer")
	ERR_HOSTS_CONFIGURE_VALUE_REQUIRES_POSITIVE_INTEGER = EC(311004, "hosts configure value requires positive integer")
	ERR_DUPLICATE_HOST                                  = EC(311005, "host is duplicate")

	// 320: configure (topology.yaml: parse failed)
	ERR_TOPOLOGY_FILE_NOT_FOUND         = EC(320000, "topology file not found")
	ERR_READ_TOPOLOGY_FILE_FAILED       = EC(320001, "read topology file failed")
	ERR_EMPTY_CLUSTER_TOPOLOGY          = EC(320002, "cluster topology is empty")
	ERR_PARSE_TOPOLOGY_FAILED           = EC(320003, "parse topology failed")
	ERR_REGISTER_VARIABLE_FAILED        = EC(320004, "register variable failed")
	ERR_RESOLVE_VARIABLE_FAILED         = EC(320005, "resolve variable failed")
	ERR_RENDERING_VARIABLE_FAILED       = EC(320006, "rendering variable failed")
	ERR_CREATE_HASH_FOR_TOPOLOGY_FAILED = EC(320007, "create hash for topology failed")
	// 321: configure (topology.yaml: invalid configure value)
	ERR_UNSUPPORT_CLUSTER_KIND                    = EC(321000, "unsupport cluster kind")
	ERR_NO_SERVICES_IN_TOPOLOGY                   = EC(321001, "no services in topology")
	ERR_REPLICAS_REQUIRES_POSITIVE_INTEGER        = EC(321002, "replicas requires a positive integer")
	ERR_INVALID_VARIABLE_SECTION                  = EC(321003, "invalid variable section")
	ERR_UNSUPPORT_CONFIGURE_VALUE_TYPE            = EC(321004, "unsupport configure value type")
	ERR_CONFIGURE_VALUE_REQUIRES_BOOL             = EC(321005, "configure value requires bool")
	ERR_CONFIGURE_VALUE_REQUIRES_INTEGER          = EC(321006, "configure value requires integer")
	ERR_CONFIGURE_VALUE_REQUIRES_NON_EMPTY_STRING = EC(321007, "configure value requires non-empty string")
	ERR_CONFIGURE_VALUE_REQUIRES_POSITIVE_INTEGER = EC(321008, "configure value requires positive integer")
	ERR_DUPLICATE_SERVICE_ID                      = EC(321009, "service id is duplicate")
	// 322: configure (topology.yaml: update topology)
	ERR_DELETE_SERVICE_WHILE_COMMIT_TOPOLOGY_IS_DENIED = EC(322000, "delete service while commit topology is denied")
	ERR_ADD_SERVICE_WHILE_COMMIT_TOPOLOGY_IS_DENIED    = EC(322001, "add service while commit topology is denied")

	// 330: configure (topology.yaml: parse failed)
	ERR_FORMAT_CONFIGURE_FILE_NOT_EXIST = EC(330000, "format configure file not exits")
	ERR_PARSE_FORMAT_CONFIGURE_FAILED   = EC(330001, "parse format configure failed")
	// 331: configure (topology.yaml: invalid configure value)
	ERR_INVALID_DISK_FORMAT                      = EC(331000, "invalid disk format")
	ERR_INVALID_DEVICE                           = EC(331001, "invalid device")
	ERR_MOUNT_POINT_REQUIRE_ABSOLUTE_PATH        = EC(331002, "mount point must be an absolute path")
	ERR_FORMAT_PERCENT_REQUIRES_INTERGET         = EC(331003, "format percentage requires an integer")
	ERR_FORMAT_PERCENT_MUST_BE_BETWEEN_1_AND_100 = EC(301004, "format percentage must be between 1 and 100")

	// 340: configure (client.yaml: parse failed)
	ERR_PARSE_CLIENT_CONFIGURE_FAILED  = EC(340000, "parse client configure failed")
	ERR_RESOLVE_CLIENT_VARIABLE_FAILED = EC(340001, "resolve client variable failed")

	// 341: configure (client.yaml: invalid configure value)
	ERR_UNSUPPORT_CLIENT_KIND                 = EC(341000, "unsupport client kind")
	ERR_UNSUPPORT_CLIENT_CONFIGURE_VALUE_TYPE = EC(341001, "unsupport client configure value type")
	ERR_INVALID_CLUSTER_LISTEN_MDS_ADDRESS    = EC(341002, "invalid cluster MDS listen address")

	// 400: common (hosts)
	ERR_HOST_NOT_FOUND = EC(400000, "host not found")

	// 410: common (services command)
	ERR_NO_CLUSTER_SPECIFIED            = EC(410001, "no cluster specified")
	ERR_NO_SERVICES_FOR_PRECHECK        = EC(410003, "no services for precheck")
	ERR_NO_SERVICES_FOR_DEPLOY          = EC(410004, "no services for deploy")
	ERR_SERVICE_CONTAINER_ID_NOT_FOUND  = EC(410005, "service container id not found")
	ERR_CLUSTER_ALREADY_EXIST           = EC(410006, "cluster already exist")
	ERR_CLUSTER_NOT_FOUND               = EC(410007, "cluster not found")
	ERR_ACTIVE_SERVICE_IN_CLUSTER       = EC(410008, "please clean all services before remove cluster")
	ERR_UNSUPPORT_CONFIG_TYPE           = EC(410009, "unsupport config type")
	ERR_UNKNOWN_TASK_TYPE               = EC(410010, "unknown task type")
	ERR_DECODE_CLUSTER_POOL_JSON_FAILED = EC(410011, "decode cluster pool json to string failed")

	// 420: common (curvebs client)
	ERR_VOLUME_ALREADY_MAPPED     = EC(420000, "volume already mapped")
	ERR_VOLUME_CONTAINER_LOSED    = EC(420001, "volume container is losed")
	ERR_VOLUME_CONTAINER_ABNORMAL = EC(420002, "volume container is abnormal")
	ERR_CREATE_VOLUME_FAILED      = EC(420003, "create volume failed")
	ERR_MAP_VOLUME_FAILED         = EC(420004, "map volume to NBD device failed")

	// 430: common (curvebs client)
	ERR_FS_PATH_ALREADY_MOUNTED             = EC(430000, "path already mounted")
	ERR_FS_MOUNTPOINT_REQUIRE_ABSOLUTE_PATH = EC(430001, "mount point must be an absolute path")

	// 500: checker (topology/s3)
	ERR_INVALID_S3_ACCESS_KEY  = EC(500000, "invalid S3 access key")
	ERR_INVALID_S3_SECRET_KEY  = EC(500001, "invalid S3 secret key")
	ERR_INVALID_S3_ADDRESS     = EC(500002, "invalid S3 address")
	ERR_INVALID_S3_BUCKET_NAME = EC(500003, "invalid S3 bucket name")
	// 501: checker (topology/directory)
	ERR_DIRECTORY_REQUIRE_ABSOLUTE_PATH = EC(501000, "directory must be an absolute path")
	ERR_DATA_DIRECTORY_ALREADY_IN_USE   = EC(501001, "data directory already in use")
	// 502: checker (topology/address)
	ERR_DUPLICATE_LISTEN_ADDRESS = EC(502000, "listen address is duplicate")
	// 503: checker (topology/service)
	ERR_ETCD_REQUIRES_3_SERVICES          = EC(503000, "etcd requires at least 3 services")
	ERR_MDS_REQUIRES_3_SERVICES           = EC(503001, "mds requires at least 3 services")
	ERR_CHUNKSERVER_REQUIRES_3_SERVICES   = EC(503002, "chunkserver requires at least 3 services")
	ERR_SNAPSHOTCLONE_REQUIRES_3_SERVICES = EC(503003, "snapshotclone requires at least 3 services")
	ERR_METASERVER_REQUIRES_3_SERVICES    = EC(503004, "metaserver requires at least 3 services")
	ERR_CHUNKSERVER_REQUIRES_3_HOSTS      = EC(503005, "chunkserver requires at least 3 hosts to distrubute zones")
	ERR_METASERVER_REQUIRES_3_HOSTS       = EC(503006, "metaserver requires at least 3 hosts to distrubute zones")

	// 510: checker (ssh)
	ERR_PRIVATE_KEY_FILE_REQUIRE_ABSOLUTE_PATH   = EC(510000, "SSH private key file needs to be an absolute path")
	ERR_PRIVATE_KEY_FILE_NOT_EXIST               = EC(510001, "SSH private key file not exist")
	ERR_PRIVATE_KEY_FILE_REQUIRE_600_PERMISSIONS = EC(510002, "SSH private key file require 600 permissions")
	ERR_SSH_CONNECT_FAILED                       = EC(510003, "SSH connect failed")

	// 520: checker (permission)
	ERR_CREATE_DIRECOTRY_PERMISSION_DENIED       = EC(520000, "create direcotry permission denied")
	ERR_EXECUTE_DOCKER_COMMAND_PERMISSION_DENIED = EC(520001, "execute docker command permission denied")

	// 530: checker (kernel)
	ERR_UNRECOGNIZED_KERNEL_VERSION              = EC(530000, "unrecognized kernel version")
	ERR_RENAMEAT_NOT_SUPPORTED_IN_CURRENT_KERNEL = EC(530001, "renameat() not supported in current kernel version")

	// 540: checker (network)
	ERR_PORT_ALREADY_IN_USE = EC(540000, "port is already in use")

	// 550: checker (date)
	ERR_INVALID_DATE_FORMAT                  = EC(550000, "invalid date format")
	ERR_HOST_TIME_DIFFERENCE_OVER_30_SECONDS = EC(550001, "host time difference over 30 seconds")

	// 560: checker (service)
	ERR_CHUNKFILE_POOL_NOT_EXIST = EC(560000, "there is no chunkfile pool in data directory")

	// 590: checker (others)
	ERR_DOCKER_NOT_INSTALLED         = EC(590000, "docker not installed")
	ERR_DOCKER_DAEMON_IS_NOT_RUNNING = EC(590001, "docker daemon is not running")
	ERR_NO_SPACE_LEFT_ON_DEVICE      = EC(590002, "no space left on device")

	// 610: execute task (shell command)
	ERR_EDIT_FILE_FAILED                      = EC(610000, "edit file failed (sed)")
	ERR_LIST_DIRECTORY_CONTENTS_FAILED        = EC(610001, "list directory contents failed (ls)")
	ERR_CREATE_DIRECTORY_FAILED               = EC(610002, "create directory failed (mkdir)")
	ERR_REMOVE_FILES_OR_DIRECTORIES_FAILED    = EC(610003, "remove files or directories failed (rm)")
	ERR_COPY_FILES_AND_DIRECTORIES_FAILED     = EC(610004, "copy files and directories failed (cp)")
	ERR_BUILD_A_LINUX_FILE_SYSTEM_FAILED      = EC(610005, "build a Linux file system failed (mkfs)")
	ERR_MOUNT_A_FILESYSTEM_FAILED             = EC(610006, "mount a filesystem failed (mount)")
	ERR_UNMOUNT_FILE_SYSTEMS_FAILED           = EC(610007, "unmount file systems failed (umount)")
	ERR_FIND_WHICH_PROCESS_USING_FILE_FAILED  = EC(610008, "find which process is using file failed (fuser)")
	ERR_GET_DISK_SPACE_USAGE_FAILED           = EC(610009, "get disk space usage failed (df)")
	ERR_LIST_BLOCK_DEVICES_FAILED             = EC(610010, "list block devices failed (lsblk)")
	ERR_GET_FILE_STATUS_FAILED                = EC(610011, "get file status failed (stat)")
	ERR_GET_CONNECTION_INFORMATION_FAILED     = EC(610012, "get connection information failed (ss)")
	ERR_GET_SYSTEM_INFORMATION_FAILED         = EC(610013, "get system information failed (uname)")
	ERR_GET_SYSTEM_TIME_FAILED                = EC(610014, "get system time failed (date)")
	ERR_SEND_ICMP_ECHO_REQUEST_TO_HOST_FAILED = EC(610015, "send ICMP ECHO_REQUEST to host failed (ping)")
	ERR_RUN_A_BASH_COMMAND_FAILED             = EC(610999, "run a bash command failed (bash -c)")

	// 620: execute task (docker command)
	ERR_GET_DOCKER_INFO_FAILED          = EC(620000, "get docker info failed")
	ERR_PULL_IMAGE_FAILED               = EC(620001, "pull image failed")
	ERR_CREATE_CONTAINER_FAILED         = EC(620002, "create container failed")
	ERR_START_CONTAINER_FAILED          = EC(620003, "start container failed")
	ERR_STOP_CONTAINER_FAILED           = EC(620004, "stop container failed")
	ERR_RESTART_CONTAINER_FAILED        = EC(620005, "restart container failed")
	ERR_WAIT_CONTAINER_STOP_FAILED      = EC(620006, "wait container stop failed")
	ERR_REMOVE_CONTAINER_FAILED         = EC(620007, "remove container failed")
	ERR_LIST_CONTAINERS_FAILED          = EC(620008, "list containers failed")
	ERR_RUN_COMMAND_IN_CONTAINER_FAILED = EC(620009, "run a command in container failed")
	ERR_COPY_FROM_CONTAINER_FAILED      = EC(620010, "copy file from container failed")
	ERR_COPY_INTO_CONTAINER_FAILED      = EC(620011, "copy file into container failed")
	ERR_INSPECT_CONTAINER_FAILED        = EC(620012, "get container low-level information failed")

	// 690: execuetr task (others)
	ERR_SERVICE_IS_ABNORMAL = EC(690000, "service is abnormal")

	// 900: others
	ERR_CANCEL_OPERATION = EC(CODE_CANCEL_OPERATION, "cancel operation")
	// 999
	ERR_UNKNOWN = EC(999999, "unknown error")
)


// 	ERR_PRIVATE_KEY_FILE_REQUIRE_ABSOLUTE_PATH
// 	ERR_PRIVATE_KEY_FILE_NOT_EXIST
// 	ERR_PRIVATE_KEY_FILE_REQUIRE_600_PERMISSIONS
// 	ERR_SSH_CONNECT_FAILED
//
//	// 520: checker (permission)
//	ERR_CREATE_DIRECOTRY_PERMISSION_DENIED
//	ERR_EXECUTE_DOCKER_COMMAND_PERMISSION_DENIED
//
//	// 530: checker (kernel)
//	ERR_UNRECOGNIZED_KERNEL_VERSION
//	ERR_RENAMEAT_NOT_SUPPORTED_IN_CURRENT_KERNEL
//
//	// 540: checker (network)
//	ERR_PORT_ALREADY_IN_USE
//
//	// 550: checker (date)
//	ERR_INVALID_DATE_FORMAT                  = EC(550000, "invalid date format")
//	ERR_HOST_TIME_DIFFERENCE_OVER_30_SECONDS = EC(550001, "host time difference over 30 seconds")
//
//	// 560: checker (service)
//	ERR_CHUNKFILE_POOL_NOT_EXIST = EC(560000, "there is no chunkfile pool in data directory")
//
//	// 590: checker (others)
//	ERR_DOCKER_NOT_INSTALLED         = EC(590000, "docker not installed")
//	ERR_DOCKER_DAEMON_IS_NOT_RUNNING = EC(590001, "docker daemon is not running")
//	ERR_NO_SPACE_LEFT_ON_DEVICE      = EC(590002, "no space left on device")
//
//	// 610: execute task (shell command)
//	ERR_EDIT_FILE_FAILED                     
//	ERR_LIST_DIRECTORY_CONTENTS_FAILED       
//	ERR_CREATE_DIRECTORY_FAILED              
//	ERR_REMOVE_FILES_OR_DIRECTORIES_FAILED   
//	ERR_COPY_FILES_AND_DIRECTORIES_FAILED    
//	ERR_BUILD_A_LINUX_FILE_SYSTEM_FAILED     
//	ERR_MOUNT_A_FILESYSTEM_FAILED            
//	ERR_UNMOUNT_FILE_SYSTEMS_FAILED          
//	ERR_FIND_WHICH_PROCESS_USING_FILE_FAILED 
//	ERR_GET_DISK_SPACE_USAGE_FAILED          
//	ERR_LIST_BLOCK_DEVICES_FAILED            
//	ERR_GET_FILE_STATUS_FAILED               
//	ERR_GET_CONNECTION_INFORMATION_FAILED    
//	ERR_GET_SYSTEM_INFORMATION_FAILED        
//	ERR_GET_SYSTEM_TIME_FAILED               
//	ERR_SEND_ICMP_ECHO_REQUEST_TO_HOST_FAILED
//	ERR_RUN_A_BASH_COMMAND_FAILED            
//	ERR_GET_DOCKER_INFO_FAILED           
//	ERR_PULL_IMAGE_FAILED                
//	ERR_CREATE_CONTAINER_FAILED          
//	ERR_START_CONTAINER_FAILED           
//	ERR_STOP_CONTAINER_FAILED            
//	ERR_RESTART_CONTAINER_FAILED         
//	ERR_WAIT_CONTAINER_STOP_FAILED       
//	ERR_REMOVE_CONTAINER_FAILED          
//	ERR_LIST_CONTAINERS_FAILED           
//	ERR_RUN_COMMAND_IN_CONTAINER_FAILED  
//	ERR_COPY_FROM_CONTAINER_FAILED       
//	ERR_COPY_INTO_CONTAINER_FAILED       
//	ERR_INSPECT_CONTAINER_FAILED         
//
//	// 690: execuetr task (others)
//	ERR_SERVICE_IS_ABNORMAL = EC(690000, "service is abnormal")
//
//	// 900: others
//	ERR_CANCEL_OPERATION = EC(CODE_CANCEL_OPERATION, "cancel operation")
//	// 999
//	ERR_UNKNOWN = EC(999999, "unknown error")

