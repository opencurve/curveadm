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
 * Created Date: 2022-05-18
 * Author: Jingli Chen (Wine93)
 */

package plugin

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/opencurve/curveadm/internal/utils"
)

const (
	PLUGIN_METADATA_FILE   = "META"
	PLUGIN_ENTRYPOINT_FILE = "main.yaml"
	KEY_META_NAME          = "Name"
	KEY_META_VERSION       = "Version"
	KEY_META_RELEASED      = "Released"
	KEY_META_DESCRIPTION   = "Description"
	ENV_CURVEADM_PLUGIN    = "CURVEADM_PLUGIN"
	URL_INSTALL_SCRIPT     = "http://curveadm.nos-eastchina1.126.net/script/install.sh"
	REGEX_META_PAIR        = "^([^:]+):(.*)$"
)

type Plugin struct {
	Name           string
	Version        string
	ReleasedTime   string
	Description    string
	EntrypointPath string
}

type PluginManager struct {
	rootDir string
}

func NewPluginManager(rootDir string) *PluginManager {
	return &PluginManager{
		rootDir: rootDir,
	}
}

func (pm *PluginManager) Install(name string) error {
	cmd := exec.Command("/bin/bash", "-c", fmt.Sprintf("curl -fsSL %s | bash", URL_INSTALL_SCRIPT))
	cmd.Env = append(os.Environ(), fmt.Sprintf("%s=%s", ENV_CURVEADM_PLUGIN, name))
	return cmd.Run()
}

func (pm *PluginManager) Remove(name string) error {
	pluginDir := filepath.Join(pm.rootDir, name)
	metaPath := filepath.Join(pluginDir, PLUGIN_METADATA_FILE)
	if !utils.PathExist(metaPath) {
		return fmt.Errorf("Plugin '%s' not installed", name)
	}
	return os.RemoveAll(pluginDir)
}

func (pm *PluginManager) Load(name string) (*Plugin, error) {
	metaPath := filepath.Join(pm.rootDir, name, PLUGIN_METADATA_FILE)
	if !utils.PathExist(metaPath) {
		return nil, nil
	}

	data, err := utils.ReadFile(metaPath)
	if err != nil {
		return nil, err
	}

	plugin := &Plugin{
		EntrypointPath: filepath.Join(pm.rootDir, name, PLUGIN_ENTRYPOINT_FILE),
	}
	scanner := bufio.NewScanner(strings.NewReader(data))
	for scanner.Scan() {
		line := scanner.Text()
		regex, err := regexp.Compile(REGEX_META_PAIR)
		if err != nil {
			return nil, err
		}

		mu := regex.FindStringSubmatch(line)
		if len(mu) == 0 {
			continue
		}

		key, value := mu[1], mu[2]
		switch key {
		case KEY_META_NAME:
			plugin.Name = strings.TrimSpace(value)
		case KEY_META_VERSION:
			plugin.Version = strings.TrimSpace(value)
		case KEY_META_RELEASED:
			plugin.ReleasedTime = strings.TrimSpace(value)
		case KEY_META_DESCRIPTION:
			plugin.Description = strings.TrimSpace(value)
		}
	}
	return plugin, err
}

func (pm *PluginManager) List() ([]*Plugin, error) {
	files, err := ioutil.ReadDir(pm.rootDir)
	if err != nil {
		return nil, err
	}

	var plugins []*Plugin
	for _, file := range files {
		if !file.IsDir() {
			continue
		}

		plugin, _ := pm.Load(file.Name())
		if plugin != nil && len(plugin.Name) > 0 {
			plugins = append(plugins, plugin)
		} else {
			plugins = append(plugins, &Plugin{
				Name:         file.Name(),
				Version:      "-",
				ReleasedTime: "-",
				Description:  "<MetaData Broken>",
			})
		}
	}
	return plugins, nil
}
