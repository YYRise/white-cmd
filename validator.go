package whitecmd

import (
	"errors"
	"slices"
	"strings"

	shellwords "github.com/YYRise/white-cmd/parse"
)

// Validator 白名单验证器
type Validator struct {
	config *Config // 白名单配置
}

// NewValidator 创建验证器实例
func NewValidator(configPath string) (*Validator, error) {
	cfg, err := LoadConfig(configPath)
	if err != nil {
		return nil, err
	}
	return &Validator{config: cfg}, nil
}

// Validate 验证单个命令是否符合白名单规则
func (v *Validator) Validate(cmd string) (bool, error) {
	cmd = strings.TrimSpace(cmd)
	if cmd == "" {
		return false, errors.New("empty command")
	}

	args, err := shellwords.Parse(cmd)
	if err != nil {
		return false, err
	}
	if len(args) == 0 {
		return false, errors.New("invalid command format")
	}

	// 基础命令检查
	baseCmd := strings.ToLower(args[0])
	whiteArgs, exists := v.config.Commands[baseCmd]
	if !exists {
		return false, errors.New("command is not white")
	}

	// 参数规则检查
	return v.validateArgs(args[1:], whiteArgs), nil
}

// validateArgs 验证命令参数是否符合白名单规则
func (v *Validator) validateArgs(args []string, whiteArgs []string) bool {
	if len(whiteArgs) == 0 && len(args) > 1 {
		return false // 无参数白名单但有额外参数
	}

	if len(whiteArgs) == 1 && whiteArgs[0] == "*" {
		return true // 允许所有参数
	}

	for _, arg := range args {
		arg = strings.TrimSpace(arg)
		if len(arg) == 0 || !strings.HasPrefix(arg, "-") {
			continue
		}

		// 处理带=的参数（如--param=value）
		param := strings.SplitN(arg, "=", 2)[0]
		if slices.Index(whiteArgs, param) < 0 {
			return false
		}

	}
	return true
}
