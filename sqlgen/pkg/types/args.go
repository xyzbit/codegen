package types

type Mode int

const (
	SQL Mode = iota
	GORM
	XORM
	SQLX
	BUN
)

const (
	MockSQLite = "sqlite"
	MockDocker = "docker"
)

// RunArg 代表运行参数，同时也用于配置文件的解析
type RunArg struct {
	// DSN 数据库连接字符串
	DSN string `yaml:"dsn"`
	// Filename SQL文件模式
	Filename []string `yaml:"filename"`
	// Table 要生成的表名模式
	Table []string `yaml:"table"`
	// Mode 生成模式（仅用于命令行）
	Mode Mode `yaml:"-"`
	// Output 适配器输出目录
	Output string `yaml:"output"`
	// EntityOutput 实体输出目录
	EntityOutput string `yaml:"entity_output"`
	// RepoOutput 仓库接口输出目录
	RepoOutput string `yaml:"repo_output"`
	// RepoPackage 仓库接口包名
	RepoPackage string `yaml:"repo_package"`
	// EntityPackage 实体包名
	EntityPackage string `yaml:"entity_package"`
	// AutoAudit 是否开启自动审计
	AutoAudit bool `yaml:"auto_audit"`
	// MockTypes 要生成的 mock 类型
	MockTypes []string `yaml:"mock_types"`
}

// DefaultRunArg 返回默认运行参数
func DefaultRunArg() RunArg {
	return RunArg{
		Table:        []string{"*"},
		Filename:     []string{"*.sql"},
		Output:       ".",
		EntityOutput: ".",
		RepoOutput:   ".",
		AutoAudit:    false,
	}
}
