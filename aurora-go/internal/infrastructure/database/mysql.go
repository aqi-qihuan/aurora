package database

import (
	"log/slog"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"github.com/aurora-go/aurora/internal/config"
)

var DB *gorm.DB

// InitMySQL 初始化 MySQL 数据库连接
func InitMySQL(cfg *config.MySQLConfig) {
	var err error

	DB, err = gorm.Open(mysql.Open(cfg.DSN()), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Info),
	})

	if err != nil {
		slog.Error("Failed to connect to MySQL", "error", err)
		panic("Failed to connect to MySQL: " + err.Error())
	}

	sqlDB, err := DB.DB()
	if err != nil {
		slog.Error("Failed to get sql.DB instance", "error", err)
		panic(err.Error())
	}

	// 连接池配置（对标Java版HikariCP优化参数）
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.MaxLifetime) * time.Second)

	slog.Info("MySQL connected successfully",
		"host", cfg.Host,
		"port", cfg.Port,
		"database", cfg.DBName,
		"max_open_conns", cfg.MaxOpenConns,
		"max_idle_conns", cfg.MaxIdleConns,
	)
}

// Close 关闭数据库连接
func Close() error {
	if DB == nil {
		return nil
	}
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// GetDB 获取数据库实例
func GetDB() *gorm.DB {
	return DB
}
