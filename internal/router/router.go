package router

import (
	"github.com/gin-gonic/gin"
	"github.com/lvtao/go-gin-api-admin/internal/handler"
	"github.com/lvtao/go-gin-api-admin/internal/middleware"
	"github.com/lvtao/go-gin-api-admin/pkg/validator"
)

func Setup(r *gin.Engine) {
	// 全局中间件
	r.Use(gin.Recovery())
	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.XSSMiddleware()) // XSS 防护
	r.Use(middleware.LoggerMiddleware())
	r.Use(middleware.OperationLogMiddleware())

	// 健康检查
	healthHandler := handler.NewHealthHandler()
	r.GET("/", healthHandler.Index)
	r.GET("/api/health", healthHandler.Health)

	// API 路由组
	api := r.Group("/api")

	// 添加速率限制中间件（默认启用）
	api.Use(middleware.RateLimitMiddleware())

	{
		// ==================== 公开接口 ====================

		// 管理员认证（公开）
		adminHandler := handler.NewAdminHandler()
		api.POST("/admins/auth-with-password", adminHandler.Login)
		api.POST("/admins/auth-refresh", adminHandler.RefreshToken)
		api.GET("/admins/auth-methods", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"code":    0,
				"message": "success",
				"data": gin.H{
					"emailPassword": gin.H{
						"enabled": true,
					},
				},
			})
		})

		// 文件上传（公开）- 只保留单文件上传
		fileHandler := handler.NewFileHandler()
		api.POST("/files/upload", fileHandler.Upload)

		// API 文档（公开）
		apiDocHandler := handler.NewAPIDocHandler()
		api.GET("/doc", apiDocHandler.GetDoc)

		// ==================== 前台用户 API（受 API 权限限制）====================

		// 记录 CRUD（前台用户，受 API 权限限制）
		// Base/Auth 类型: /api/collections/:collection/records
		// View 类型: /api/collections/:collection (GET)
		// Transaction 类型: /api/collections/:collection (POST)
		recordHandler := handler.NewRecordHandler()
		collectionAPI := api.Group("/collections/:collection")
		collectionAPI.Use(middleware.OptionalAuthMiddleware())
		{
			// Base/Auth: CRUD 操作
			collectionAPI.GET("/records", recordHandler.List)
			collectionAPI.GET("/records/:id", recordHandler.Get)              // 通过主键查找
			collectionAPI.GET("/records/by/:field/:value", recordHandler.GetByField) // 通过指定字段查找
			collectionAPI.POST("/records", recordHandler.Create)
			collectionAPI.PATCH("/records/:id", recordHandler.Update)
			collectionAPI.DELETE("/records/:id", recordHandler.Delete)

			// View: 查询
			collectionAPI.GET("", recordHandler.ViewQuery)

			// Transaction: 执行
			collectionAPI.POST("", recordHandler.TransactionExecute)
		}

		// Auth 集合认证路由（前台用户）- 简化
		authHandler := handler.NewAuthHandler()
		auth := api.Group("/collections/:collection")
		{
			// 注册
			auth.POST("/register", authHandler.Register)
			// 登录
			auth.POST("/auth-with-password", authHandler.AuthWithPassword)
			// 刷新Token
			auth.POST("/auth-refresh", authHandler.AuthRefresh)
			// 请求验证码（用于注册验证邮箱、找回密码等）
			auth.POST("/request-otp", authHandler.RequestOTP)
			// 重置密码（通过验证码）
			auth.POST("/reset-password", authHandler.ResetPassword)
		}

		// ==================== 后台管理 API（不受 API 权限限制）====================

		// 管理员认证路由
		adminAuth := api.Group("/admins")
		adminAuth.Use(middleware.AdminAuthMiddleware())
		{
			adminAuth.GET("/profile", adminHandler.GetProfile)
			adminAuth.GET("/stats", adminHandler.GetStats)
			adminAuth.POST("/change-password", adminHandler.ChangePassword)
			adminAuth.GET("", adminHandler.List)
			adminAuth.POST("", adminHandler.Create)
			adminAuth.DELETE("/:id", adminHandler.Delete)
			adminAuth.PATCH("/:id", adminHandler.Update)
		}

		// 后台管理路由组
		manage := api.Group("/manage")
		manage.Use(middleware.AdminAuthMiddleware())
		{
			// 集合管理
			collectionHandler := handler.NewCollectionHandler()
			manage.GET("/collections", collectionHandler.List)
			manage.POST("/collections", collectionHandler.Create)
			manage.GET("/collections/:collection", collectionHandler.GetByName)
			manage.PATCH("/collections/:collection", collectionHandler.UpdateByName)
			manage.DELETE("/collections/:collection", collectionHandler.DeleteByName)
			manage.GET("/collections/:collection/check-delete", collectionHandler.CheckDelete)
			manage.GET("/collections/:collection/preview", collectionHandler.PreviewView)

			// 记录 CRUD（后台管理，不受 API 权限限制）
			adminRecordHandler := handler.NewAdminRecordHandler()
			manage.GET("/collections/:collection/records", adminRecordHandler.List)
			manage.GET("/collections/:collection/records/:id", adminRecordHandler.Get)
			manage.POST("/collections/:collection/records", adminRecordHandler.Create)
			manage.PATCH("/collections/:collection/records/:id", adminRecordHandler.Update)
			manage.DELETE("/collections/:collection/records/:id", adminRecordHandler.Delete)
			manage.POST("/collections/:collection/records/batch-delete", adminRecordHandler.BatchDelete)
			manage.GET("/collections/:collection/fields", adminRecordHandler.GetCollectionFields)

			// 备份管理
			backupHandler := handler.NewBackupHandler()
			manage.POST("/backups", backupHandler.CreateBackup)
			manage.POST("/backups/restore", backupHandler.RestoreBackup)
			manage.GET("/backups", backupHandler.ListBackups)
			manage.DELETE("/backups/:filename", backupHandler.DeleteBackup)

			// 操作日志
			logHandler := handler.NewLogHandler()
			manage.GET("/logs", logHandler.List)
			manage.GET("/logs/stats", logHandler.GetStats)
			manage.DELETE("/logs", logHandler.DeleteOldLogs)

			// 邮件模板
			emailTemplateHandler := handler.NewEmailTemplateHandler()
			manage.GET("/email-templates", emailTemplateHandler.List)
			manage.GET("/email-templates/:type", emailTemplateHandler.Get)
			manage.POST("/email-templates", emailTemplateHandler.Create)
			manage.PATCH("/email-templates/:type", emailTemplateHandler.Update)
			manage.POST("/email-templates/:type/test", emailTemplateHandler.Test)

			// 字典管理
			dictionaryHandler := handler.NewDictionaryHandler()
			manage.GET("/dictionaries", dictionaryHandler.List)
			manage.GET("/dictionaries/:id", dictionaryHandler.Get)
			manage.GET("/dictionaries/name/:name", dictionaryHandler.GetByName)
			manage.POST("/dictionaries", dictionaryHandler.Create)
			manage.PATCH("/dictionaries/:id", dictionaryHandler.Update)
			manage.DELETE("/dictionaries/:id", dictionaryHandler.Delete)
			manage.GET("/dictionaries/:id/items", dictionaryHandler.GetItems)
			manage.POST("/dictionaries/:id/items", dictionaryHandler.CreateItem)
			manage.PATCH("/dictionaries/:id/items/:itemId", dictionaryHandler.UpdateItem)
			manage.DELETE("/dictionaries/:id/items/:itemId", dictionaryHandler.DeleteItem)

			// 验证规则列表
			manage.GET("/validation-rules", func(c *gin.Context) {
				c.JSON(200, gin.H{
					"code":    0,
					"message": "success",
					"data": gin.H{
						"rules":      validator.GetBuiltInRules(),
						"byCategory": validator.GetRuleByCategory(),
					},
				})
			})
		}
	}
}
