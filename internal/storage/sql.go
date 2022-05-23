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
	// tables (clusters/containers/audit)
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

	CHECK_POOl_COLUMN = `
		SELECT COUNT(*) AS total
		FROM pragma_table_info('clusters')
		WHERE name='pool'
	`

	RENAME_CLUSTERS_TABLE = `ALTER TABLE clusters RENAME TO clusters_old`

	INSERT_CLUSTERS_FROM_OLD_TABLE = `
		INSERT INTO clusters(id, uuid, name, description, topology, pool, create_time, current)
		SELECT id, hex(randomblob(16)) uuid, name, description, topology, "", create_time, current
		FROM clusters_old
	`

	DROP_OLD_CLUSTERS_TABLE = `DROP TABLE clusters_old`

	// id: clusterId_role_host_(sequence/name)
	CREATE_CONTAINERS_TABLE = `
		CREATE TABLE IF NOT EXISTS containers (
			id TEXT PRIMARY KEY,
			cluster_id INTEGER NOT NULL,
			container_id TEXT NOT NULL
		)
	`

	CREATE_AUDIT_TABLE = `
        CREATE TABLE IF NOT EXISTS audit (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
            execute_time DATE NOT NULL,
			command TEXT NOT NULL,
			success INTEGER DEFAULT 0
		)
    `

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

	SET_CLUSTER_POOL = `UPDATE clusters SET pool = ? WHERE id = ?`

	// container
	INSERT_SERVICE = `INSERT INTO containers(id, cluster_id, container_id) VALUES(?, ?, ?)`

	SELECT_SERVICE = `SELECT * FROM containers WHERE id = ?`

	SELECT_SERVICE_IN_CLUSTER = `SELECT * FROM containers WHERE cluster_id = ?`

	SET_CONTAINER_ID = `UPDATE containers SET container_id = ? WHERE id = ?`

	// audit
	INSERT_AUDIT_LOG = `INSERT INTO audit(execute_time, command, success) VALUES(?, ?, ?)`

	SELECT_AUDIT_LOG = `SELECT * FROM audit`
)
