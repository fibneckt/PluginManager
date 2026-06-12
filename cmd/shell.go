package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"PluginsManager/handler"
	"PluginsManager/logger"
	"PluginsManager/model"
)

// Shell 提供交互式命令行界面，用于管理插件和发送数据。
type Shell struct {
	pm      *handler.PluginManager
	scanner *bufio.Scanner
}

// NewShell 创建一个新的 Shell 实例。
func NewShell(pm *handler.PluginManager) *Shell {
	return &Shell{
		pm:      pm,
		scanner: bufio.NewScanner(os.Stdin),
	}
}

// Run 启动命令行主循环。返回 true 表示通过 /quit 退出，false 表示 stdin 结束。
func (s *Shell) Run() (quit bool) {
	fmt.Println("PluginManager ready. Type /help for commands.")

	for s.scanner.Scan() {
		line := strings.TrimSpace(s.scanner.Text())
		if line == "" {
			continue
		}
		if !s.handleCommand(line) {
			return true
		}
	}
	return false
}

// handleCommand 解析并执行单条命令。返回 false 表示应该退出。
func (s *Shell) handleCommand(line string) bool {
	parts := strings.SplitN(line, " ", 2)
	cmd := parts[0]
	arg := ""
	if len(parts) > 1 {
		arg = parts[1]
	}

	switch cmd {
	case "/plugins":
		enabled, _ := s.pm.GetEnabledPlugins()
		disabled, _ := s.pm.GetDisabledPlugins()
		errored, _ := s.pm.GetErroredPlugins()
		all := append(append(enabled, disabled...), errored...)
		printPluginList("All Registered", all)

	case "/enabled":
		list, _ := s.pm.GetEnabledPlugins()
		printPluginList("Enabled", list)

	case "/disabled":
		list, _ := s.pm.GetDisabledPlugins()
		printPluginList("Disabled", list)

	case "/errored":
		list, _ := s.pm.GetErroredPlugins()
		printPluginList("Errored", list)

	case "/info":
		if arg == "" {
			fmt.Println("Usage: /info <plugin_name>")
			return true
		}
		info, err := s.pm.GetInfoPlugin(arg)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return true
		}
		fmt.Printf("  %-16s %s\n", "Name:", info.Name)
		fmt.Printf("  %-16s %s\n", "Version:", info.Version)
		fmt.Printf("  %-16s %s\n", "Status:", info.Status)

	case "/enable":
		if arg == "" {
			fmt.Println("Usage: /enable <plugin_name>")
			return true
		}
		if err := s.pm.EnablePlugin(arg); err != nil {
			fmt.Printf("Error: %v\n", err)
		} else {
			fmt.Printf("Plugin %s enabled\n", arg)
		}

	case "/disable":
		if arg == "" {
			fmt.Println("Usage: /disable <plugin_name>")
			return true
		}
		if err := s.pm.DisablePlugin(arg); err != nil {
			fmt.Printf("Error: %v\n", err)
		} else {
			fmt.Printf("Plugin %s disabled\n", arg)
		}

	case "/unregister":
		if arg == "" {
			fmt.Println("Usage: /unregister <plugin_name>")
			return true
		}
		if err := s.pm.UnregisterPlugin(arg); err != nil {
			fmt.Printf("Error: %v\n", err)
		} else {
			fmt.Printf("Plugin %s unregistered\n", arg)
		}

	case "/send":
		pairs := parseSendArgs(arg)
		if len(pairs) == 0 {
			fmt.Println("Usage: /send <plugin_name> <json> [<plugin_name> <json> ...]")
			return true
		}
		for _, pair := range pairs {
			if err := s.pm.SendTo(pair.name, []byte(pair.json)); err != nil {
				fmt.Printf("Error [%s]: %v\n", pair.name, err)
			} else {
				logger.Infof("Send to %s: %s", pair.name, pair.json)
			}
		}

	case "/sendall":
		if arg == "" {
			fmt.Println("Usage: /sendall <json>")
			return true
		}
		if err := s.pm.Send([]byte(arg)); err != nil {
			fmt.Printf("Error: %v\n", err)
		} else {
			logger.Infof("SendAll: %s", arg)
		}

	case "/help":
		printHelp()

	case "/quit":
		return false

	default:
		fmt.Printf("Unknown command: %s (type /help)\n", cmd)
	}
	return true
}

func printPluginList(title string, list []*model.PluginInfo) {
	fmt.Printf("\n--- %s Plugins (%d) ---\n", title, len(list))
	if len(list) == 0 {
		fmt.Println("  (none)")
		return
	}
	for _, p := range list {
		fmt.Printf("  %-20s %-10s %s\n", p.Name, p.Version, p.Status)
	}
	fmt.Println()
}

func printHelp() {
	fmt.Println("Commands:")
	fmt.Println("  /plugins              List all registered plugins")
	fmt.Println("  /enabled              List enabled plugins")
	fmt.Println("  /disabled             List disabled plugins")
	fmt.Println("  /errored              List errored plugins")
	fmt.Println("  /info <name>          Show plugin detail")
	fmt.Println("  /enable <name>        Enable a plugin")
	fmt.Println("  /disable <name>       Disable a plugin")
	fmt.Println("  /unregister <name>    Unregister a plugin")
	fmt.Println("  /send <name> <json>   Send JSON to specific plugin")
	fmt.Println("  /sendall <json>       Send JSON to all plugins")
	fmt.Println("  /quit                 Exit")
}

type sendPair struct {
	name string
	json string
}

// parseSendArgs 解析 "plugin1 {json1} plugin2 {json2} ..." 格式的参数
func parseSendArgs(arg string) []sendPair {
	var pairs []sendPair
	s := strings.TrimSpace(arg)

	for len(s) > 0 {
		// 1. 提取插件名（第一个空白之前的词）
		space := strings.IndexByte(s, ' ')
		if space < 0 {
			break
		}
		name := s[:space]
		s = strings.TrimSpace(s[space+1:])

		// 2. 提取 JSON 对象（通过大括号计数）
		if len(s) == 0 || s[0] != '{' {
			break
		}
		depth := 0
		jsonEnd := -1
		for i, ch := range s {
			if ch == '{' {
				depth++
			} else if ch == '}' {
				depth--
				if depth == 0 {
					jsonEnd = i + 1
					break
				}
			}
		}
		if jsonEnd < 0 {
			break
		}
		jsonStr := s[:jsonEnd]
		s = strings.TrimSpace(s[jsonEnd:])

		pairs = append(pairs, sendPair{name: name, json: jsonStr})
	}
	return pairs
}
