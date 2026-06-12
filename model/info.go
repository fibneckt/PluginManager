package model

type PluginsStatus string

const (
	PluginsStatusEnabled  PluginsStatus = "Enabled"  // 启用
	PluginsStatusDisabled PluginsStatus = "Disabled" // 禁用
	PluginsStatusError    PluginsStatus = "error"    // 异常
)

type Plugin interface {
	// 插件接口方法
	Name() string
	Version() string
	Run(data map[string]interface{}) (map[string]interface{}, error)
}

type PluginInfo struct {
	// 插件基本元信息
	Name    string        `json:"name"`
	Version string        `json:"version"`
	Status  PluginsStatus `json:"status"`
}

type PluginResult struct {
	// 插件运行结果
	Name   string
	Output map[string]interface{}
	Err    error
}

// Configurable 可配置插件接口，需要读取 YAML 配置的插件实现此接口
type Configurable interface {
	Plugin
	Configure(config map[string]interface{}) error
}

type PluginConfig struct {
	// PluginConfig YAML 中的单个插件配置
	Name    string                 `yaml:"name"`
	Enabled bool                   `yaml:"Enabled"`
	Config  map[string]interface{} `yaml:"config"`
}

type PluginsConfig struct {
	// PluginsConfig 顶层 YAML 配置
	Plugins []PluginConfig `yaml:"Plugins"`
}
