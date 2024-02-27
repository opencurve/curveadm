/*
*  Copyright (c) 2023 NetEase Inc.
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
* Created Date: 2023-12-13
* Author: liuminjian
 */

package module

import (
	"context"
	"fmt"
	"github.com/go-resty/resty/v2"
	log "github.com/opencurve/curveadm/pkg/log/glg"
)

const (
	HTTP_PROTOCOL = "http"
)

type (
	HTTPConfig struct {
		Host string
		Port uint
	}

	HttpClient struct {
		config HTTPConfig
		client *resty.Client
	}

	HttpResult struct {
		Data      string `json:"data"`
		ErrorCode string `json:"errorCode"`
		ErrorMsg  string `json:"errorMsg"`
	}
)

func (client *HttpClient) Protocol() string {
	return HTTP_PROTOCOL
}

func (client *HttpClient) WrapperCommand(command string, execInLocal bool) (wrapperCmd string) {
	return command
}

func (client *HttpClient) RunCommand(ctx context.Context, command string) (out []byte, err error) {
	result := &HttpResult{}
	_, err = client.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(map[string]interface{}{"command": command}).
		SetQueryParam("method", "cluster.deploy.cmd").
		SetResult(result).
		Post(fmt.Sprintf("http://%s:%d", client.config.Host, client.config.Port))
	if err != nil {
		return
	}
	log.Info("http resp", log.Field("result", result))
	if result.ErrorCode != "0" {
		return []byte(result.Data), fmt.Errorf(result.ErrorMsg)
	}
	return []byte(result.Data), nil
}

func (client *HttpClient) RemoteAddr() (addr string) {
	config := client.Config()
	return fmt.Sprintf("%s:%d", config.Host, config.Port)
}

func (client *HttpClient) Upload(localPath string, remotePath string) (err error) {
	_, err = client.client.R().
		SetFile("file", localPath).
		SetFormData(map[string]string{"filepath": remotePath}).
		SetQueryParam("method", "cluster.deploy.upload").
		Post(fmt.Sprintf("http://%s:%d", client.config.Host, client.config.Port))
	return err
}

func (client *HttpClient) Download(remotePath string, localPath string) (err error) {
	_, err = client.client.R().
		SetQueryParam("method", "cluster.deploy.download").
		SetQueryParam("filepath", remotePath).
		SetOutput(localPath).
		Get(fmt.Sprintf("http://%s:%d", client.config.Host, client.config.Port))
	return
}

func (client *HttpClient) Close() {

}

func (client *HttpClient) Config() HTTPConfig {
	return client.config
}

func NewHTTPClient(config HTTPConfig) (*HttpClient, error) {
	return &HttpClient{
		config: config,
		client: resty.New(),
	}, nil
}
