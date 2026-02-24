package modao

import (
	"fmt"
	"os"
	"path/filepath"

	"gorm.io/gorm"
)

// OutputPath 生成文件路径
var OutputPath string

// Config 链式配置器
type Config struct {
	debugKey DebugKey
	gormDbs  map[ConnectInfo]*gorm.DB
}

// Init 应用配置到全局注册中心
func (c *Config) Init() *Config {
	RegisterDebugKey(c.debugKey)

	for name, db := range c.gormDbs {
		RegisterGormDb(name, db)
	}
	return c
}

// SetDebugKey 设置调试密钥
func (c *Config) SetDebugKey(key DebugKey) *Config {
	c.debugKey = key
	return c
}

// SetGormDb 注册 GORM 数据库连接
func (c *Config) SetGormDb(connectionInfo ConnectInfo, db *gorm.DB) *Config {
	if c.gormDbs == nil {
		c.gormDbs = make(map[ConnectInfo]*gorm.DB)
	}
	c.gormDbs[connectionInfo] = db
	return c
}

// SetGenMdPath 设置生成 md 文件路径
func (c *Config) SetGenMdPath(path string) *Config {
	if path == "" {
		panic("FATAL: 输出路径不能为空 (示例: /app/output)")
	}

	if !filepath.IsAbs(path) {
		panic(fmt.Sprintf(
			"FATAL: 输出路径必须是绝对路径！\n"+
				"  当前值: %q\n"+
				"  修正建议: filepath.Join(os.Getwd(), \"output\")\n"+
				"  示例: /home/project/output (Linux) 或 C:\\project\\output (Windows)",
			path,
		))
	}

	if err := os.MkdirAll(path, 0755); err != nil {
		panic(fmt.Sprintf("FATAL: 无法创建输出目录 [%s]: %v", path, err))
	}

	OutputPath = path

	return c
}

// NewConfig 创建配置实例
func NewConfig() *Config {
	return &Config{}
}
