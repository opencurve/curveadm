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

package module

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	log "github.com/opencurve/curveadm/pkg/log/glg"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
)

const (
	HTTP_PROTOCOL = "http"
)

type (
	HttpConfig struct {
		Host string
		Port uint
	}

	HttpClient struct {
		config HttpConfig
		client *http.Client
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
	data := make(map[string]interface{})
	data["command"] = command
	bytesData, _ := json.Marshal(data)

	baseURL, _ := url.Parse(fmt.Sprintf("http://%s:%d", client.config.Host, client.config.Port))
	params := url.Values{}
	params.Add("method", "cluster.deploy.cmd")
	baseURL.RawQuery = params.Encode()
	resp, err := client.client.Post(baseURL.String(), "application/json", bytes.NewReader(bytesData))
	if err != nil {
		return
	}
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	result := &HttpResult{}
	err = json.Unmarshal(respData, result)
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
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)
	fh, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer fh.Close()
	fileWriter, err := bodyWriter.CreateFormFile("file", localPath)
	if err != nil {
		return err
	}
	_, err = io.Copy(fileWriter, fh)
	if err != nil {
		return err
	}

	baseURL, _ := url.Parse(fmt.Sprintf("http://%s:%d", client.config.Host, client.config.Port))
	params := url.Values{}
	params.Add("method", "cluster.deploy.upload")
	baseURL.RawQuery = params.Encode()
	boundary := "--boundary"
	bodyWriter.SetBoundary(boundary)
	bodyWriter.WriteField("filepath", remotePath)
	bodyWriter.Close()
	resp, err := client.client.Post(baseURL.String(), bodyWriter.FormDataContentType(), bodyBuf)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return err
}

func (client *HttpClient) Download(remotePath string, localPath string) (err error) {
	baseURL, _ := url.Parse(fmt.Sprintf("http://%s:%d", client.config.Host, client.config.Port))
	params := url.Values{}
	params.Add("method", "cluster.deploy.download")
	params.Add("filepath", remotePath)
	baseURL.RawQuery = params.Encode()
	resp, err := client.client.Get(baseURL.String())
	if err != nil {
		return
	}
	defer resp.Body.Close()
	localFile, err := os.Create(localPath)
	if err != nil {
		return
	}
	defer localFile.Close()
	_, err = io.Copy(localFile, resp.Body)
	return
}

func (client *HttpClient) Close() {

}

func (client *HttpClient) Config() HttpConfig {
	return client.config
}

func NewHttpClient(config HttpConfig) (*HttpClient, error) {
	return &HttpClient{
		config: config,
		client: &http.Client{},
	}, nil
}
