package config

import (
	"PluginsManager/handler"
	"PluginsManager/model"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// LoadFromConfig 从 YAML 配置文件对已注册插件应用配置。
// pm 是由调用方显式传入的 PluginManager，避免隐式依赖全局变量。
func LoadFromConfig(path string, pm *handler.PluginManager) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read config: %v", err)
	}

	var cfg model.PluginsConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("failed to parse config: %v", err)
	}

	for _, pc := range cfg.Plugins {
		// 应用启用/禁用状态
		if pc.Enabled {
			if err := pm.EnablePlugin(pc.Name); err != nil {
				return fmt.Errorf("plugin %s: %v", pc.Name, err)
			}
		} else {
			if err := pm.DisablePlugin(pc.Name); err != nil {
				return fmt.Errorf("plugin %s: %v", pc.Name, err)
			}
		}

		// 如果插件实现了 Configurable，传入专属配置
		if plugin, exists := pm.GetPlugin(pc.Name); exists && pc.Config != nil {
			if configurable, ok := plugin.(model.Configurable); ok {
				if err := configurable.Configure(pc.Config); err != nil {
					return fmt.Errorf("plugin %s configure failed: %v", pc.Name, err)
				}
			}
		}
	}

	return nil
}
