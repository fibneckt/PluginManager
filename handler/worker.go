package handler

import (
	"encoding/json"
	"fmt"

	"PluginsManager/logger"
	"PluginsManager/model"
)

// worker 插件工作协程，从通道读取 JSON → 解析 → 调用插件 Run → 发送结果
func (m *PluginManager) worker(name string, ch chan []byte) {
	m.Mu.RLock()
	plugin := m.Plugins[name]
	m.Mu.RUnlock()

	for {
		select {
		case data := <-ch:
			m.safeRun(plugin, name, data)
		case <-m.stopCh:
			return
		}
	}
}

// safeRun 安全执行插件，捕获 panic 防止 worker 崩溃
func (m *PluginManager) safeRun(plugin model.Plugin, name string, data []byte) {
	defer func() {
		if r := recover(); r != nil {
			m.resultCh <- model.PluginResult{Name: name, Err: fmt.Errorf("panic: %v", r)}
			m.Mu.Lock()
			m.Errored[name] = true
			m.Mu.Unlock()
			logger.Errorf("[PluginManager] plugin %s panic: %v", name, r)
		}
	}()

	var input map[string]interface{}
	if err := json.Unmarshal(data, &input); err != nil {
		m.resultCh <- model.PluginResult{Name: name, Err: fmt.Errorf("json unmarshal: %v", err)}
		m.Mu.Lock()
		m.Errored[name] = true
		m.Mu.Unlock()
		logger.Errorf("[PluginManager] plugin %s json unmarshal error: %s", name, err)
		return
	}

	output, err := plugin.Run(input)
	m.resultCh <- model.PluginResult{Name: name, Output: output, Err: err}

	if err != nil {
		m.Mu.Lock()
		m.Errored[name] = true
		m.Mu.Unlock()
		logger.Errorf("[PluginManager] plugin %s run error: %v", name, err)
	} else {
		m.Mu.Lock()
		delete(m.Errored, name)
		m.Mu.Unlock()
	}
}
