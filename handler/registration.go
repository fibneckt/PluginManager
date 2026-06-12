package handler

import (
	"fmt"

	"PluginsManager/model"
)

// RegisterPlugin 注册插件
func (m *PluginManager) RegisterPlugin(name string, version string, plugin model.Plugin) error {
	m.Mu.Lock()
	defer m.Mu.Unlock()

	if _, exists := m.Plugins[name]; exists {
		return fmt.Errorf("plugin %s already registered", name)
	}

	m.Plugins[name] = plugin
	m.Enabled[name] = true
	delete(m.Disabled, name)

	return nil
}

// EnablePlugin 启用插件。如果管理器已启动，自动创建通道和 worker。
func (m *PluginManager) EnablePlugin(name string) error {
	m.Mu.Lock()
	defer m.Mu.Unlock()

	if _, exists := m.Plugins[name]; !exists {
		return fmt.Errorf("plugin %s not found", name)
	}
	m.Enabled[name] = true
	delete(m.Disabled, name)
	delete(m.Errored, name)

	if m.started {
		ch := make(chan []byte, m.bufferSize)
		m.chs[name] = ch
		go m.worker(name, ch)
	}
	return nil
}

// DisablePlugin 禁用插件
func (m *PluginManager) DisablePlugin(name string) error {
	m.Mu.Lock()
	defer m.Mu.Unlock()

	if _, exists := m.Plugins[name]; !exists {
		return fmt.Errorf("plugin %s not found", name)
	}
	m.Disabled[name] = true
	delete(m.Enabled, name)
	return nil
}

// UnregisterPlugin 注销插件
func (m *PluginManager) UnregisterPlugin(name string) error {
	m.Mu.Lock()
	defer m.Mu.Unlock()

	if _, exists := m.Plugins[name]; !exists {
		return fmt.Errorf("plugin %s not found", name)
	}
	delete(m.Plugins, name)
	delete(m.Enabled, name)
	delete(m.Disabled, name)
	delete(m.Errored, name)
	return nil
}
