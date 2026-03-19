package seed

import (
	"log"

	"github.com/lvtao/go-gin-api-admin/pkg/database"
)

// InitShopData 初始化商城示例数据
func InitShopData() {
	db := database.GetDB()

	// 检查是否已有商品数据
	var productCount int64
	db.Table("products").Count(&productCount)
	if productCount > 0 {
		log.Println("商品数据已存在，跳过初始化")
		return
	}

	// 创建商品分类
	categories := []map[string]interface{}{
		{"name": "数码产品", "sort": 1, "status": "active", "icon": "📱"},
		{"name": "服装鞋帽", "sort": 2, "status": "active", "icon": "👕"},
		{"name": "食品饮料", "sort": 3, "status": "active", "icon": "🍔"},
		{"name": "家居用品", "sort": 4, "status": "active", "icon": "🏠"},
		{"name": "图书文具", "sort": 5, "status": "active", "icon": "📚"},
	}

	categoryIDs := make([]uint64, 0)
	for _, cat := range categories {
		result := db.Table("categories").Create(cat)
		if result.Error != nil {
			log.Printf("创建分类失败: %v", result.Error)
			continue
		}
		var lastID uint64
		db.Raw("SELECT LAST_INSERT_ID()").Scan(&lastID)
		categoryIDs = append(categoryIDs, lastID)
	}
	log.Printf("创建了 %d 个商品分类", len(categoryIDs))

	// 创建示例商品
	products := []map[string]interface{}{
		{
			"name":          "iPhone 15 Pro Max",
			"description":   "Apple iPhone 15 Pro Max，A17 Pro芯片，钛金属设计，4800万像素主摄，支持USB-C接口。",
			"image":         "https://picsum.photos/seed/iphone/400/400",
			"price":         9999.00,
			"originalPrice": 10999.00,
			"stock":         100,
			"sales":         258,
			"category":      categoryIDs[0],
			"status":        "active",
			"sort":          1,
		},
		{
			"name":          "MacBook Pro 14寸",
			"description":   "Apple MacBook Pro 14英寸，M3 Pro芯片，18GB统一内存，512GB固态硬盘，Liquid Retina XDR显示屏。",
			"image":         "https://picsum.photos/seed/macbook/400/400",
			"price":         14999.00,
			"originalPrice": 16999.00,
			"stock":         50,
			"sales":         128,
			"category":      categoryIDs[0],
			"status":        "active",
			"sort":          2,
		},
		{
			"name":          "AirPods Pro 2",
			"description":   "Apple AirPods Pro第二代，主动降噪，自适应通透模式，个性化空间音频，MagSafe充电盒。",
			"image":         "https://picsum.photos/seed/airpods/400/400",
			"price":         1899.00,
			"originalPrice": 1999.00,
			"stock":         200,
			"sales":         568,
			"category":      categoryIDs[0],
			"status":        "active",
			"sort":          3,
		},
		{
			"name":          "纯棉T恤男款",
			"description":   "100%纯棉面料，舒适透气，多色可选，简约百搭款式。",
			"image":         "https://picsum.photos/seed/tshirt/400/400",
			"price":         99.00,
			"originalPrice": 199.00,
			"stock":         500,
			"sales":         1256,
			"category":      categoryIDs[1],
			"status":        "active",
			"sort":          1,
		},
		{
			"name":          "休闲运动鞋",
			"description":   "轻便舒适运动鞋，透气网面设计，防滑耐磨大底，多尺码可选。",
			"image":         "https://picsum.photos/seed/shoes/400/400",
			"price":         299.00,
			"originalPrice": 499.00,
			"stock":         300,
			"sales":         892,
			"category":      categoryIDs[1],
			"status":        "active",
			"sort":          2,
		},
		{
			"name":          "牛仔裤男款",
			"description":   "经典直筒牛仔裤，优质牛仔面料，舒适弹力，多尺码可选。",
			"image":         "https://picsum.photos/seed/jeans/400/400",
			"price":         199.00,
			"originalPrice": 399.00,
			"stock":         400,
			"sales":         756,
			"category":      categoryIDs[1],
			"status":        "active",
			"sort":          3,
		},
		{
			"name":          "进口零食大礼包",
			"description":   "精选全球进口零食，包含日韩欧美多种美味，超值组合装。",
			"image":         "https://picsum.photos/seed/snacks/400/400",
			"price":         128.00,
			"originalPrice": 198.00,
			"stock":         200,
			"sales":         2345,
			"category":      categoryIDs[2],
			"status":        "active",
			"sort":          1,
		},
		{
			"name":          "精品咖啡豆",
			"description":   "精选阿拉比卡咖啡豆，中度烘焙，香气浓郁，口感醇厚，500g装。",
			"image":         "https://picsum.photos/seed/coffee/400/400",
			"price":         88.00,
			"originalPrice": 128.00,
			"stock":         150,
			"sales":         567,
			"category":      categoryIDs[2],
			"status":        "active",
			"sort":          2,
		},
		{
			"name":          "有机绿茶",
			"description":   "高山有机绿茶，明前采摘，清香回甘，250g礼盒装。",
			"image":         "https://picsum.photos/seed/tea/400/400",
			"price":         168.00,
			"originalPrice": 268.00,
			"stock":         100,
			"sales":         345,
			"category":      categoryIDs[2],
			"status":        "active",
			"sort":          3,
		},
		{
			"name":          "简约台灯",
			"description":   "LED护眼台灯，三档调光，触控开关，简约设计，适合书房卧室。",
			"image":         "https://picsum.photos/seed/lamp/400/400",
			"price":         89.00,
			"originalPrice": 159.00,
			"stock":         300,
			"sales":         678,
			"category":      categoryIDs[3],
			"status":        "active",
			"sort":          1,
		},
		{
			"name":          "记忆棉枕头",
			"description":   "慢回弹记忆棉枕头，人体工学设计，有效缓解颈椎压力，助眠好帮手。",
			"image":         "https://picsum.photos/seed/pillow/400/400",
			"price":         128.00,
			"originalPrice": 198.00,
			"stock":         200,
			"sales":         456,
			"category":      categoryIDs[3],
			"status":        "active",
			"sort":          2,
		},
		{
			"name":          "多功能收纳盒",
			"description":   "大容量收纳盒，可折叠设计，多格分类，适合衣物杂物收纳。",
			"image":         "https://picsum.photos/seed/box/400/400",
			"price":         39.00,
			"originalPrice": 69.00,
			"stock":         500,
			"sales":         1234,
			"category":      categoryIDs[3],
			"status":        "active",
			"sort":          3,
		},
		{
			"name":          "《深入理解计算机系统》",
			"description":   "程序员必读经典，从程序员视角全面讲解计算机系统原理，第三版中文版。",
			"image":         "https://picsum.photos/seed/book1/400/400",
			"price":         98.00,
			"originalPrice": 139.00,
			"stock":         100,
			"sales":         567,
			"category":      categoryIDs[4],
			"status":        "active",
			"sort":          1,
		},
		{
			"name":          "《Go语言设计与实现》",
			"description":   "深入剖析Go语言设计与实现原理，适合Go语言进阶学习。",
			"image":         "https://picsum.photos/seed/book2/400/400",
			"price":         79.00,
			"originalPrice": 99.00,
			"stock":         80,
			"sales":         234,
			"category":      categoryIDs[4],
			"status":        "active",
			"sort":          2,
		},
		{
			"name":          "精美笔记本套装",
			"description":   "A5尺寸笔记本3本装，加厚纸张，精美封面，适合学习办公记录。",
			"image":         "https://picsum.photos/seed/notebook/400/400",
			"price":         29.00,
			"originalPrice": 49.00,
			"stock":         300,
			"sales":         890,
			"category":      categoryIDs[4],
			"status":        "active",
			"sort":          3,
		},
	}

	for _, product := range products {
		result := db.Table("products").Create(product)
		if result.Error != nil {
			log.Printf("创建商品失败: %v", result.Error)
		}
	}

	log.Printf("创建了 %d 个示例商品", len(products))
}
