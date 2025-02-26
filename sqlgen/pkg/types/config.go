package types

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// LoadConfig 从文件加载配置
func LoadConfig(filename string) (*RunArg, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	config := DefaultRunArg()
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	return &config, nil
}
