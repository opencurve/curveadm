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
	CREATE_CLUSTERS_TABLE = `
		CREATE TABLE IF NOT EXISTS clusters (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			description TEXT,
			topology TEXT NULL,
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

	// cluster
	INSERT_CLUSTER = `INSERT INTO clusters(name, description, topology, create_time)
                                  VALUES(?, ?, ?, datetime('now','localtime'))`

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

	// container
	INSERT_SERVICE = `INSERT INTO containers(id, cluster_id, container_id) VALUES(?, ?, ?)`

	SELECT_SERVICE = `SELECT * FROM containers WHERE id = ?`

	SELECT_SERVICE_IN_CLUSTER = `SELECT * FROM containers WHERE cluster_id = ?`

	SET_CONTAINER_ID = `UPDATE containers SET container_id = ? WHERE id = ?`
)
