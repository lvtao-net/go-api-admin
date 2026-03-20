package database

import (
	"fmt"
	"time"

	"github.com/lvtao/go-gin-api-admin/internal/config"
	"github.com/lvtao/go-gin-api-admin/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

var DB *gorm.DB

func Init() error {
	cfg := config.GetConfig()
	dbCfg := cfg.Database

	// 数据库日志配置
	var gormLogger gormlogger.Interface
	if dbCfg.LogMode {
		gormLogger = gormlogger.Default.LogMode(gormlogger.Info)
	} else {
		gormLogger = gormlogger.Default.LogMode(gormlogger.Silent)
	}

	// 连接数据库
	db, err := gorm.Open(mysql.Open(dbCfg.DSN()), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return fmt.Errorf("failed to connect database: %w", err)
	}

	// 获取底层 sql.DB
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}

	// 设置连接池
	sqlDB.SetMaxIdleConns(dbCfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(dbCfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Hour)

	DB = db

	logger.Info("Database connected successfully",
		zap.String("host", dbCfg.Host),
		zap.Int("port", dbCfg.Port),
		zap.String("database", dbCfg.Name),
	)

	return nil
}

func GetDB() *gorm.DB {
	return DB
}

func Close() error {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}
