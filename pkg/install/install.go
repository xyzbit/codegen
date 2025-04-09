package install

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	// 规则文件名
	ruleFileName = "codegen.mdc"
)

// Command 实现 install 子命令
type Command struct{}

// NewCommand 创建新的 install 命令实例
func NewCommand() *Command {
	return &Command{}
}

// Run 执行 install 命令
func (c *Command) Run() error {
	// 获取当前工作目录
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("获取当前工作目录失败: %w", err)
	}

	// 确定 Cursor 规则目录（在当前项目目录下）
	cursorRulesDir := filepath.Join(currentDir, ".cursor", "rules")

	// 创建规则目录（如果不存在）
	if err := os.MkdirAll(cursorRulesDir, 0755); err != nil {
		return fmt.Errorf("创建规则目录失败: %w", err)
	}

	// 规则文件的完整路径
	ruleFilePath := filepath.Join(cursorRulesDir, ruleFileName)

	// 写入规则文件
	if err := os.WriteFile(ruleFilePath, []byte(DefaultRuleContent), 0644); err != nil {
		return fmt.Errorf("写入规则文件失败: %w", err)
	}

	fmt.Printf("成功安装 codegen 规则文件到: %s\n", ruleFilePath)
	return nil
}
