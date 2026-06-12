package model

import (
	"fmt"
	"sync"
)

// registry 轻量级注册表，仅在 init() 阶段使用。
// 避免 handler 包暴露全局 DefaultManager，方便测试和多实例场景。
type registry struct {
	mu      sync.RWMutex
	plugins map[string]Plugin
	enabled map[string]bool
}

var defaultRegistry = &registry{
	plugins: make(map[string]Plugin),
	enabled: make(map[string]bool),
}

// RegisterPlugin 注册插件到默认注册表（供 init() 调用）。
func RegisterPlugin(name string, version string, plugin Plugin) error {
	defaultRegistry.mu.Lock()
	defer defaultRegistry.mu.Unlock()

	if _, exists := defaultRegistry.plugins[name]; exists {
		return fmt.Errorf("plugin %s already registered", name)
	}

	defaultRegistry.plugins[name] = plugin
	defaultRegistry.enabled[name] = true
	return nil
}

// GetRegistered 返回所有已注册的插件及其启用状态。
// PluginManager 在启动时调用此方法导入 init() 阶段注册的插件。
func GetRegistered() (plugins map[string]Plugin, enabled map[string]bool) {
	defaultRegistry.mu.RLock()
	defer defaultRegistry.mu.RUnlock()

	plugins = make(map[string]Plugin, len(defaultRegistry.plugins))
	enabled = make(map[string]bool, len(defaultRegistry.enabled))
	for k, v := range defaultRegistry.plugins {
		plugins[k] = v
	}
	for k, v := range defaultRegistry.enabled {
		enabled[k] = v
	}
	return
}
