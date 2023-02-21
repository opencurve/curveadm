/*
 *  Copyright (c) 2021 NetEase Inc.
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
 * Created Date: 2021-10-15
 * Author: Jingli Chen (Wine93)
 */

/*
 * +------------------------------+
 * |go        | sqlite3           |
 * |----------|-------------------|
 * |nil       | null              |
 * |int       | integer           |
 * |int64     | integer           |
 * |float64   | float             |
 * |bool      | integer           |
 * |[]byte    | blob              |
 * |string    | text              |
 * |time.Time | timestamp/datetime|
 * +------------------------------+
 */

package storage

var (
	// tables (hosts/clusters/containers(service)/clients/playrgound/audit/disk/disks)
	CREATE_VERSION_TABLE = `
		CREATE TABLE IF NOT EXISTS version (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			version TEXT NOT NULL,
			lastconfirm TEXT NOT NULL
		)
	`

	CREATE_HOSTS_TABLE = `
		CREATE TABLE IF NOT EXISTS hosts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			data TEXT NOT NULL,
			lastmodified_time DATE NOT NULL
		)
	`
	CREATE_DISKS_TABLE = `
		CREATE TABLE IF NOT EXISTS disks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			data TEXT NOT NULL,
			lastmodified_time DATE NOT NULL
		)
	`

	CREATE_DISK_TABLE = `
		CREATE TABLE IF NOT EXISTS disk (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			host TEXT NOT NULL,
			device TEXT NOT NULL,
			size TEXT NOT NULL,
			uri TEXT NOT NULL,
			disk_format_mount_point TEXT NOT NULL,
			format_percent TEXT NOT NULL,
			container_image_location TEXT NOT NULL,
			chunkserver_id TEXT NOT NULL,
			lastmodified_time DATE NOT NULL
		)
	`

	CREATE_CLUSTERS_TABLE = `
		CREATE TABLE IF NOT EXISTS clusters (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid TEXT NOT NULL,
			name TEXT NOT NULL UNIQUE,
			description TEXT,
			topology TEXT NULL,
			pool TEXT NULL,
			create_time DATE NOT NULL,
			current INTEGER DEFAULT 0
		)
	`

	// id: clusterId_role_host_(sequence/name)
	CREATE_CONTAINERS_TABLE = `
		CREATE TABLE IF NOT EXISTS containers (
			id TEXT PRIMARY KEY,
			cluster_id INTEGER NOT NULL,
			container_id TEXT NOT NULL
		)
	`

	CREATE_CLIENTS_TABLE = `
		CREATE TABLE IF NOT EXISTS clients (
			id TEXT PRIMARY KEY,
			kind TEXT NOT NULL,
			host TEXT NOT NULL,
			container_id TEXT NOT NULL,
			aux_info TEXT NOT NULL
		)
	`

	CREATE_PLAYGROUND_TABLE = `
       CREATE TABLE IF NOT EXISTS playgrounds (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			create_time DATE NOT NULL,
			mount_point TEXT NOT NULL,
            status TEXT NOT NULL
		)
    `

	CREATE_AUDIT_TABLE = `
        CREATE TABLE IF NOT EXISTS audit (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
            execute_time DATE NOT NULL,
            work_directory TEXT NOT NULL,
			command TEXT NOT NULL,
			status INTEGER DEFAULT 0,
            error_code INTEGET DEFAULT 0
		)
    `

	CHECK_POOl_COLUMN = `
		SELECT COUNT(*) AS total
		FROM pragma_table_info('clusters')
		WHERE name='pool'
	`

	RENAME_CLUSTERS_TABLE = `ALTER TABLE clusters RENAME TO clusters_old`

	INSERT_CLUSTERS_FROM_OLD_TABLE = `
		INSERT INTO clusters(id, uuid, name, description, topology, pool, create_time, current)
		SELECT id, uuid, name, description, topology, "", create_time, current
		FROM clusters_old
	`

	DROP_OLD_CLUSTERS_TABLE = `DROP TABLE clusters_old`

	// version
	INSERT_VERSION = `INSERT INTO version(version, lastconfirm) VALUES(?, "")`

	SET_VERSION = `UPDATE version SET version = ?, lastconfirm = ? WHERE id = ?`

	SELECT_VERSION = `SELECT * FROM version`

	// hosts
	INSERT_HOSTS = `INSERT INTO hosts(data, lastmodified_time) VALUES(?, datetime('now','localtime'))`

	SET_HOSTS = `UPDATE hosts SET data = ?, lastmodified_time = datetime('now','localtime') WHERE id = ?`

	SELECT_HOSTS = `SELECT * FROM hosts`

	// disks
	INSERT_DISKS = `INSERT INTO disks(data, lastmodified_time) VALUES(?, datetime('now','localtime'))`

	SET_DISKS = `UPDATE disks SET data = ?, lastmodified_time = datetime('now','localtime') WHERE id = ?`

	SELECT_DISKS = `SELECT * FROM disks`

	// disk
	INSERT_DISK = `INSERT INTO disk(
		host,
		device,
		size,
		uri,
		disk_format_mount_point,
		format_percent,
		container_image_location,
		chunkserver_id,
		lastmodified_time
		) VALUES(?, ?, ?, ?, ?, ?, ?, ?, datetime('now','localtime'))`

	SET_DISK = `UPDATE disk SET disk_format_mount_point = ?, format_percent = ?,
	container_image_location = ?,lastmodified_time = datetime('now','localtime') WHERE id = ?`

	SET_DISK_URI = `UPDATE disk SET uri = ?,
		lastmodified_time = datetime('now','localtime') WHERE host = ? AND device = ?`

	SET_DISK_SIZE = `UPDATE disk SET size = ?,
		lastmodified_time = datetime('now','localtime') WHERE host = ? AND device = ?`

	SET_DISK_CHUNKSERVER_ID = `UPDATE disk SET chunkserver_id = ?,
	lastmodified_time = datetime('now','localtime') WHERE host = ? AND disk_format_mount_point = ?`

	SELECT_DISK_ALL = `SELECT * FROM disk`

	SELECT_DISK_BY_HOST = `SELECT * FROM disk where host = ?`

	SELECT_DISK_BY_CHUNKSERVER_ID = `SELECT * FROM disk where chunkserver_id = ?`

	SELECT_DISK_BY_DEVICE_PATH = `SELECT * from disk WHERE host = ? AND device = ?`

	SELECT_DISK_BY_DISK_FORMAT_MOUNTPOINT = `SELECT * from disk WHERE host = ? AND disk_format_mount_point = ?`

	DELETE_DISK_HOST = `DELETE from disk WHERE host = ?`

	DELETE_DISK_HOST_DEVICE = `DELETE from disk WHERE host = ? AND device = ?`

	// cluster
	INSERT_CLUSTER = `INSERT INTO clusters(uuid, name, description, topology, pool, create_time)
                                  VALUES(hex(randomblob(16)), ?, ?, ?, "", datetime('now','localtime'))`

	DELETE_CLUSTER = `DELETE from clusters WHERE name = ?`

	SELECT_CLUSTER = `SELECT * FROM clusters WHERE name LIKE ?`

	GET_CURRENT_CLUSTER = `SELECT * FROM clusters WHERE current = 1`

	CHECKOUT_CLUSTER = `
		UPDATE clusters
		SET current = CASE name
    		WHEN ? THEN 1
			ELSE 0
		END
	`

	SET_CLUSTER_TOPOLOGY = `UPDATE clusters SET topology = ? WHERE id = ?`

	SET_CLUSTER_POOL = `UPDATE clusters SET topology = ?, pool = ? WHERE id = ?`

	// service
	INSERT_SERVICE = `INSERT INTO containers(id, cluster_id, container_id) VALUES(?, ?, ?)`

	SELECT_SERVICE = `SELECT * FROM containers WHERE id = ?`

	SELECT_SERVICE_IN_CLUSTER = `SELECT * FROM containers WHERE cluster_id = ?`

	SET_CONTAINER_ID = `UPDATE containers SET container_id = ? WHERE id = ?`

	// client
	INSERT_CLIENT = `INSERT INTO clients(id, kind, host, container_id, aux_info) VALUES(?, ?, ?, ?, ?)`

	SELECT_CLIENTS = `SELECT * FROM clients`

	SELECT_CLIENT_BY_ID = `SELECT * FROM clients WHERE id = ?`

	DELETE_CLIENT = `DELETE from clients WHERE id = ?`

	// playground
	INSERT_PLAYGROUND = `INSERT INTO playgrounds(name, create_time, mount_point, status)
                                     VALUES(?, datetime('now','localtime'), ?, ?)`

	SET_PLAYGROUND_STATUS = `UPDATE playgrounds SET status = ? WHERE name = ?`

	SELECT_PLAYGROUND = `SELECT * FROM playgrounds WHERE name LIKE ?`

	SELECT_PLAYGROUND_BY_ID = `SELECT * FROM playgrounds WHERE id = ?`

	DELETE_PLAYGROUND = `DELETE from playgrounds WHERE name = ?`

	// audit
	INSERT_AUDIT_LOG = `INSERT INTO audit(execute_time, work_directory, command, status)
                                    VALUES(?, ?, ?, ?)`

	SET_AUDIT_LOG_STATUS = `UPDATE audit SET status = ?, error_code = ? WHERE id = ?`

	SELECT_AUDIT_LOG = `SELECT * FROM audit`

	SELECT_AUDIT_LOG_BY_ID = `SELECT * FROM audit WHERE id = ?`
)
