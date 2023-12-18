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
	"fmt"
	"regexp"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/opencurve/curveadm/internal/storage/driver"
)

var (
	ErrInvalidDBUrl = fmt.Errorf("invalid database url")
)

// rqlite://127.0.0.1:4000
// sqlite:///home/curve/.curveadm/data/curveadm.db
const (
	REGEX_DB_URL = "^(sqlite|rqlite)://(.+)$"
)

type Monitor struct {
	ClusterId int
	Monitor   string
}

type Storage struct {
	db driver.IDataBaseDriver
}

func NewStorage(dbURL string) (*Storage, error) {
	pattern := regexp.MustCompile(REGEX_DB_URL)
	mu := pattern.FindStringSubmatch(dbURL)
	if len(mu) == 0 {
		return nil, ErrInvalidDBUrl
	}

	storage := &Storage{}
	if mu[1] == "sqlite" {
		storage.db = driver.NewSQLiteDB()
	} else {
		storage.db = driver.NewRQLiteDB()
	}

	err := storage.db.Open(dbURL)
	if err == nil {
		err = storage.init()
	}
	return storage, err
}

func (s *Storage) init() error {
	sqls := []string{
		CreateVersionTable,
		CreateHostsTable,
		CreateClustersTable,
		CreateContainersTable,
		CreateClientsTable,
		CreatePlaygroundTable,
		CreateAuditTable,
		CreateMonitorTable,
		CreateAnyTable,
		CreateClustersView,
		InsertTrigger,
		UpdateTrigger,
		DeleteTrigger,
	}

	tablesOK, err := s.CheckViewExists()
	if err != nil {
		return err
	}
	if !tablesOK {
		// allow failed to execute
		_, err = s.db.Write(RenameClusters)
		_, err = s.db.Write(AddTypeField)
	}

	for _, sql := range sqls {
		_, err := s.db.Write(sql)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Storage) write(query string, args ...any) error {
	_, err := s.db.Write(query, args...)
	return err
}

func (s *Storage) Close() error {
	return s.db.Close()
}

// version
func (s *Storage) SetVersion(version, lastConfirm string) error {
	versions, err := s.GetVersions()
	if err != nil {
		return err
	} else if len(versions) == 0 {
		return s.write(InsertVersion, version)
	}
	return s.write(SetVersion, version, lastConfirm, versions[0].Id)
}

func (s *Storage) GetVersions() ([]Version, error) {
	result, err := s.db.Query(SelectVersion)
	if err != nil {
		return nil, err
	}
	defer result.Close()

	var versions []Version
	var version Version
	for result.Next() {
		err = result.Scan(&version.Id, &version.Version, &version.LastConfirm)
		versions = append(versions, version)
		break
	}
	return versions, err
}

// hosts
func (s *Storage) SetHosts(data string) error {
	hostses, err := s.GetHostses()
	if err != nil {
		return err
	} else if len(hostses) == 0 {
		return s.write(InsertHosts, data)
	}
	return s.write(SetHosts, data, hostses[0].Id)
}

func (s *Storage) GetHostses() ([]Hosts, error) {
	result, err := s.db.Query(SelectHosts)
	if err != nil {
		return nil, err
	}
	defer result.Close()

	var hostses []Hosts
	var hosts Hosts
	for result.Next() {
		err = result.Scan(&hosts.Id, &hosts.Data, &hosts.LastModifiedTime)
		hostses = append(hostses, hosts)
		break
	}
	return hostses, err
}

// cluster
func (s *Storage) InsertCluster(name, uuid, description, topology, deployType string) error {
	return s.write(InsertCluster, uuid, name, description, topology, deployType)
}

func (s *Storage) CheckViewExists() (bool, error) {
	result, err := s.db.Query(isViewExist)
	if err != nil {
		return false, err
	}
	defer result.Close()

	var exists bool
	for result.Next() {
		err = result.Scan(&exists)
		if err != nil {
			return false, err
		}
		break
	}
	return exists, nil
}

func (s *Storage) DeleteCluster(name string) error {
	return s.write(DeleteCluster, name)
}

func (s *Storage) getClusters(query string, args ...interface{}) ([]Cluster, error) {
	result, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer result.Close()

	clusters := []Cluster{}
	for result.Next() {
		cluster := Cluster{}
		err = result.Scan(
			&cluster.Id,
			&cluster.UUId,
			&cluster.Name,
			&cluster.Description,
			&cluster.Topology,
			&cluster.Pool,
			&cluster.CreateTime,
			&cluster.Current,
			&cluster.Type,
		)
		if err != nil {
			return nil, err
		}
		clusters = append(clusters, cluster)
	}

	return clusters, nil
}

func (s *Storage) GetClusters(name string) ([]Cluster, error) {
	return s.getClusters(SelectCluster, name)
}

func (s *Storage) CheckoutCluster(name string) error {
	return s.write(CheckoutCluster, name)
}

func (s *Storage) GetCurrentCluster() (Cluster, error) {
	cluster := Cluster{Id: -1, Name: ""}
	clusters, err := s.getClusters(GetCurrentCluster)
	if err != nil {
		return cluster, err
	} else if len(clusters) == 1 {
		return clusters[0], nil
	}

	return cluster, nil
}

func (s *Storage) SetClusterTopology(id int, topology string) error {
	return s.write(SetClusterTopology, topology, id)
}

func (s *Storage) SetClusterPool(id int, topology, pool string) error {
	return s.write(SetClusterPool, topology, pool, id)
}

// service
func (s *Storage) InsertService(clusterId int, serviceId, containerId string) error {
	return s.write(InsertService, serviceId, clusterId, containerId)
}

func (s *Storage) getServices(query string, args ...interface{}) ([]Service, error) {
	result, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer result.Close()

	services := []Service{}
	var service Service
	for result.Next() {
		err = result.Scan(&service.Id, &service.ClusterId, &service.ContainerId)
		if err != nil {
			return nil, err
		}
		services = append(services, service)
	}

	return services, nil
}

func (s *Storage) GetServices(clusterId int) ([]Service, error) {
	return s.getServices(SelectServicesInCluster, clusterId)
}

func (s *Storage) GetContainerId(serviceId string) (string, error) {
	services, err := s.getServices(SelectService, serviceId)
	if err != nil || len(services) == 0 {
		return "", err
	}

	return services[0].ContainerId, nil
}

func (s *Storage) SetContainId(serviceId, containerId string) error {
	return s.write(SetContainerId, containerId, serviceId)
}

// client
func (s *Storage) InsertClient(id, kind, host, containerId, auxInfo string) error {
	return s.write(InsertClient, id, kind, host, containerId, auxInfo)
}

func (s *Storage) SetClientAuxInfo(id, auxInfo string) error {
	return s.write(SetClientAuxInfo, id, auxInfo)
}

func (s *Storage) getClients(query string, args ...interface{}) ([]Client, error) {
	result, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer result.Close()

	clients := []Client{}
	var client Client
	for result.Next() {
		err = result.Scan(&client.Id, &client.Kind, &client.Host, &client.ContainerId, &client.AuxInfo)
		if err != nil {
			return nil, err
		}
		clients = append(clients, client)
	}

	return clients, nil
}

func (s *Storage) GetClientContainerId(id string) (string, error) {
	clients, err := s.getClients(SelectClientById, id)
	if err != nil || len(clients) == 0 {
		return "", err
	}

	return clients[0].ContainerId, nil
}

func (s *Storage) GetClient(id string) ([]Client, error) {
	return s.getClients(SelectClientById, id)
}

func (s *Storage) GetClients() ([]Client, error) {
	return s.getClients(SelectClients)
}

func (s *Storage) DeleteClient(id string) error {
	return s.write(DeleteClient, id)
}

// playground
func (s *Storage) InsertPlayground(name, mountPoint string) error {
	// FIXME: remove status
	return s.write(InsertPlayground, name, mountPoint, "")
}

func (s *Storage) SetPlaygroundStatus(name, status string) error {
	return s.write(SetPlaygroundStatus, status, name)
}

func (s *Storage) DeletePlayground(name string) error {
	return s.write(DeletePlayground, name)
}

func (s *Storage) getPlaygrounds(query string, args ...interface{}) ([]Playground, error) {
	result, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer result.Close()

	playgrounds := []Playground{}
	var playground Playground
	for result.Next() {
		err = result.Scan(
			&playground.Id,
			&playground.Name,
			&playground.CreateTime,
			&playground.MountPoint,
			&playground.Status)
		if err != nil {
			return nil, err
		}
		playgrounds = append(playgrounds, playground)
	}

	return playgrounds, nil
}

func (s *Storage) GetPlaygrounds(name string) ([]Playground, error) {
	return s.getPlaygrounds(SelectPlayground, name)
}

func (s *Storage) GetPlaygroundById(id string) ([]Playground, error) {
	return s.getPlaygrounds(SelectPlaygroundById, id)
}

// audit
func (s *Storage) InsertAuditLog(time time.Time, workDir, command string, status int) (int64, error) {
	result, err := s.db.Write(InsertAuditLog, time, workDir, command, status)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func (s *Storage) SetAuditLogStatus(id int64, status, errorCode int) error {
	return s.write(SetAuditLogStatus, status, errorCode, id)
}

func (s *Storage) getAuditLogs(query string, args ...interface{}) ([]AuditLog, error) {
	result, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer result.Close()

	auditLogs := []AuditLog{}
	var auditLog AuditLog
	for result.Next() {
		err = result.Scan(&auditLog.Id,
			&auditLog.ExecuteTime,
			&auditLog.WorkDirectory,
			&auditLog.Command,
			&auditLog.Status,
			&auditLog.ErrorCode)
		if err != nil {
			return nil, err
		}
		auditLogs = append(auditLogs, auditLog)
	}

	return auditLogs, nil
}

func (s *Storage) GetAuditLogs() ([]AuditLog, error) {
	return s.getAuditLogs(SelectAuditLog)
}

func (s *Storage) GetAuditLog(id int64) ([]AuditLog, error) {
	return s.getAuditLogs(SelectAuditLogById, id)
}

// any item prefix
const (
	PREFIX_CLIENT_CONFIG = 0x01
)

func (s *Storage) realId(prefix int, id string) string {
	return fmt.Sprintf("%d:%s", prefix, id)
}

func (s *Storage) InsertClientConfig(id, data string) error {
	id = s.realId(PREFIX_CLIENT_CONFIG, id)
	return s.write(InsertAnyItem, id, data)
}

func (s *Storage) GetClientConfig(id string) ([]Any, error) {
	id = s.realId(PREFIX_CLIENT_CONFIG, id)
	result, err := s.db.Query(SelectAnyItem, id)
	if err != nil {
		return nil, err
	}
	defer result.Close()

	items := []Any{}
	var item Any
	for result.Next() {
		err = result.Scan(&item.Id, &item.Data)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, nil
}

func (s *Storage) DeleteClientConfig(id string) error {
	id = s.realId(PREFIX_CLIENT_CONFIG, id)
	return s.write(DeleteAnyItem, id)
}

func (s *Storage) GetMonitor(clusterId int) (Monitor, error) {
	monitor := Monitor{
		ClusterId: clusterId,
	}
	rows, err := s.db.Query(SelectMonitor, clusterId)
	if err != nil {
		return monitor, err
	}

	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&monitor.Monitor)
		if err != nil {
			return monitor, err
		}
	}
	return monitor, nil
}

func (s *Storage) InsertMonitor(m Monitor) error {
	return s.write(InsertMonitor, m.ClusterId, m.Monitor)
}

func (s *Storage) UpdateMonitor(m Monitor) error {
	return s.write(UpdateMonitor, m.Monitor, m.ClusterId)
}

func (s *Storage) DeleteMonitor(clusterId int) error {
	return s.write(DeleteMonitor, clusterId)
}

func (s *Storage) ReplaceMonitor(m Monitor) error {
	return s.write(ReplaceMonitor, m.ClusterId, m.Monitor)
}
