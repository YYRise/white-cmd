package whitecmd

import (
	"os"
	"gopkg.in/yaml.v3"
)

// Config 白名单配置结构
type Config struct {
	Commands map[string][]string `yaml:"commands"` // 命令: 允许的参数列表
}

// LoadConfig 从文件加载配置
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
