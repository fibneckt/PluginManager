package handler

import (
	"fmt"

	"PluginsManager/model"
)

// GetInfoPlugin 获取插件信息
func (m *PluginManager) GetInfoPlugin(name string) (*model.PluginInfo, error) {
	m.Mu.RLock()
	defer m.Mu.RUnlock()

	plugin, exists := m.Plugins[name]
	if !exists {
		return nil, fmt.Errorf("plugin %s not found", name)
	}

	info := &model.PluginInfo{
		Name:    plugin.Name(),
		Version: plugin.Version(),
	}
	if m.Errored[name] {
		info.Status = model.PluginsStatusError
	} else if m.Enabled[name] {
		info.Status = model.PluginsStatusEnabled
	} else if m.Disabled[name] {
		info.Status = model.PluginsStatusDisabled
	}

	return info, nil
}

// GetEnabledPlugins 获取已启用插件列表
func (m *PluginManager) GetEnabledPlugins() ([]*model.PluginInfo, error) {
	m.Mu.RLock()
	defer m.Mu.RUnlock()

	var infos []*model.PluginInfo
	for name := range m.Enabled {
		if m.Errored[name] {
			continue
		}
		if plugin, ok := m.Plugins[name]; ok {
			infos = append(infos, &model.PluginInfo{
				Name:    plugin.Name(),
				Version: plugin.Version(),
				Status:  model.PluginsStatusEnabled,
			})
		}
	}
	return infos, nil
}

// GetDisabledPlugins 获取已禁用插件列表
func (m *PluginManager) GetDisabledPlugins() ([]*model.PluginInfo, error) {
	m.Mu.RLock()
	defer m.Mu.RUnlock()

	var infos []*model.PluginInfo
	for name := range m.Disabled {
		if plugin, ok := m.Plugins[name]; ok {
			infos = append(infos, &model.PluginInfo{
				Name:    plugin.Name(),
				Version: plugin.Version(),
				Status:  model.PluginsStatusDisabled,
			})
		}
	}
	return infos, nil
}

// GetErroredPlugins 获取运行异常的插件列表（仍在启用列表中）
func (m *PluginManager) GetErroredPlugins() ([]*model.PluginInfo, error) {
	m.Mu.RLock()
	defer m.Mu.RUnlock()

	var infos []*model.PluginInfo
	for name := range m.Errored {
		if plugin, ok := m.Plugins[name]; ok {
			infos = append(infos, &model.PluginInfo{
				Name:    plugin.Name(),
				Version: plugin.Version(),
				Status:  model.PluginsStatusError,
			})
		}
	}
	return infos, nil
}
