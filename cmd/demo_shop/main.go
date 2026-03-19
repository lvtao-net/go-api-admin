package main

import (
	"flag"
	"fmt"
	"html/template"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/lvtao/go-gin-api-admin/cmd/demo_shop/seed"
	"github.com/lvtao/go-gin-api-admin/internal/config"
	"github.com/lvtao/go-gin-api-admin/internal/model"
	"github.com/lvtao/go-gin-api-admin/internal/router"
	"github.com/lvtao/go-gin-api-admin/internal/service"
	"github.com/lvtao/go-gin-api-admin/pkg/database"
	"github.com/lvtao/go-gin-api-admin/pkg/logger"
	"go.uber.org/zap"
)

var (
	configPath string
	port       int
)

func init() {
	flag.StringVar(&configPath, "config", "config.yaml", "config file path")
	flag.IntVar(&port, "port", 8080, "shop server port")
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

	logger.Info("Starting Shop Demo...",
		zap.String("name", "Shop Demo"),
		zap.Int("port", port),
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

	// 初始化商城集合
	seed.InitShopCollections()

	// 初始化商城示例数据
	seed.InitShopData()

	// 设置 Gin 模式
	if cfg.App.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建 Gin 引擎
	r := gin.New()

	// 设置模板函数
	r.SetFuncMap(template.FuncMap{
		"safe": func(str string) template.HTML { return template.HTML(str) },
	})

	// 设置 API 路由（包含必要的中间件）
	router.Setup(r)

	// 配置商城静态文件服务
	shopDir := filepath.Join(".", "shop")
	if _, err := os.Stat(shopDir); os.IsNotExist(err) {
		// 如果当前目录没有shop，尝试cmd/demo_shop/shop
		shopDir = filepath.Join("cmd", "demo_shop", "shop")
	}

	logger.Info("Serving shop static files", zap.String("dir", shopDir))

	// 静态文件路由
	r.Static("/shop/assets", filepath.Join(shopDir, "assets"))
	r.Static("/shop/css", filepath.Join(shopDir, "css"))
	r.Static("/shop/js", filepath.Join(shopDir, "js"))

	// 商城页面路由
	r.GET("/shop", func(c *gin.Context) {
		c.File(filepath.Join(shopDir, "index.html"))
	})
	r.GET("/shop/login.html", func(c *gin.Context) {
		c.File(filepath.Join(shopDir, "login.html"))
	})
	r.GET("/shop/register.html", func(c *gin.Context) {
		c.File(filepath.Join(shopDir, "register.html"))
	})
	r.GET("/shop/forgot-password.html", func(c *gin.Context) {
		c.File(filepath.Join(shopDir, "forgot-password.html"))
	})
	r.GET("/shop/index.html", func(c *gin.Context) {
		c.File(filepath.Join(shopDir, "index.html"))
	})
	r.GET("/shop/product.html", func(c *gin.Context) {
		c.File(filepath.Join(shopDir, "product.html"))
	})
	r.GET("/shop/orders.html", func(c *gin.Context) {
		c.File(filepath.Join(shopDir, "orders.html"))
	})
	r.GET("/shop/wallet.html", func(c *gin.Context) {
		c.File(filepath.Join(shopDir, "wallet.html"))
	})
	r.GET("/shop/profile.html", func(c *gin.Context) {
		c.File(filepath.Join(shopDir, "profile.html"))
	})

	// 启动服务器
	addr := fmt.Sprintf(":%d", port)
	logger.Info(fmt.Sprintf("Shop server starting on %s", addr))
	logger.Info(fmt.Sprintf("Shop URL: http://localhost:%d/shop", port))
	logger.Info(fmt.Sprintf("Admin API: http://localhost:%d/api", port))

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

	logger.Info("Shutting down shop server...")
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
