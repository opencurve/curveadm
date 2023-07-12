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

import "time"

// version
type Version struct {
	Id          int
	Version     string
	LastConfirm string
}

var (
	// table: version
	CreateVersionTable = `
		CREATE TABLE IF NOT EXISTS version (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			version TEXT NOT NULL,
			lastconfirm TEXT NOT NULL
		)
	`

	// insert version
	InsertVersion = `INSERT INTO version(version, lastconfirm) VALUES(?, "")`

	// set version
	SetVersion = `UPDATE version SET version = ?, lastconfirm = ? WHERE id = ?`

	// select version
	SelectVersion = `SELECT * FROM version`
)

// hosts
type Hosts struct {
	Id               int
	Data             string
	LastModifiedTime time.Time
}

var (
	// table: hosts
	CreateHostsTable = `
		CREATE TABLE IF NOT EXISTS hosts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			data TEXT NOT NULL,
			lastmodified_time DATE NOT NULL
		)
	`

	// insert hosts
	InsertHosts = `INSERT INTO hosts(data, lastmodified_time) VALUES(?, datetime('now','localtime'))`

	// set hosts
	SetHosts = `UPDATE hosts SET data = ?, lastmodified_time = datetime('now','localtime') WHERE id = ?`

	// select hosts
	SelectHosts = `SELECT * FROM hosts`
)

// cluster
type Cluster struct {
	Id          int
	UUId        string
	Name        string
	Description string
	CreateTime  time.Time
	Topology    string
	Pool        string
	Current     bool
}

var (
	// table: clusters
	CreateClustersTable = `
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

	// insert cluster
	InsertCluster = `
		INSERT INTO clusters(uuid, name, description, topology, pool, create_time)
		VALUES(?, ?, ?, ?, "", datetime('now','localtime'))
	`

	// delete cluster
	DeleteCluster = `DELETE from clusters WHERE name = ?`

	// select cluster
	SelectCluster = `SELECT * FROM clusters WHERE name LIKE ?`

	// get current cluster
	GetCurrentCluster = `SELECT * FROM clusters WHERE current = 1`

	// checkout cluster
	CheckoutCluster = `
		UPDATE clusters
		SET current = CASE name
			WHEN ? THEN 1
			ELSE 0
		END
	`

	// set cluster topology
	SetClusterTopology = `UPDATE clusters SET topology = ? WHERE id = ?`

	// set cluster pool
	SetClusterPool = `UPDATE clusters SET topology = ?, pool = ? WHERE id = ?`
)

// service
type Service struct {
	Id          string
	ClusterId   int
	ContainerId string
}

var (
	// table: containers
	// id: clusterId_role_host_(sequence/name)
	CreateContainersTable = `
		CREATE TABLE IF NOT EXISTS containers (
			id TEXT PRIMARY KEY,
			cluster_id INTEGER NOT NULL,
			container_id TEXT NOT NULL
		)
	`

	// insert service
	InsertService = `INSERT INTO containers(id, cluster_id, container_id) VALUES(?, ?, ?)`

	// select service
	SelectService = `SELECT * FROM containers WHERE id = ?`

	// select services in cluster
	SelectServicesInCluster = `SELECT * FROM containers WHERE cluster_id = ?`

	// set service container id
	SetContainerId = `UPDATE containers SET container_id = ? WHERE id = ?`
)

// client
type Client struct {
	Id          string
	Kind        string
	Host        string
	ContainerId string
	AuxInfo     string
}

var (
	// table: clients
	CreateClientsTable = `
		CREATE TABLE IF NOT EXISTS clients (
			id TEXT PRIMARY KEY,
			kind TEXT NOT NULL,
			host TEXT NOT NULL,
			container_id TEXT NOT NULL,
			aux_info TEXT NOT NULL
		)
	`
	// insert client
	InsertClient = `INSERT INTO clients(id, kind, host, container_id, aux_info) VALUES(?, ?, ?, ?, ?)`

	// set client aux info
	SetClientAuxInfo = `UPDATE clients SET aux_info = ? WHERE id = ?`

	// select clients
	SelectClients = `SELECT * FROM clients`

	// select client by id
	SelectClientById = `SELECT * FROM clients WHERE id = ?`

	// delete client
	DeleteClient = `DELETE from clients WHERE id = ?`
)

// playground
type Playground struct {
	Id         int
	Name       string
	CreateTime time.Time
	MountPoint string
	Status     string
}

var (
	// table: playground
	CreatePlaygroundTable = `
		CREATE TABLE IF NOT EXISTS playgrounds (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			create_time DATE NOT NULL,
			mount_point TEXT NOT NULL,
			status TEXT NOT NULL
		)
	`

	// insert playground
	InsertPlayground = `
		INSERT INTO playgrounds(name, create_time, mount_point, status)
                    VALUES(?, datetime('now','localtime'), ?, ?)
	`

	// set playground status
	SetPlaygroundStatus = `UPDATE playgrounds SET status = ? WHERE name = ?`

	// select playground
	SelectPlayground = `SELECT * FROM playgrounds WHERE name LIKE ?`

	// select playground by id
	SelectPlaygroundById = `SELECT * FROM playgrounds WHERE id = ?`

	// delete playground
	DeletePlayground = `DELETE from playgrounds WHERE name = ?`
)

// audit log
type AuditLog struct {
	Id            int
	ExecuteTime   time.Time
	WorkDirectory string
	Command       string
	Status        int
	ErrorCode     int
}

var (
	// table: audit
	CreateAuditTable = `
		CREATE TABLE IF NOT EXISTS audit (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			execute_time DATE NOT NULL,
			work_directory TEXT NOT NULL,
			command TEXT NOT NULL,
			status INTEGER DEFAULT 0,
			error_code INTEGET DEFAULT 0
		)
	`

	// insert audit log
	InsertAuditLog = `
		INSERT INTO audit(execute_time, work_directory, command, status)
		            VALUES(?, ?, ?, ?)
	`

	// set audit log status
	SetAuditLogStatus = `UPDATE audit SET status = ?, error_code = ? WHERE id = ?`

	// select audit log
	SelectAuditLog = `SELECT * FROM audit`

	// select audit log by id
	SelectAuditLogById = `SELECT * FROM audit WHERE id = ?`
)

var (
	// check pool column
	CheckPoolColumn = `
		SELECT COUNT(*) AS total
		FROM pragma_table_info('clusters')
		WHERE name='pool'
	`

	// rename clusters table
	RenameClustersTable = `ALTER TABLE clusters RENAME TO clusters_old`

	// insert clusters from old table
	InsertClustersFromOldTable = `
		INSERT INTO clusters(id, uuid, name, description, topology, pool, create_time, current)
		SELECT id, uuid, name, description, topology, "", create_time, current
		FROM clusters_old
	`

	// statement: drom old clusters table
	DropOldClustersTable = `DROP TABLE clusters_old`
)
