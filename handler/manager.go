package handler

import (
	"sync"

	"PluginsManager/model"
)

type PluginManager struct {
	Plugins  map[string]model.Plugin // 已注册的插件实例
	Enabled  map[string]bool         // 启用的插件名
	Disabled map[string]bool         // 禁用的插件名
	Errored  map[string]bool         // 运行异常的插件名

	// 通道执行模式
	chs        map[string]chan []byte     // 每个插件独立的数据通道
	resultCh   chan model.PluginResult    // 结果通道，外部从此读取执行结果
	stopCh     chan struct{}              // 停止信号
	bufferSize int                        // 通道缓冲区大小
	started    bool                       // 是否已启动
	stopOnce   sync.Once                  // 确保 Stop 只执行一次
	Mu         sync.RWMutex               // 并发保护
}

// NewPluginManager 创建插件管理器
func NewPluginManager() *PluginManager {
	return &PluginManager{
		Plugins:  make(map[string]model.Plugin),
		Enabled:  make(map[string]bool),
		Disabled: make(map[string]bool),
		Errored:  make(map[string]bool),
	}
}

// ImportFrom 从外部注册表导入插件（通常在启动时从 model.GetRegistered() 调用）。
func (m *PluginManager) ImportFrom(plugins map[string]model.Plugin, enabled map[string]bool) {
	m.Mu.Lock()
	defer m.Mu.Unlock()
	for name, p := range plugins {
		m.Plugins[name] = p
		if enabled[name] {
			m.Enabled[name] = true
		}
	}
}

// GetPlugin 返回已注册的插件实例（线程安全）。
// 外部代码应使用此方法访问插件，不要直接操作 Plugins map。
func (m *PluginManager) GetPlugin(name string) (model.Plugin, bool) {
	m.Mu.RLock()
	defer m.Mu.RUnlock()
	p, ok := m.Plugins[name]
	return p, ok
}
