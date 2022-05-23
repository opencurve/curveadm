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
 *  limitations under the Licensele().
 */

/*
 * Project: CurveAdm
 * Created Date: 2021-10-15
 * Author: Jingli Chen (Wine93)
 */

package storage

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

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

type Service struct {
	Id          string
	ClusterId   int
	ContainerId string
}

type AuditLog struct {
	Id          int
	Command     string
	ExecuteTime time.Time
	Success     bool
}

type Storage struct {
	db    *sql.DB
	mutex *sync.Mutex
}

func NewStorage(dbfile string) (*Storage, error) {
	db, err := sql.Open("sqlite3", dbfile)
	if err != nil {
		return nil, err
	}

	s := &Storage{db: db, mutex: &sync.Mutex{}}
	if err = s.init(); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Storage) init() error {
	if err := s.execSQL(CREATE_CLUSTERS_TABLE); err != nil {
		return err
	} else if err := s.execSQL(CREATE_CONTAINERS_TABLE); err != nil {
		return err
	} else if err := s.execSQL(CREATE_AUDIT_TABLE); err != nil {
		return err
	} else if err := s.compatible(); err != nil {
		return err
	}

	return nil
}

func (s *Storage) compatible() error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	err = func() error {
		rows, err := tx.Query(CHECK_POOl_COLUMN)
		if err != nil {
			return err
		}

		if rows.Next() {
			count := 0
			err = rows.Scan(&count)
			rows.Close()
			if err != nil {
				return err
			}
			if count != 0 {
				return nil
			}
		}

		alterSQL := fmt.Sprintf("%s;%s;%s;%s",
			RENAME_CLUSTERS_TABLE,
			CREATE_CLUSTERS_TABLE,
			INSERT_CLUSTERS_FROM_OLD_TABLE,
			DROP_OLD_CLUSTERS_TABLE,
		)
		_, err = tx.Exec(alterSQL)
		return err
	}()

	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (s *Storage) execSQL(query string, args ...interface{}) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	stmt, err := s.db.Prepare(query)
	if err != nil {
		return err
	}

	_, err = stmt.Exec(args...)
	return err
}

func (s *Storage) Close() error {
	return s.db.Close()
}

// cluster
func (s *Storage) InsertCluster(name, description, topology string) error {
	return s.execSQL(INSERT_CLUSTER, name, description, topology)
}

func (s *Storage) DeleteCluster(name string) error {
	return s.execSQL(DELETE_CLUSTER, name)
}

func (s *Storage) getClusters(query string, args ...interface{}) ([]Cluster, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	clusters := []Cluster{}
	for rows.Next() {
		cluster := Cluster{}
		err = rows.Scan(&cluster.Id, &cluster.UUId, &cluster.Name, &cluster.Description,
			&cluster.Topology, &cluster.Pool, &cluster.CreateTime, &cluster.Current)
		if err != nil {
			return nil, err
		}
		clusters = append(clusters, cluster)
	}

	return clusters, nil
}

func (s *Storage) GetClusters(name string) ([]Cluster, error) {
	return s.getClusters(SELECT_CLUSTER, name)
}

func (s *Storage) CheckoutCluster(name string) error {
	return s.execSQL(CHECKOUT_CLUSTER, name)
}

func (s *Storage) GetCurrentCluster() (Cluster, error) {
	cluster := Cluster{Id: -1, Name: ""}
	clusters, err := s.getClusters(GET_CURRENT_CLUSTER)
	if err != nil {
		return cluster, err
	} else if len(clusters) == 1 {
		return clusters[0], nil
	}

	return cluster, nil
}

func (s *Storage) SetClusterTopology(id int, topology string) error {
	return s.execSQL(SET_CLUSTER_TOPOLOGY, topology, id)
}

func (s *Storage) SetClusterPool(id int, pool string) error {
	return s.execSQL(SET_CLUSTER_POOL, pool, id)
}

// service
func (s *Storage) InsertService(clusterId int, serviceId, containerId string) error {
	return s.execSQL(INSERT_SERVICE, serviceId, clusterId, containerId)
}

func (s *Storage) getServices(query string, args ...interface{}) ([]Service, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	services := []Service{}
	var service Service
	for rows.Next() {
		err = rows.Scan(&service.Id, &service.ClusterId, &service.ContainerId)
		if err != nil {
			return nil, err
		}
		services = append(services, service)
	}

	return services, nil
}

func (s *Storage) GetServices(clusterId int) ([]Service, error) {
	return s.getServices(SELECT_SERVICE_IN_CLUSTER, clusterId)
}

func (s *Storage) GetContainerId(serviceId string) (string, error) {
	services, err := s.getServices(SELECT_SERVICE, serviceId)
	if err != nil || len(services) == 0 {
		return "", err
	}

	return services[0].ContainerId, nil
}

func (s *Storage) SetContainId(serviceId, containerId string) error {
	return s.execSQL(SET_CONTAINER_ID, containerId, serviceId)
}

// audit
func (s *Storage) InsertAuditLog(time time.Time, command string, success bool) error {
	return s.execSQL(INSERT_AUDIT_LOG, time, command, success)
}

func (s *Storage) getAuditLogs(query string, args ...interface{}) ([]AuditLog, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	auditLogs := []AuditLog{}
	var auditLog AuditLog
	for rows.Next() {
		err = rows.Scan(&auditLog.Id, &auditLog.ExecuteTime, &auditLog.Command, &auditLog.Success)
		if err != nil {
			return nil, err
		}
		auditLogs = append(auditLogs, auditLog)
	}

	return auditLogs, nil
}

func (s *Storage) GetAuditLogs() ([]AuditLog, error) {
	return s.getAuditLogs(SELECT_AUDIT_LOG)
}
