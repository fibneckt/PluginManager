package handler

import (
	"PluginsManager/logger"
	"PluginsManager/model"
)

// Start 启动所有已启用插件，每个插件一个独立 goroutine 监听数据（多次调用安全）。
func (m *PluginManager) Start(bufferSize int) {
	m.Mu.Lock()
	defer m.Mu.Unlock()

	if m.started {
		return
	}
	m.started = true
	m.bufferSize = bufferSize

	m.chs = make(map[string]chan []byte)
	m.resultCh = make(chan model.PluginResult, bufferSize*2)
	m.stopCh = make(chan struct{})

	for name := range m.Enabled {
		ch := make(chan []byte, bufferSize)
		m.chs[name] = ch
		go m.worker(name, ch)
	}

	logger.Infof("[PluginManager] started %d Plugins, buffer=%d", len(m.chs), bufferSize)
}

// Stop 停止所有插件 goroutine（多次调用安全）。
func (m *PluginManager) Stop() {
	m.stopOnce.Do(func() {
		close(m.stopCh)
		logger.Info("[PluginManager] stopped")
	})
}
