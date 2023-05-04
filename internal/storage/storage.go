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
	comm "github.com/opencurve/curveadm/internal/common"
	"github.com/opencurve/curveadm/internal/errno"
)

type Version struct {
	Id          int
	Version     string
	LastConfirm string
}

type Hosts struct {
	Id               int
	Data             string
	LastmodifiedTime time.Time
}

type Disks struct {
	Id               int
	Data             string
	LastmodifiedTime time.Time
}

type Disk struct {
	Id               int
	Host             string
	Device           string
	Size             string
	URI              string
	MountPoint       string
	FormatPercent    int
	ContainerImage   string
	ChunkServerID    string
	LastmodifiedTime time.Time
}

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

type Client struct {
	Id          string
	Kind        string
	Host        string
	ContainerId string
	AuxInfo     string
}

type Playground struct {
	Id         int
	Name       string
	CreateTime time.Time
	MountPoint string
	Status     string
}

type AuditLog struct {
	Id            int
	ExecuteTime   time.Time
	WorkDirectory string
	Command       string
	Status        int
	ErrorCode     int
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
	if err := s.execSQL(CREATE_VERSION_TABLE); err != nil {
		return err
	} else if err := s.execSQL(CREATE_HOSTS_TABLE); err != nil {
		return err
	} else if err := s.execSQL(CREATE_HOSTS_TABLE); err != nil {
		return err
	} else if err := s.execSQL(CREATE_CLUSTERS_TABLE); err != nil {
		return err
	} else if err := s.execSQL(CREATE_CONTAINERS_TABLE); err != nil {
		return err
	} else if err := s.execSQL(CREATE_CLIENTS_TABLE); err != nil {
		return err
	} else if err := s.execSQL(CREATE_PLAYGROUND_TABLE); err != nil {
		return err
	} else if err := s.execSQL(CREATE_AUDIT_TABLE); err != nil {
		return err
	} else if err := s.execSQL(CREATE_DISKS_TABLE); err != nil {
		return err
	} else if err := s.execSQL(CREATE_DISK_TABLE); err != nil {
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

// version
func (s *Storage) SetVersion(version, lastConfirm string) error {
	versions, err := s.GetVersions()
	if err != nil {
		return err
	} else if len(versions) == 0 {
		return s.execSQL(INSERT_VERSION, version)
	}
	return s.execSQL(SET_VERSION, version, lastConfirm, versions[0].Id)
}

func (s *Storage) GetVersions() ([]Version, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	rows, err := s.db.Query(SELECT_VERSION)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	var versions []Version
	var version Version
	for rows.Next() {
		err = rows.Scan(&version.Id, &version.Version, &version.LastConfirm)
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
		return s.execSQL(INSERT_HOSTS, data)
	}
	return s.execSQL(SET_HOSTS, data, hostses[0].Id)
}

func (s *Storage) GetHostses() ([]Hosts, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	rows, err := s.db.Query(SELECT_HOSTS)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	var hostses []Hosts
	var hosts Hosts
	for rows.Next() {
		err = rows.Scan(&hosts.Id, &hosts.Data, &hosts.LastmodifiedTime)
		hostses = append(hostses, hosts)
		break
	}
	return hostses, err
}

// disks
func (s *Storage) SetDisks(data string) error {
	diskses, err := s.GetDisks()
	if err != nil {
		return err
	} else if len(diskses) == 0 {
		return s.execSQL(INSERT_DISKS, data)
	}
	return s.execSQL(SET_DISKS, data, diskses[0].Id)
}

func (s *Storage) GetDisks() ([]Disks, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	rows, err := s.db.Query(SELECT_DISKS)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var diskses []Disks
	var disks Disks
	for rows.Next() {
		err = rows.Scan(&disks.Id, &disks.Data, &disks.LastmodifiedTime)
		diskses = append(diskses, disks)
		break
	}
	return diskses, err
}

// disk
func (s *Storage) SetDisk(host, device, mount, containerImage string, formatPercent int) error {
	diskRecords, err := s.GetDisk(SELECT_DISK_BY_DEVICE_PATH, host, device)
	if err != nil {
		return err
	} else if len(diskRecords) == 0 {
		return s.execSQL(
			INSERT_DISK,
			host,
			device,
			comm.DISK_DEFAULT_NULL_SIZE,
			comm.DISK_DEFAULT_NULL_URI,
			mount,
			formatPercent,
			containerImage,
			comm.DISK_DEFAULT_NULL_CHUNKSERVER_ID,
		)
	}
	return s.execSQL(SET_DISK, mount, formatPercent, containerImage, diskRecords[0].Id)
}

func (s *Storage) UpdateDiskURI(host, device, devUri string) error {
	return s.execSQL(SET_DISK_URI, devUri, host, device)
}

func (s *Storage) UpdateDiskSize(host, device, size string) error {
	return s.execSQL(SET_DISK_SIZE, size, host, device)
}

func (s *Storage) UpdateDiskChunkServerID(host, mountPoint, chunkserverId string) error {
	_, err := s.GetDiskByMountPoint(host, mountPoint)
	if err != nil {
		return err
	}
	return s.execSQL(SET_DISK_CHUNKSERVER_ID, chunkserverId, host, mountPoint)
}

func (s *Storage) DeleteDisk(host, device string) error {
	if len(device) > 0 {
		return s.execSQL(DELETE_DISK_HOST_DEVICE, host, device)
	} else {
		return s.execSQL(DELETE_DISK_HOST, host)
	}
}

func (s *Storage) GetDiskByMountPoint(host, mountPoint string) (Disk, error) {
	var disk Disk
	diskRecords, err := s.GetDisk(comm.DISK_FILTER_MOUNT, host, mountPoint)
	if len(diskRecords) == 0 {
		return disk, errno.ERR_DATABASE_EMPTY_QUERY_RESULT.
			F("The disk[host=%s, disk_format_mount_point=%s] was not found in database.",
				host, mountPoint)
	}
	disk = diskRecords[0]
	return disk, err
}

func (s *Storage) CleanDiskChunkServerId(serviceId string) error {
	diskRecords, err := s.GetDisk(comm.DISK_FILTER_SERVICE, serviceId)
	if err != nil {
		return err
	}

	for _, disk := range diskRecords {
		if err := s.UpdateDiskChunkServerID(
			disk.Host,
			disk.MountPoint,
			comm.DISK_DEFAULT_NULL_CHUNKSERVER_ID,
		); err != nil {
			return err
		}
	}
	return nil
}

func (s *Storage) GetDisk(filter string, args ...interface{}) ([]Disk, error) {
	var query string
	switch filter {
	case comm.DISK_FILTER_ALL:
		query = SELECT_DISK_ALL
	case comm.DISK_FILTER_HOST:
		query = SELECT_DISK_BY_HOST
	case comm.DISK_FILTER_DEVICE:
		query = SELECT_DISK_BY_DEVICE_PATH
	case comm.DISK_FILTER_MOUNT:
		query = SELECT_DISK_BY_MOUNTPOINT
	case comm.DISK_FILTER_SERVICE:
		query = SELECT_DISK_BY_CHUNKSERVER_ID
	default:
		query = filter
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	var diskRecords []Disk
	var disk Disk

	for rows.Next() {
		err = rows.Scan(&disk.Id,
			&disk.Host,
			&disk.Device,
			&disk.Size,
			&disk.URI,
			&disk.MountPoint,
			&disk.FormatPercent,
			&disk.ContainerImage,
			&disk.ChunkServerID,
			&disk.LastmodifiedTime)
		diskRecords = append(diskRecords, disk)
	}
	return diskRecords, err
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

func (s *Storage) SetClusterPool(id int, topology, pool string) error {
	return s.execSQL(SET_CLUSTER_POOL, topology, pool, id)
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

// client
func (s *Storage) InsertClient(id, kind, host, containerId, auxInfo string) error {
	return s.execSQL(INSERT_CLIENT, id, kind, host, containerId, auxInfo)
}

func (s *Storage) getClients(query string, args ...interface{}) ([]Client, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	clients := []Client{}
	var client Client
	for rows.Next() {
		err = rows.Scan(&client.Id, &client.Kind, &client.Host, &client.ContainerId, &client.AuxInfo)
		if err != nil {
			return nil, err
		}
		clients = append(clients, client)
	}

	return clients, nil
}

func (s *Storage) GetClientContainerId(id string) (string, error) {
	clients, err := s.getClients(SELECT_CLIENT_BY_ID, id)
	if err != nil || len(clients) == 0 {
		return "", err
	}

	return clients[0].ContainerId, nil
}

func (s *Storage) GetClient(id string) ([]Client, error) {
	return s.getClients(SELECT_CLIENT_BY_ID, id)
}

func (s *Storage) GetClients() ([]Client, error) {
	return s.getClients(SELECT_CLIENTS)
}

func (s *Storage) DeleteClient(id string) error {
	return s.execSQL(DELETE_CLIENT, id)
}

// playground
func (s *Storage) InsertPlayground(name, mountPoint string) error {
	// FIXME: remove status
	return s.execSQL(INSERT_PLAYGROUND, name, mountPoint, "")
}

func (s *Storage) SetPlaygroundStatus(name, status string) error {
	return s.execSQL(SET_PLAYGROUND_STATUS, status, name)
}

func (s *Storage) DeletePlayground(name string) error {
	return s.execSQL(DELETE_PLAYGROUND, name)
}

func (s *Storage) getPlaygrounds(query string, args ...interface{}) ([]Playground, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	playgrounds := []Playground{}
	var playground Playground
	for rows.Next() {
		err = rows.Scan(
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
	return s.getPlaygrounds(SELECT_PLAYGROUND, name)
}

func (s *Storage) GetPlaygroundById(id string) ([]Playground, error) {
	return s.getPlaygrounds(SELECT_PLAYGROUND_BY_ID, id)
}

// audit
func (s *Storage) InsertAuditLog(time time.Time, workDir, command string, status int) (int64, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	stmt, err := s.db.Prepare(INSERT_AUDIT_LOG)
	if err != nil {
		return -1, err
	}

	result, err := stmt.Exec(time, workDir, command, status)
	if err != nil {
		return -1, err
	}

	return result.LastInsertId()
}

func (s *Storage) SetAuditLogStatus(id int64, status, errorCode int) error {
	return s.execSQL(SET_AUDIT_LOG_STATUS, status, errorCode, id)
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
		err = rows.Scan(&auditLog.Id,
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
	return s.getAuditLogs(SELECT_AUDIT_LOG)
}

func (s *Storage) GetAuditLog(id int64) ([]AuditLog, error) {
	return s.getAuditLogs(SELECT_AUDIT_LOG_BY_ID, id)
}
