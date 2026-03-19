package main

import (
	"flag"
	"fmt"
	"html/template"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/lvtao/go-gin-api-admin/internal/config"
	"github.com/lvtao/go-gin-api-admin/internal/embed"
	"github.com/lvtao/go-gin-api-admin/internal/model"
	"github.com/lvtao/go-gin-api-admin/internal/router"
	"github.com/lvtao/go-gin-api-admin/internal/service"
	"github.com/lvtao/go-gin-api-admin/pkg/database"
	"github.com/lvtao/go-gin-api-admin/pkg/logger"
	"go.uber.org/zap"
)

var (
	configPath string
	version    string
)

func init() {
	flag.StringVar(&configPath, "config", "config.yaml", "config file path")
	flag.StringVar(&version, "version", "1.0.0", "application version")
}

func main() {
	flag.Parse()

	// 加载配置
	if err := config.Init(configPath); err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}
	cfg := config.GetConfig()

	// 初始化日志
	if err := logger.Init(); err != nil {
		fmt.Printf("Failed to init logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("Starting application...",
		zap.String("name", cfg.App.Name),
		zap.String("version", version),
	)

	// 初始化数据库
	if err := database.Init(); err != nil {
		logger.Fatal("Failed to init database", zap.Error(err))
	}
	defer database.Close()

	// 自动迁移系统表
	if err := autoMigrate(); err != nil {
		logger.Fatal("Failed to migrate database", zap.Error(err))
	}
	logger.Info("Database migrated successfully")

	// 初始化默认管理员
	if err := service.EnsureDefaultAdmin(); err != nil {
		logger.Fatal("Failed to ensure default admin", zap.Error(err))
	}
	logger.Info("Default admin ensured")

	// 初始化系统集合
	if err := service.InitSystemCollections(); err != nil {
		logger.Warn("Failed to init system collections", zap.Error(err))
	}

	// 初始化默认设置
	if err := service.NewSettingService().InitDefaultSettings(); err != nil {
		logger.Warn("Failed to init default settings", zap.Error(err))
	}

	// 初始化默认邮件模板
	if err := service.NewEmailTemplateService().InitDefaultTemplates(); err != nil {
		logger.Warn("Failed to init email templates", zap.Error(err))
	}

	// 初始化系统字典
	if err := service.NewDictionaryService().InitSystemDictionaries(); err != nil {
		logger.Warn("Failed to init system dictionaries", zap.Error(err))
	}

	// 设置 Gin 模式
	if cfg.App.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建 Gin 引擎
	r := gin.New()
	r.Use(gin.LoggerWithWriter(os.Stdout))
	r.Use(gin.Recovery())

	// 设置模板函数
	r.SetFuncMap(template.FuncMap{
		"safe": func(str string) template.HTML { return template.HTML(str) },
	})

	// 设置路由
	router.Setup(r)

	// 配置静态文件服务（支持嵌入模式）
	useEmbed := cfg.Embed.Enabled && embed.IsEmbedded()
	if useEmbed {
		logger.Info("Using embedded static files")
	} else {
		logger.Info("Using external static files from ./web/dist")
	}
	embed.ServeStatic(r, useEmbed)

	// 启动服务器
	addr := fmt.Sprintf(":%d", cfg.App.Port)
	logger.Info(fmt.Sprintf("Server starting on %s", addr))

	// 优雅关闭
	go func() {
		if err := r.Run(addr); err != nil {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")
}

func autoMigrate() error {
	return database.GetDB().AutoMigrate(
		&model.Collection{},
		&model.Admin{},
		&model.Setting{},
		&model.OperationLog{},
		&model.EmailTemplate{},
		&model.Dictionary{},
	)
}
