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

// __SIGN_BY_WINE93__

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

func (e *ErrorCode) GetDescription() string {
	return e.description
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

func (e *ErrorCode) FD(format string, s ...interface{}) *ErrorCode {
	newEC := &ErrorCode{
		code:        e.code,
		description: e.description,
	}
	newEC.description = fmt.Sprintf(newEC.description+" "+format, s...)
	return newEC
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
 *     * 116: any table
 *
 * 2xx: command options
 *   20*: hosts
 *   21*: cluster
 *   22*: client
 *
 * 3xx: configure (curveadm.cfg, hosts.yaml, topology.yaml, format.yaml...)
 *   300: common
 *   31*: curvreadm.cfg
 *     * 310: parse failed
 *     * 311: invalid configure value
 *   32*: hosts.yaml
 *     * 320: parse failed
 *     * 321: invalid configure value
 *   33*: topology.yaml
 *     * 330: parse failed
 *     * 331: invalid configure value
 *     * 332: update topology
 *   34*: format.yaml
 *     * 340: parse failed
 *     * 341: invalid configure value
 *   35*: client.yaml
 *   36*: plugin
 *
 * 4xx: common
 *   40*: hosts
 *   41*: services command
 *   42*: curvebs client
 *   43*: curvefs client
 *   44*: polarfs
 *   45*: playground
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
 *   57*: client
 *   59*: others
 *
 * 6xx: execute task
 *  60*: common
 *  61*: ssh command
 *  62*: shell command
 *  63*: docker command
 *  64*: file command
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
	ERR_INSERT_CLUSTER_FAILED          = EC(111000, "execute SQL failed which insert cluster")
	ERR_GET_CURRENT_CLUSTER_FAILED     = EC(111001, "execute SQL failed which get current cluster")
	ERR_GET_CLUSTER_BY_NAME_FAILED     = EC(111002, "execute SQL failed which get cluster by name")
	ERR_GET_ALL_CLUSTERS_FAILED        = EC(111003, "execute SQL failed which get all clusters")
	ERR_CHECKOUT_CLUSTER_FAILED        = EC(111004, "execute SQL failed which checkout cluster")
	ERR_DELETE_CLUSTER_FAILED          = EC(111005, "execute SQL failed which delete cluster")
	ERR_UPDATE_CLUSTER_TOPOLOGY_FAILED = EC(111006, "execute SQL failed which update cluster topology")
	ERR_UPDATE_CLUSTER_POOL_FAILED     = EC(111007, "execute SQL failed which update cluster pool")
	// 112: database/SQL (execute SQL statement: containers table)
	ERR_INSERT_SERVICE_CONTAINER_ID_FAILED   = EC(112000, "execute SQL failed which insert service container id")
	ERR_SET_SERVICE_CONTAINER_ID_FAILED      = EC(112001, "execute SQL failed which set service container id")
	ERR_GET_SERVICE_CONTAINER_ID_FAILED      = EC(112002, "execute SQL failed which get service container id")
	ERR_GET_ALL_SERVICES_CONTAINER_ID_FAILED = EC(112003, "execute SQL failed which get all services container id")
	// 113: database/SQL (execute SQL statement: clients table)
	ERR_INSERT_CLIENT_FAILED           = EC(113000, "execute SQL failed which insert client")
	ERR_GET_CLIENT_CONTAINER_ID_FAILED = EC(113001, "execute SQL failed which get client container id")
	ERR_GET_CLIENT_BY_ID_FAILED        = EC(113002, "execute SQL failed which get client by id")
	ERR_GET_ALL_CLIENTS_FAILED         = EC(113003, "execute SQL failed which get all clients")
	ERR_DELETE_CLIENT_FAILED           = EC(113004, "execute SQL failed which delete client")
	ERR_SET_CLIENT_AUX_INFO_FAILED     = EC(113005, "execute SQL failed which set client aux info")
	// 114: database/SQL (execute SQL statement: playground table)
	ERR_INSERT_PLAYGROUND_FAILED      = EC(114000, "execute SQL failed which insert playground")
	ERR_GET_ALL_PLAYGROUND_FAILED     = EC(114001, "execute SQL failed which get all playgrounds")
	ERR_GET_PLAYGROUND_BY_NAME_FAILED = EC(114002, "execute SQL failed which get playground by name")
	ERR_DELETE_PLAYGROUND_FAILED      = EC(114003, "execute SQL failed which delete playground")
	// 115: database/SQL (execute SQL statement: audit table)
	ERR_GET_AUDIT_LOGS_FAILE = EC(115000, "execute SQL failed which get audit logs")
	// 116: database/SQL (execute SQL statement: any table)
	ERR_INSERT_CLIENT_CONFIG_FAILED = EC(116000, "execute SQL failed which insert client config")
	ERR_SELECT_CLIENT_CONFIG_FAILED = EC(116001, "execute SQL failed which select client config")
	ERR_DELETE_CLIENT_CONFIG_FAILED = EC(116002, "execute SQL failed which delete client config")
	// 117: database/SQL (execute SQL statement: monitor table)
	ERR_GET_MONITOR_FAILED     = EC(117000, "execute SQL failed while get monitor")
	ERR_REPLACE_MONITOR_FAILED = EC(117001, "execute SQL failed while replace monitor")
	ERR_UPDATE_MONITOR_FAILED  = EC(117002, "execute SQL failed while update monitor")

	// 200: command options (hosts)

	// 210: command options (cluster)
	ERR_ID_NOT_FOUND                   = EC(210000, "id not found")
	ERR_UNSUPPORT_CURVEBS_ROLE         = EC(210001, "unsupport curvebs role (etcd/mds/chunkserver/snapshotclone)")
	ERR_UNSUPPORT_CURVEFS_ROLE         = EC(210002, "unsupport curvefs role (etcd/mds/metaserver)")
	ERR_UNSUPPORT_SKIPPED_SERVICE_ROLE = EC(210003, "unsupport skipped service role")
	ERR_UNSUPPORT_SKIPPED_CHECK_ITEM   = EC(210004, "unsupport skipped check item")
	ERR_UNSUPPORT_CLEAN_ITEM           = EC(210005, "unsupport clean item")
	ERR_NO_SERVICES_MATCHED            = EC(210006, "no services matched")
	// TODO: please check pool set disk type
	ERR_INVALID_DISK_TYPE                   = EC(210007, "poolset disk type must be lowercase and can only be one of ssd, hdd and nvme")
	ERR_UNSUPPORT_DEPLOY_TYPE               = EC(210008, "unknown deploy type")
	ERR_NO_LEADER_OR_RANDOM_CONTAINER_FOUND = EC(210009, "no leader or random container found")
	
	// 220: commad options (client common)
	ERR_UNSUPPORT_CLIENT_KIND = EC(220000, "unsupport client kind")
	// 221: command options (client/bs)
	ERR_INVALID_VOLUME_FORMAT                      = EC(221000, "invalid volume format")
	ERR_ROOT_VOLUME_USER_NOT_ALLOWED               = EC(221001, "root as volume user is not allowed")
	ERR_VOLUME_NAME_MUST_START_WITH_SLASH_PREFIX   = EC(221002, "volume name must start with \"/\" prefix")
	ERR_VOLUME_SIZE_MUST_END_WITH_GiB_SUFFIX       = EC(221003, "volume size must end with \"GiB\" suffix")
	ERR_VOLUME_SIZE_REQUIRES_POSITIVE_INTEGER      = EC(221004, "volume size requires a positive integer")
	ERR_VOLUME_SIZE_MUST_BE_MULTIPLE_OF_10_GiB     = EC(221005, "volume size must be a multiple of 10GiB, like 10GiB, 20GiB, 30GiB...")
	ERR_CLIENT_CONFIGURE_FILE_NOT_EXIST            = EC(221006, "client configure file not exist")
	ERR_NO_CLIENT_MATCHED                          = EC(221007, "no client matched")
	ERR_VOLUME_NAME_CAN_NOT_CONTAIN_UNDERSCORE     = EC(221008, "volume name can't contain \"_\" symbol")
	ERR_VOLUME_BLOCKSIZE_MUST_END_WITH_BYTE_SUFFIX = EC(201009, "volume block size must end with \"B\" suffix")
	ERR_VOLUME_BLOCKSIZE_REQUIRES_POSITIVE_INTEGER = EC(221010, "volume block size requires a positive integer")
	ERR_VOLUME_BLOCKSIZE_BE_MULTIPLE_OF_512        = EC(221011, "volume block size be a multiple of 512B, like 1KiB, 2KiB, 3KiB...")
	// 222: command options (client/fs)
	ERR_FS_MOUNTPOINT_REQUIRE_ABSOLUTE_PATH = EC(222000, "mount point must be an absolute path")

	// 230: command options (playground)
	ERR_UNSUPPORT_PLAYGROUND_KIND                      = EC(230000, "unsupport playground kind")
	ERR_MUST_SPECIFY_MOUNTPOINT_FOR_CURVEFS_PLAYGROUND = EC(230001, "you must specify mountpoint for curvefs playground")
	ERR_PLAYGROUND_MOUNTPOINT_REQUIRE_ABSOLUTE_PATH    = EC(230002, "mount point must be an absolute path")
	ERR_PLAYGROUND_MOUNTPOINT_NOT_EXIST                = EC(230003, "mount point not exist")

	// 301: configure (common: invalid configure value)
	ERR_UNSUPPORT_CONFIGURE_VALUE_TYPE = EC(301000, "unsupport configure value type")
	// lose 301001
	ERR_CONFIGURE_VALUE_REQUIRES_BOOL             = EC(301002, "configure value requires bool")
	ERR_CONFIGURE_VALUE_REQUIRES_INTEGER          = EC(301003, "configure value requires integer")
	ERR_CONFIGURE_VALUE_REQUIRES_NON_EMPTY_STRING = EC(301004, "configure value requires non-empty string")
	ERR_CONFIGURE_VALUE_REQUIRES_POSITIVE_INTEGER = EC(301005, "configure value requires positive integer")
	ERR_CONFIGURE_VALUE_REQUIRES_STRING_SLICE     = EC(301006, "configure value requires string array")
	ERR_UNSUPPORT_VARIABLE_VALUE_TYPE             = EC(301100, "unsupport variable value type")
	ERR_INVALID_VARIABLE_VALUE                    = EC(301101, "invalid variable value")

	// 310: configure (curveadm.cfg: parse failed)
	ERR_PARSE_CURVRADM_CONFIGURE_FAILED = EC(310000, "parse curveadm configure failed")
	// 311: configure (curveadm.cfg: invalid configure value)
	ERR_UNSUPPORT_CURVEADM_LOG_LEVEL      = EC(311000, "unsupport curveadm log level")
	ERR_UNSUPPORT_CURVEADM_CONFIGURE_ITEM = EC(311001, "unsupport curveadm configure item")
	ERR_UNSUPPORT_CURVEADM_DATABASE_URL   = EC(311002, "unsupport curveadm database url")

	// 320: configure (hosts.yaml: parse failed)
	ERR_HOSTS_FILE_NOT_FOUND   = EC(320000, "hosts file not found")
	ERR_READ_HOSTS_FILE_FAILED = EC(320001, "read hosts file failed")
	ERR_EMPTY_HOSTS            = EC(320002, "hosts is empty")
	ERR_PARSE_HOSTS_FAILED     = EC(320003, "parse hosts failed")
	// 321: configure (hosts.yaml: invalid configure value)
	ERR_UNSUPPORT_HOSTS_CONFIGURE_ITEM           = EC(321000, "unsupport hosts configure item")
	ERR_HOST_FIELD_MISSING                       = EC(321001, "host field missing")
	ERR_HOSTNAME_FIELD_MISSING                   = EC(321002, "hostname field missing")
	ERR_HOSTS_SSH_PORT_EXCEED_MAX_PORT_NUMBER    = EC(321003, "ssh_port exceed max port number")
	ERR_PRIVATE_KEY_FILE_REQUIRE_ABSOLUTE_PATH   = EC(321004, "SSH private key file needs to be an absolute path")
	ERR_PRIVATE_KEY_FILE_NOT_EXIST               = EC(321005, "SSH private key file not exist")
	ERR_PRIVATE_KEY_FILE_REQUIRE_600_PERMISSIONS = EC(321006, "SSH private key file require 600 permissions")
	ERR_DUPLICATE_NAME                           = EC(321007, "name is duplicate")
	ERR_HOSTNAME_REQUIRES_VALID_IP_ADDRESS       = EC(321008, "hostname requires valid IP address")

	// 322: configure (monitor.yaml: parse failed)
	ERR_PARSE_MONITOR_CONFIGURE_FAILED   = EC(322000, "parse monitor configure failed")
	ERR_READ_MONITOR_FILE_FAILED         = EC(322001, "read monitor file failed")
	ERR_PARSE_PROMETHEUS_TARGET_FAILED   = EC(322002, "parse prometheus targets failed")
	ERR_PARSE_CURVE_MANAGER_CONF_FAILED  = EC(322003, "parse curve-manager configure failed")
	ERR_UPDATE_CURVE_MANAGER_CONF_FAILED = EC(322004, "update curve-manager configure failed")

	// 330: configure (topology.yaml: parse failed)
	ERR_TOPOLOGY_FILE_NOT_FOUND         = EC(330000, "topology file not found")
	ERR_READ_TOPOLOGY_FILE_FAILED       = EC(330001, "read topology file failed")
	ERR_EMPTY_CLUSTER_TOPOLOGY          = EC(330002, "cluster topology is empty")
	ERR_PARSE_TOPOLOGY_FAILED           = EC(330003, "parse topology failed")
	ERR_REGISTER_VARIABLE_FAILED        = EC(330004, "register variable failed")
	ERR_RESOLVE_VARIABLE_FAILED         = EC(330005, "resolve variable failed")
	ERR_SET_VARIABLE_VALUE_FAILED       = EC(330006, "set variable value failed")
	ERR_RENDERING_VARIABLE_FAILED       = EC(330007, "rendering variable failed")
	ERR_CREATE_HASH_FOR_TOPOLOGY_FAILED = EC(330008, "create hash for topology failed")
	// 331: configure (topology.yaml: invalid configure value)
	ERR_UNSUPPORT_CLUSTER_KIND              = EC(331000, "unsupport cluster kind")
	ERR_NO_SERVICES_IN_TOPOLOGY             = EC(331001, "no services in topology")
	ERR_INSTANCES_REQUIRES_POSITIVE_INTEGER = EC(331002, "instances requires a positive integer")
	ERR_INVALID_VARIABLE_SECTION            = EC(331003, "invalid variable section")
	ERR_DUPLICATE_SERVICE_ID                = EC(331004, "service id is duplicate")
	// 332: configure (topology.yaml: update topology)
	ERR_DELETE_SERVICE_WHILE_COMMIT_TOPOLOGY_IS_DENIED   = EC(332000, "delete service while commit topology is denied")
	ERR_ADD_SERVICE_WHILE_COMMIT_TOPOLOGY_IS_DENIED      = EC(332001, "add service while commit topology is denied")
	ERR_DELETE_SERVICE_WHILE_SCALE_OUT_CLUSTER_IS_DENIED = EC(332002, "delete service while scale out cluster is denied")
	ERR_NO_SERVICES_FOR_SCALE_OUT_CLUSTER                = EC(332003, "no service for scale out cluster")
	ERR_REQUIRE_SAME_ROLE_SERVICES_FOR_SCALE_OUT_CLUSTER = EC(332004, "require same role services for scale out cluster")
	ERR_CHUNKSERVER_REQUIRES_3_HOSTS_WHILE_SCALE_OUT     = EC(332005, "chunkserver requires at least 3 new hosts to distrubute zones while scale out")
	ERR_METASERVER_REQUIRES_3_HOSTS_WHILE_SCALE_OUT      = EC(332006, "metaserver requires at least 3 new hosts to distrubute zones while scale out")
	ERR_ADD_SERVICE_WHILE_MIGRATING_IS_DENIED            = EC(332007, "add service while migrating is denied")
	ERR_DELETE_SERVICE_WHILE_MIGRATING_IS_DENIED         = EC(332008, "delete service while migrating is denied")
	ERR_NO_SERVICES_FOR_MIGRATING                        = EC(332009, "no service for migrating")
	ERR_REQUIRE_SAME_ROLE_SERVICES_FOR_MIGRATING         = EC(332010, "require same role services for migrating")
	ERR_REQUIRE_WHOLE_HOST_SERVICES_FOR_MIGRATING        = EC(332011, "require whole host services for migrating")

	// 340: configure (format.yaml: parse failed)
	ERR_FORMAT_CONFIGURE_FILE_NOT_EXIST = EC(340000, "format configure file not exits")
	ERR_PARSE_FORMAT_CONFIGURE_FAILED   = EC(340001, "parse format configure failed")
	// 341: configure (format.yaml: invalid configure value)
	ERR_INVALID_DISK_FORMAT                      = EC(341000, "invalid disk format")
	ERR_INVALID_DEVICE                           = EC(341001, "invalid device")
	ERR_MOUNT_POINT_REQUIRE_ABSOLUTE_PATH        = EC(341002, "mount point must be an absolute path")
	ERR_FORMAT_PERCENT_REQUIRES_INTERGET         = EC(341003, "format percentage requires an integer")
	ERR_FORMAT_PERCENT_MUST_BE_BETWEEN_1_AND_100 = EC(341004, "format percentage must be between 1 and 100")
	ERR_INVALID_BLOCK_SIZE                       = EC(341005, "invalid block size, support 512,4096")

	// 350: configure (client.yaml: parse failed)
	ERR_PARSE_CLIENT_CONFIGURE_FAILED  = EC(350000, "parse client configure failed")
	ERR_RESOLVE_CLIENT_VARIABLE_FAILED = EC(350001, "resolve client variable failed")
	// 351: configure (client.yaml: invalid configure value)
	ERR_UNSUPPORT_CLIENT_CONFIGURE_KIND            = EC(351000, "unsupport client configure kind")
	ERR_UNSUPPORT_CLIENT_CONFIGURE_VALUE_TYPE      = EC(351001, "unsupport client configure value type")
	ERR_REQUIRE_CURVEBS_KIND_CLIENT_CONFIGURE_FILE = EC(351002, "require curvebs kind client configure file")
	ERR_REQUIRE_CURVEFS_KIND_CLIENT_CONFIGURE_FILE = EC(351003, "require curvefs kind client configure file")
	ERR_INVALID_CLUSTER_LISTEN_MDS_ADDRESS         = EC(351004, "invalid cluster MDS listen address")

	// 400: common (hosts)
	ERR_HOST_NOT_FOUND = EC(400000, "host not found")

	// 410: common (services command)
	ERR_NO_CLUSTER_SPECIFIED                 = EC(410001, "no cluster specified")
	ERR_NO_DISK_FOR_FORMATTING               = EC(410002, "no host/disk for formating")
	ERR_GET_DEVICE_UUID_FAILED               = EC(410003, "get device uuid failed")
	ERR_NOT_A_BLOCK_DEVICE                   = EC(410004, "not a block device")
	ERR_INVALID_UUID                         = EC(410005, "invalid device uuid")
	ERR_NO_SERVICES_FOR_PRECHECK             = EC(410006, "no services for precheck")
	ERR_NO_SERVICES_FOR_DEPLOY               = EC(410007, "no services for deploy")
	ERR_SERVICE_CONTAINER_ID_NOT_FOUND       = EC(410008, "service container id not found")
	ERR_CLUSTER_ALREADY_EXIST                = EC(410009, "cluster already exist")
	ERR_CLUSTER_NOT_FOUND                    = EC(410010, "cluster not found")
	ERR_ACTIVE_SERVICE_IN_CLUSTER            = EC(410011, "please clean all services before remove cluster")
	ERR_UNSUPPORT_CONFIG_TYPE                = EC(410012, "unsupport config type")
	ERR_UNKNOWN_TASK_TYPE                    = EC(410013, "unknown task type")
	ERR_CONTAINER_ALREADT_REMOVED            = EC(410014, "container already removed")
	ERR_CONTAINER_IS_ABNORMAL                = EC(410015, "container is abnormal")
	ERR_DECODE_CLUSTER_POOL_JSON_FAILED      = EC(410016, "decode cluster pool json to string failed")
	ERR_WAIT_MDS_ELECTION_SUCCESS_TIMEOUT    = EC(410017, "wait mds election success timeout")
	ERR_WAIT_ALL_CHUNKSERVERS_ONLINE_TIMEOUT = EC(410018, "wait all chunkservers online timeout")
	ERR_CREATE_LOGICAL_POOL_FAILED           = EC(410019, "create logical pool failed")
	ERR_INVALID_DEVICE_USAGE                 = EC(410020, "invalid device usage")
	ERR_ENCRYPT_FILE_FAILED                  = EC(410021, "encrypt file failed")
	ERR_CLIENT_ID_NOT_FOUND                  = EC(410022, "client id not found")
	ERR_ENABLE_ETCD_AUTH_FAILED              = EC(410023, "enable etcd auth failed")

	// 420: common (curvebs client)
	ERR_VOLUME_ALREADY_MAPPED             = EC(420000, "volume already mapped")
	ERR_VOLUME_CONTAINER_LOSED            = EC(420001, "volume container is losed")
	ERR_VOLUME_CONTAINER_ABNORMAL         = EC(420002, "volume container is abnormal")
	ERR_CREATE_VOLUME_FAILED              = EC(420003, "create volume failed")
	ERR_MAP_VOLUME_FAILED                 = EC(420004, "map volume to NBD device failed")
	ERR_ENCODE_VOLUME_INFO_TO_JSON_FAILED = EC(420005, "encode volume info to json failed")
	ERR_UNMAP_VOLUME_FAILED               = EC(420006, "unmap volume failed")
	ERR_OLD_TARGET_DAEMON_IS_ABNORMAL     = EC(420007, "old target daemon is abnormal")
	ERR_TARGET_DAEMON_IS_ABNORMAL         = EC(420008, "target daemon is abnormal")

	// 430: common (curvefs client)
	ERR_FS_PATH_ALREADY_MOUNTED  = EC(430000, "path already mounted")
	ERR_CREATE_FILESYSTEM_FAILED = EC(430001, "create filesystem failed")
	ERR_MOUNT_FILESYSTEM_FAILED  = EC(430002, "mount filesystem failed")
	ERR_UMOUNT_FILESYSTEM_FAILED = EC(430003, "umount filesystem failed")

	// 440: common (polarfs)
	ERR_GET_OS_REELASE_FAILED       = EC(440000, "get os release failed")
	ERR_UNSUPPORT_LINUX_OS_REELASE  = EC(440001, "unsupport linux os release")
	ERR_INSTALL_PFSD_PACKAGE_FAILED = EC(440002, "install pfsd package failed")

	// 450: common (playground)
	ERR_PLAYGROUND_NOT_FOUND = EC(450000, "playground not found")

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
	ERR_ETCD_REQUIRES_3_HOSTS             = EC(503005, "etcd requires at least 3 hosts for deploy")
	ERR_MDS_REQUIRES_3_HOSTS              = EC(503006, "mds requires at least 3 hosts for deploy")
	ERR_CHUNKSERVER_REQUIRES_3_HOSTS      = EC(503007, "chunkserver requires at least 3 hosts to distrubute zones")
	ERR_SNAPSHOTCLONE_REQUIRES_3_HOSTS    = EC(503008, "snapshotclone requires at least 3 hosts for deploy")
	ERR_METASERVER_REQUIRES_3_HOSTS       = EC(503009, "metaserver requires at least 3 hosts to distrubute zones")

	// 510: checker (ssh)
	ERR_SSH_CONNECT_FAILED = EC(510000, "SSH connect failed")

	// 520: checker (permission)
	ERR_USER_NOT_FOUND                                     = EC(520000, "user not found")
	ERR_HOSTNAME_NOT_RESOLVED                              = EC(520001, "hostname not resolved")
	ERR_CREATE_DIRECOTRY_PERMISSION_DENIED                 = EC(520002, "create direcotry permission denied")
	ERR_EXECUTE_CONTAINER_ENGINE_COMMAND_PERMISSION_DENIED = EC(520003, "execute docker/podman command permission denied")

	// 530: checker (kernel)
	ERR_UNRECOGNIZED_KERNEL_VERSION              = EC(530000, "unrecognized kernel version")
	ERR_RENAMEAT_NOT_SUPPORTED_IN_CURRENT_KERNEL = EC(530001, "renameat() not supported in current kernel version")
	ERR_KERNEL_NBD_MODULE_NOT_LOADED             = EC(530002, "kernel nbd module not loaded")
	ERR_KERNEL_FUSE_MODULE_NOT_LOADED            = EC(530003, "kernel fuse module not loaded")

	// 540: checker (network)
	ERR_PORT_ALREADY_IN_USE                = EC(540000, "port is already in use")
	ERR_DESTINATION_UNREACHABLE            = EC(540001, "destination unreachable")
	ERR_CONNET_MOCK_SERVICE_ADDRESS_FAILED = EC(540002, "try to connect mock service listen address failed")

	// 550: checker (date)
	ERR_INVALID_DATE_FORMAT                  = EC(550000, "invalid date format")
	ERR_HOST_TIME_DIFFERENCE_OVER_30_SECONDS = EC(550001, "host time difference over 30 seconds")

	// 560: checker (service)
	ERR_CHUNKFILE_POOL_NOT_EXIST = EC(560000, "there is no chunkfile pool in data directory")

	// 570: checker (client)
	ERR_INVALID_CURVEFS_CLIENT_S3_ACCESS_KEY  = EC(570000, "invalid curvefs client S3 access key")
	ERR_INVALID_CURVEFS_CLIENT_S3_SECRET_KEY  = EC(570001, "invalid curvefs client S3 secret key")
	ERR_INVALID_CURVEFS_CLIENT_S3_ADDRESS     = EC(570002, "invalid curvefs client S3 address")
	ERR_INVALID_CURVEFS_CLIENT_S3_BUCKET_NAME = EC(570003, "invalid curvefs client S3 bucket name")

	// 590: checker (others)
	ERR_CONTAINER_ENGINE_NOT_INSTALLED = EC(590000, "container engine docker/podman not installed")
	ERR_DOCKER_DAEMON_IS_NOT_RUNNING   = EC(590001, "docker daemon is not running")
	ERR_NO_SPACE_LEFT_ON_DEVICE        = EC(590002, "no space left on device")

	// 600: exeute task (common)
	ERR_EXECUTE_COMMAND_TIMED_OUT = EC(600000, "execute command timed out")
	ERR_READ_FILE_FAILED          = EC(600001, "read file failed")
	ERR_WRITE_FILE_FAILED         = EC(600002, "write file failed")
	ERR_BUILD_REGEX_FAILED        = EC(600003, "build regex failed")
	ERR_BUILD_TEMPLATE_FAILED     = EC(600004, "build template failed")

	// 610: exeute task (ssh command)
	ERR_DOWNLOAD_FILE_FROM_REMOTE_BY_SSH_FAILED         = EC(610000, "download file from remote by ssh failed")
	ERR_UPLOAD_FILE_TO_REMOTE_BY_SSH_FAILED             = EC(610001, "upload file to remote by ssh failed")
	ERR_CONNECT_REMOTE_HOST_WITH_INTERACT_BY_SSH_FAILED = EC(610002, "connect remote host with interact by ssh failed")

	// 620: execute task (shell command)
	ERR_EDIT_FILE_FAILED                           = EC(620000, "edit file failed (sed)")
	ERR_LIST_DIRECTORY_CONTENTS_FAILED             = EC(620001, "list directory contents failed (ls)")
	ERR_CREATE_DIRECTORY_FAILED                    = EC(620002, "create directory failed (mkdir)")
	ERR_REMOVE_FILES_OR_DIRECTORIES_FAILED         = EC(620003, "remove files or directories failed (rm)")
	ERR_RENAME_FILE_OR_DIRECTORY_FAILED            = EC(620004, "rename file or directory failed (mv)")
	ERR_COPY_FILES_AND_DIRECTORIES_FAILED          = EC(620005, "copy files and directories failed (cp)")
	ERR_CHANGE_FILE_MODE_FAILED                    = EC(620006, "change file mode failed (chmod)")
	ERR_GET_FILE_STATUS_FAILED                     = EC(620007, "get file status failed (stat)")
	ERR_CONCATENATE_FILE_FAILED                    = EC(620008, "concatenate file failed (cat)")
	ERR_BUILD_A_LINUX_FILE_SYSTEM_FAILED           = EC(620009, "build a Linux file system failed (mkfs)")
	ERR_MOUNT_A_FILESYSTEM_FAILED                  = EC(620010, "mount a filesystem failed (mount)")
	ERR_UNMOUNT_FILE_SYSTEMS_FAILED                = EC(620011, "unmount file systems failed (umount)")
	ERR_FIND_WHICH_PROCESS_USING_FILE_FAILED       = EC(620012, "find which process is using file failed (fuser)")
	ERR_GET_DISK_SPACE_USAGE_FAILED                = EC(620013, "get disk space usage failed (df)")
	ERR_LIST_BLOCK_DEVICES_FAILED                  = EC(620014, "list block devices failed (lsblk)")
	ERR_GET_CONNECTION_INFORMATION_FAILED          = EC(620015, "get connection information failed (ss)")
	ERR_SEND_ICMP_ECHO_REQUEST_TO_HOST_FAILED      = EC(620016, "send ICMP ECHO_REQUEST to host failed (ping)")
	ERR_TRANSFERRING_DATA_FROM_OR_TO_SERVER_FAILED = EC(620017, "transferring data from or to a server (curl)")
	ERR_GET_SYSTEM_TIME_FAILED                     = EC(620018, "get system time failed (date)")
	ERR_GET_SYSTEM_INFORMATION_FAILED              = EC(620019, "get system information failed (uname)")
	ERR_GET_KERNEL_MODULE_INFO_FAILED              = EC(620020, "get kernel module information failed (modinfo)")
	ERR_ADD_MODUDLE_FROM_LINUX_KERNEL_FAILED       = EC(620021, "add module from linux kernel failed (modprobe)")
	ERR_GET_HOSTNAME_FAILED                        = EC(620022, "get hostname failed (hostname)")
	ERR_STORES_AND_EXTRACTS_FILES_FAILED           = EC(620023, "stores and extracts files failed (tar)")
	ERR_INSTALL_OR_REMOVE_DEBIAN_PACKAGE_FAILED    = EC(620024, "install or remove debian package failed (dpkg)")
	ERR_INSTALL_OR_REMOVE_RPM_PACKAGE_FAILED       = EC(620025, "install or remove rpm package failed (rpm)")
	ERR_SECURE_COPY_FILE_TO_REMOTE_FAILED          = EC(620026, "secure copy file to remote failed (scp)")
	ERR_GET_BLOCK_DEVICE_UUID_FAILED               = EC(620027, "get block device uuid failed (blkid)")
	ERR_RESERVE_FILESYSTEM_BLOCKS_FAILED           = EC(620028, "reserve filesystem blocks (tune2fs)")
	ERR_RUN_SCRIPT_FAILED                          = EC(620998, "run script failed (bash script.sh)")
	ERR_RUN_A_BASH_COMMAND_FAILED                  = EC(620999, "run a bash command failed (bash -c)")

	// 630: execute task (docker/podman command)
	ERR_GET_CONTAINER_ENGINE_INFO_FAILED = EC(630000, "get container engine info failed")
	ERR_PULL_IMAGE_FAILED                = EC(630001, "pull image failed")
	ERR_CREATE_CONTAINER_FAILED          = EC(630002, "create container failed")
	ERR_START_CONTAINER_FAILED           = EC(630003, "start container failed")
	ERR_STOP_CONTAINER_FAILED            = EC(630004, "stop container failed")
	ERR_RESTART_CONTAINER_FAILED         = EC(630005, "restart container failed")
	ERR_WAIT_CONTAINER_STOP_FAILED       = EC(630006, "wait container stop failed")
	ERR_REMOVE_CONTAINER_FAILED          = EC(630007, "remove container failed")
	ERR_LIST_CONTAINERS_FAILED           = EC(630008, "list containers failed")
	ERR_RUN_COMMAND_IN_CONTAINER_FAILED  = EC(630009, "run a command in container failed")
	ERR_COPY_FROM_CONTAINER_FAILED       = EC(630010, "copy file from container failed")
	ERR_COPY_INTO_CONTAINER_FAILED       = EC(630011, "copy file into container failed")
	ERR_INSPECT_CONTAINER_FAILED         = EC(630012, "get container low-level information failed")
	ERR_GET_CONTAINER_LOGS_FAILED        = EC(630013, "get container logs failed")
	ERR_UPDATE_CONTAINER_FAILED          = EC(630014, "update container failed")
	ERR_TOP_CONTAINER_FAILED             = EC(630015, "top container failed")

	// 690: execuetr task (others)
	ERR_START_CRONTAB_IN_CONTAINER_FAILED = EC(690000, "start crontab in container failed")

	// 900: others
	ERR_CANCEL_OPERATION = EC(CODE_CANCEL_OPERATION, "cancel operation")
	// 999
	ERR_UNKNOWN = EC(999999, "unknown error")
)
