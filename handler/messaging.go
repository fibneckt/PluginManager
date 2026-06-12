package handler

import (
	"fmt"

	"PluginsManager/logger"
	"PluginsManager/model"
)

// Send 发送 JSON 数据到所有已启用插件。数据满时丢弃并返回错误。
func (m *PluginManager) Send(jsonData []byte) error {
	m.Mu.RLock()
	defer m.Mu.RUnlock()

	var drops []string
	for name, ch := range m.chs {
		select {
		case ch <- jsonData:
		default:
			drops = append(drops, name)
		}
	}
	if len(drops) > 0 {
		logger.Errorf("[PluginManager] channel full, dropping data for: %v", drops)
		return fmt.Errorf("dropped data for plugins: %v", drops)
	}
	return nil
}

// SendTo 发送 JSON 数据到指定插件
func (m *PluginManager) SendTo(name string, jsonData []byte) error {
	m.Mu.RLock()
	defer m.Mu.RUnlock()

	ch, ok := m.chs[name]
	if !ok {
		return fmt.Errorf("plugin %s not found or not started", name)
	}
	select {
	case ch <- jsonData:
	default:
		return fmt.Errorf("plugin %s channel full", name)
	}
	return nil
}

// Results 返回结果通道，外部从此读取各插件的执行结果
func (m *PluginManager) Results() <-chan model.PluginResult {
	return m.resultCh
}
