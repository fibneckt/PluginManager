package plugins

import (
	"fmt"
	"math"
	"strconv"

	"PluginsManager/model"
)

func init() {
	p := &CalculatorPlugin{MaxPrecision: 6}
	if err := model.RegisterPlugin(p.Name(), p.Version(), p); err != nil {
		panic("register calculator: " + err.Error())
	}
}

// CalculatorPlugin 计算器插件
type CalculatorPlugin struct {
	MaxPrecision int // 可从 YAML 配置覆盖
}

func (p *CalculatorPlugin) Name() string    { return "calculator" }
func (p *CalculatorPlugin) Version() string { return "1.0.0" }

// Configure 从 YAML 读取插件专属配置
func (p *CalculatorPlugin) Configure(config map[string]interface{}) error {
	if v, ok := config["max_precision"]; ok {
		switch val := v.(type) {
		case int:
			p.MaxPrecision = val
		case float64:
			p.MaxPrecision = int(val)
		}
	}
	return nil
}

func (p *CalculatorPlugin) Run(data map[string]interface{}) (map[string]interface{}, error) {
	op, ok := data["op"].(string)
	if !ok {
		return nil, fmt.Errorf("missing op field")
	}

	a, err := p.toFloat(data["a"])
	if err != nil {
		return nil, fmt.Errorf("invalid a: %v", err)
	}
	b, err := p.toFloat(data["b"])
	if err != nil {
		return nil, fmt.Errorf("invalid b: %v", err)
	}

	var result float64
	switch op {
	case "add":
		result = a + b
	case "subtract":
		result = a - b
	case "multiply":
		result = a * b
	case "divide":
		if b == 0 {
			return nil, fmt.Errorf("division by zero")
		}
		result = a / b
	default:
		return nil, fmt.Errorf("unknown op: %s", op)
	}

	factor := math.Pow(10, float64(p.MaxPrecision))
	result = math.Round(result*factor) / factor

	return map[string]interface{}{
		"result": result,
		"op":     op,
		"a":      a,
		"b":      b,
	}, nil
}

func (p *CalculatorPlugin) toFloat(v interface{}) (float64, error) {
	switch val := v.(type) {
	case float64:
		return val, nil
	case int:
		return float64(val), nil
	case int64:
		return float64(val), nil
	case string:
		f, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return 0, fmt.Errorf("cannot convert %q to float", val)
		}
		return f, nil
	default:
		return 0, fmt.Errorf("unsupported type %T", v)
	}
}
