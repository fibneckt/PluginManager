package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"PluginsManager/cmd"
	"PluginsManager/config"
	"PluginsManager/handler"
	"PluginsManager/logger"
	"PluginsManager/model"
	_ "PluginsManager/plugins"
)

func main() {
	logger.Setup(&logger.Settings{
		Path:       "logs",
		Name:       "plugin-manager",
		Ext:        "log",
		TimeFormat: "2006-01-02",
	})

	// 创建管理器并导入 init() 阶段注册的插件
	pm := handler.NewPluginManager()
	pm.ImportFrom(model.GetRegistered())

	// 关键错误直接输出到 stderr，不走异步日志
	if err := config.LoadFromConfig("plugins/config.yaml", pm); err != nil {
		fmt.Fprintln(os.Stderr, "load config failed:", err)
		os.Exit(1)
	}

	pm.Start(16)

	logger.Info("PluginManager started")

	// 结果收集：异步读取所有插件的执行结果并记录日志
	go func() {
		for result := range pm.Results() {
			if result.Err != nil {
				logger.Errorf("[%s] error: %v", result.Name, result.Err)
			} else {
				output, _ := json.Marshal(result.Output)
				logger.Infof("[%s] %s", result.Name, string(output))
			}
		}
	}()

	// 列出已加载的插件
	enabled, _ := pm.GetEnabledPlugins()
	for _, p := range enabled {
		logger.Infof("Loaded: %s v%s", p.Name, p.Version)
	}

	// 进入交互式命令行
	shell := cmd.NewShell(pm)
	shell.Run()

	// Stop 由 sync.Once 保护，重复调用安全
	pm.Stop()

	// 给异步日志一点时间刷盘
	time.Sleep(100 * time.Millisecond)
}
