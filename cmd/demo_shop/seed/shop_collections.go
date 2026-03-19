package seed

import (
	"log"

	"github.com/lvtao/go-gin-api-admin/internal/model"
	"github.com/lvtao/go-gin-api-admin/internal/service"
)

// InitShopCollections 初始化商城集合
func InitShopCollections() {
	collectionService := service.NewCollectionService()

	// 1. 创建会员集合 (auth类型)
	createMembersCollection(collectionService)

	// 2. 创建商品分类集合
	createCategoriesCollection(collectionService)

	// 3. 创建商品集合
	createProductsCollection(collectionService)

	// 4. 创建订单集合
	createOrdersCollection(collectionService)

	// 5. 创建钱包集合
	createWalletsCollection(collectionService)

	// 6. 创建交易记录集合
	createTransactionsCollection(collectionService)

	// 7. 创建支付事务集合
	createPaymentTransaction(collectionService)

	// 8. 创建充值事务集合
	createRechargeTransaction(collectionService)

	log.Println("商城集合初始化完成！")
}

// 创建会员集合
func createMembersCollection(s *service.CollectionService) {
	// 检查是否已存在
	_, err := s.GetByName("members")
	if err == nil {
		log.Println("会员集合已存在，跳过创建")
		return
	}

	req := &service.CreateCollectionRequest{
		Name:  "members",
		Label: "会员",
		Type:  "auth",
		Fields: []model.CollectionField{
			{
				Name:        "nickname",
				Type:        "text",
				Label:       "昵称",
				Description: "会员昵称",
				Required:    false,
			},
			{
				Name:        "mobile",
				Type:        "text",
				Label:       "手机号",
				Description: "手机号码",
				Required:    false,
			},
			{
				Name:        "phone",
				Type:        "text",
				Label:       "联系电话",
				Description: "联系电话",
				Required:    false,
			},
			{
				Name:        "address",
				Type:        "text",
				Label:       "收货地址",
				Description: "默认收货地址",
				Required:    false,
			},
			{
				Name:        "avatar",
				Type:        "url",
				Label:       "头像",
				Description: "会员头像URL",
				Required:    false,
			},
		},
		CreateRule: ptrString(""), // 允许公开注册
	}

	_, err = s.Create(req)
	if err != nil {
		log.Printf("创建会员集合失败: %v", err)
	} else {
		log.Println("会员集合创建成功")
	}
}

// 创建商品分类集合
func createCategoriesCollection(s *service.CollectionService) {
	_, err := s.GetByName("categories")
	if err == nil {
		log.Println("商品分类集合已存在，跳过创建")
		return
	}

	req := &service.CreateCollectionRequest{
		Name:  "categories",
		Label: "商品分类",
		Type:  "base",
		Fields: []model.CollectionField{
			{
				Name:        "name",
				Type:        "text",
				Label:       "分类名称",
				Description: "分类名称",
				Required:    true,
			},
			{
				Name:         "sort",
				Type:         "number",
				Label:        "排序",
				Description:  "排序值，越小越靠前",
				Required:     false,
				DefaultValue: 0,
			},
			{
				Name:         "status",
				Type:         "select",
				Label:        "状态",
				Description:  "分类状态",
				Required:     true,
				DefaultValue: "active",
				FieldOptions: []model.FieldOption{
					{Label: "启用", Value: "active"},
					{Label: "禁用", Value: "inactive"},
				},
			},
			{
				Name:        "icon",
				Type:        "text",
				Label:       "图标",
				Description: "分类图标",
				Required:    false,
			},
		},
		ListRule:   ptrString(""),
		ViewRule:   ptrString(""),
		CreateRule: ptrString(""),
	}

	_, err = s.Create(req)
	if err != nil {
		log.Printf("创建商品分类集合失败: %v", err)
	} else {
		log.Println("商品分类集合创建成功")
	}
}

// 创建商品集合
func createProductsCollection(s *service.CollectionService) {
	_, err := s.GetByName("products")
	if err == nil {
		log.Println("商品集合已存在，跳过创建")
		return
	}

	req := &service.CreateCollectionRequest{
		Name:  "products",
		Label: "商品",
		Type:  "base",
		Fields: []model.CollectionField{
			{
				Name:        "name",
				Type:        "text",
				Label:       "商品名称",
				Description: "商品名称",
				Required:    true,
			},
			{
				Name:        "description",
				Type:        "editor",
				Label:       "商品描述",
				Description: "商品详细描述",
				Required:    false,
			},
			{
				Name:        "image",
				Type:        "url",
				Label:       "商品图片",
				Description: "商品主图URL",
				Required:    false,
			},
			{
				Name:        "images",
				Type:        "json",
				Label:       "商品图片列表",
				Description: "商品图片列表",
				Required:    false,
			},
			{
				Name:        "price",
				Type:        "number",
				Label:       "售价",
				Description: "商品售价",
				Required:    true,
				MinValue:    0,
			},
			{
				Name:        "originalPrice",
				Type:        "number",
				Label:       "原价",
				Description: "商品原价",
				Required:    false,
				MinValue:    0,
			},
			{
				Name:         "stock",
				Type:         "number",
				Label:        "库存",
				Description:  "商品库存",
				Required:     true,
				DefaultValue: 0,
				MinValue:     0,
			},
			{
				Name:         "sales",
				Type:         "number",
				Label:        "销量",
				Description:  "商品销量",
				Required:     false,
				DefaultValue: 0,
			},
			{
				Name:        "category",
				Type:        "relation",
				Label:       "分类",
				Description: "商品分类",
				Required:    false,
				RelationCollection: "categories",
				RelationField:      "id",
				RelationLabelField: "name",
				RelationType:       "belongs_to",
			},
			{
				Name:         "status",
				Type:         "select",
				Label:        "状态",
				Description:  "商品状态",
				Required:     true,
				DefaultValue: "active",
				FieldOptions: []model.FieldOption{
					{Label: "上架", Value: "active"},
					{Label: "下架", Value: "inactive"},
				},
			},
			{
				Name:         "sort",
				Type:         "number",
				Label:        "排序",
				Description:  "排序值，越小越靠前",
				Required:     false,
				DefaultValue: 0,
			},
		},
		ListRule:   ptrString("status = \"active\""),
		ViewRule:   ptrString("status = \"active\""),
		CreateRule: ptrString(""),
	}

	_, err = s.Create(req)
	if err != nil {
		log.Printf("创建商品集合失败: %v", err)
	} else {
		log.Println("商品集合创建成功")
	}
}

// 创建订单集合
func createOrdersCollection(s *service.CollectionService) {
	_, err := s.GetByName("orders")
	if err == nil {
		log.Println("订单集合已存在，跳过创建")
		return
	}

	req := &service.CreateCollectionRequest{
		Name:  "orders",
		Label: "订单",
		Type:  "base",
		Fields: []model.CollectionField{
			{
				Name:        "orderNo",
				Type:        "text",
				Label:       "订单号",
				Description: "订单编号",
				Required:    true,
				Unique:      true,
			},
			{
				Name:        "memberId",
				Type:        "relation",
				Label:       "会员",
				Description: "下单会员",
				Required:    true,
				RelationCollection: "members",
				RelationField:      "id",
				RelationLabelField: "nickname",
				RelationType:       "belongs_to",
			},
			{
				Name:        "productId",
				Type:        "relation",
				Label:       "商品",
				Description: "商品ID",
				Required:    true,
				RelationCollection: "products",
				RelationField:      "id",
				RelationLabelField: "name",
				RelationType:       "belongs_to",
			},
			{
				Name:        "productName",
				Type:        "text",
				Label:       "商品名称",
				Description: "商品名称快照",
				Required:    true,
			},
			{
				Name:        "productImage",
				Type:        "url",
				Label:       "商品图片",
				Description: "商品图片快照",
				Required:    false,
			},
			{
				Name:        "productPrice",
				Type:        "number",
				Label:       "商品单价",
				Description: "商品单价快照",
				Required:    true,
			},
			{
				Name:        "quantity",
				Type:        "number",
				Label:       "数量",
				Description: "购买数量",
				Required:    true,
				MinValue:    1,
			},
			{
				Name:        "totalAmount",
				Type:        "number",
				Label:       "订单金额",
				Description: "订单总金额",
				Required:    true,
			},
			{
				Name:         "status",
				Type:         "select",
				Label:        "订单状态",
				Description:  "订单状态",
				Required:     true,
				DefaultValue: "pending",
				FieldOptions: []model.FieldOption{
					{Label: "待支付", Value: "pending"},
					{Label: "已支付", Value: "paid"},
					{Label: "已发货", Value: "shipped"},
					{Label: "已完成", Value: "completed"},
					{Label: "已取消", Value: "cancelled"},
				},
			},
			{
				Name:        "address",
				Type:        "text",
				Label:       "收货地址",
				Description: "收货地址",
				Required:    true,
			},
			{
				Name:        "remark",
				Type:        "text",
				Label:       "备注",
				Description: "订单备注",
				Required:    false,
			},
			{
				Name:        "paidAt",
				Type:        "date",
				Label:       "支付时间",
				Description: "支付时间",
				Required:    false,
			},
			{
				Name:        "shippedAt",
				Type:        "date",
				Label:       "发货时间",
				Description: "发货时间",
				Required:    false,
			},
			{
				Name:        "completedAt",
				Type:        "date",
				Label:       "完成时间",
				Description: "完成时间",
				Required:    false,
			},
		},
		ListRule:   ptrString("memberId = @request.auth.id"),
		ViewRule:   ptrString("memberId = @request.auth.id"),
		CreateRule: ptrString(""),
	}

	_, err = s.Create(req)
	if err != nil {
		log.Printf("创建订单集合失败: %v", err)
	} else {
		log.Println("订单集合创建成功")
	}
}

// 创建钱包集合
func createWalletsCollection(s *service.CollectionService) {
	_, err := s.GetByName("wallets")
	if err == nil {
		log.Println("钱包集合已存在，跳过创建")
		return
	}

	req := &service.CreateCollectionRequest{
		Name:  "wallets",
		Label: "钱包",
		Type:  "base",
		Fields: []model.CollectionField{
			{
				Name:        "memberId",
				Type:        "relation",
				Label:       "会员",
				Description: "所属会员",
				Required:    true,
				Unique:      true,
				RelationCollection: "members",
				RelationField:      "id",
				RelationLabelField: "nickname",
				RelationType:       "has_one",
			},
			{
				Name:         "balance",
				Type:         "number",
				Label:        "余额",
				Description:  "账户余额",
				Required:     true,
				DefaultValue: 0,
				MinValue:     0,
			},
		},
		ListRule:   ptrString("memberId = @request.auth.id"),
		ViewRule:   ptrString("memberId = @request.auth.id"),
		CreateRule: ptrString(""),
	}

	_, err = s.Create(req)
	if err != nil {
		log.Printf("创建钱包集合失败: %v", err)
	} else {
		log.Println("钱包集合创建成功")
	}
}

// 创建交易记录集合
func createTransactionsCollection(s *service.CollectionService) {
	_, err := s.GetByName("transactions")
	if err == nil {
		log.Println("交易记录集合已存在，跳过创建")
		return
	}

	req := &service.CreateCollectionRequest{
		Name:  "transactions",
		Label: "交易记录",
		Type:  "base",
		Fields: []model.CollectionField{
			{
				Name:        "memberId",
				Type:        "relation",
				Label:       "会员",
				Description: "所属会员",
				Required:    true,
				RelationCollection: "members",
				RelationField:      "id",
				RelationLabelField: "nickname",
				RelationType:       "belongs_to",
			},
			{
				Name:         "type",
				Type:         "select",
				Label:        "交易类型",
				Description:  "交易类型",
				Required:     true,
				DefaultValue: "recharge",
				FieldOptions: []model.FieldOption{
					{Label: "充值", Value: "recharge"},
					{Label: "消费", Value: "payment"},
					{Label: "退款", Value: "refund"},
				},
			},
			{
				Name:        "amount",
				Type:        "number",
				Label:       "金额",
				Description: "交易金额（正数为收入，负数为支出）",
				Required:    true,
			},
			{
				Name:        "balance",
				Type:        "number",
				Label:       "余额",
				Description: "交易后余额",
				Required:    true,
			},
			{
				Name:        "description",
				Type:        "text",
				Label:       "描述",
				Description: "交易描述",
				Required:    false,
			},
			{
				Name:        "relatedId",
				Type:        "text",
				Label:       "关联ID",
				Description: "关联记录ID（如订单ID）",
				Required:    false,
			},
		},
		ListRule:   ptrString("memberId = @request.auth.id"),
		ViewRule:   ptrString("memberId = @request.auth.id"),
		CreateRule: ptrString(""),
	}

	_, err = s.Create(req)
	if err != nil {
		log.Printf("创建交易记录集合失败: %v", err)
	} else {
		log.Println("交易记录集合创建成功")
	}
}

// 创建支付事务集合
func createPaymentTransaction(s *service.CollectionService) {
	_, err := s.GetByName("payment")
	if err == nil {
		log.Println("支付事务集合已存在，跳过创建")
		return
	}

	req := &service.CreateCollectionRequest{
		Name:        "payment",
		Label:       "订单支付",
		Type:        "transaction",
		Description: "订单支付事务：验证订单 -> 检查余额 -> 扣款 -> 创建交易记录 -> 更新订单状态",
		Fields: []model.CollectionField{
			{
				Name:        "orderId",
				Type:        "number",
				Label:       "订单ID",
				Description: "要支付的订单ID",
				Required:    true,
			},
		},
		TransactionSteps: []model.TransactionStep{
			{
				Name:  "查询订单",
				Type:  "query",
				Table: "orders",
				Alias: "order",
				Conditions: []model.TransactionCondition{
					{Field: "id", ValueFrom: "params.orderId"},
					{Field: "memberId", ValueFrom: "user.id"},
					{Field: "status", Value: "pending"},
				},
				Required: true,
				Error:    "订单不存在或状态不允许支付",
			},
			{
				Name:  "查询钱包",
				Type:  "query",
				Table: "wallets",
				Alias: "wallet",
				Conditions: []model.TransactionCondition{
					{Field: "memberId", ValueFrom: "user.id"},
				},
				Required: true,
				Error:    "钱包不存在，请先创建钱包",
			},
			{
				Name:              "验证余额",
				Type:              "validate",
				ValidateCondition: "${wallet.balance} >= ${order.totalAmount}",
				Error:             "余额不足",
			},
			{
				Name:  "扣除余额",
				Type:  "update",
				Table: "wallets",
				Conditions: []model.TransactionCondition{
					{Field: "id", ValueFrom: "wallet.id"},
				},
				Data: map[string]interface{}{
					"balance": "${wallet.balance - order.totalAmount}",
				},
			},
			{
				Name:  "创建交易记录",
				Type:  "insert",
				Table: "transactions",
				Alias: "transaction",
				Data: map[string]interface{}{
					"memberId":    "${user.id}",
					"type":        "payment",
					"amount":      "-${order.totalAmount}",
					"balance":     "${wallet.balance - order.totalAmount}",
					"description": "订单支付: ${order.orderNo}",
					"relatedId":   "${order.id}",
				},
			},
			{
				Name:  "更新订单状态",
				Type:  "update",
				Table: "orders",
				Conditions: []model.TransactionCondition{
					{Field: "id", ValueFrom: "order.id"},
				},
				Data: map[string]interface{}{
					"status": "paid",
				},
			},
		},
	}

	_, err = s.Create(req)
	if err != nil {
		log.Printf("创建支付事务集合失败: %v", err)
	} else {
		log.Println("支付事务集合创建成功")
	}
}

// 创建充值事务集合
func createRechargeTransaction(s *service.CollectionService) {
	_, err := s.GetByName("recharge")
	if err == nil {
		log.Println("充值事务集合已存在，跳过创建")
		return
	}

	req := &service.CreateCollectionRequest{
		Name:        "recharge",
		Label:       "账户充值",
		Type:        "transaction",
		Description: "账户充值事务：验证金额 -> 获取/创建钱包 -> 更新余额 -> 创建交易记录",
		Fields: []model.CollectionField{
			{
				Name:        "amount",
				Type:        "number",
				Label:       "充值金额",
				Description: "充值的金额",
				Required:    true,
				MinValue:    0.01,
				MaxValue:    100000,
			},
		},
		TransactionSteps: []model.TransactionStep{
			{
				Name:              "验证金额",
				Type:              "validate",
				ValidateCondition: "${params.amount} > 0",
				Error:             "充值金额必须大于0",
			},
			{
				Name:  "查询钱包",
				Type:  "query",
				Table: "wallets",
				Alias: "wallet",
				Conditions: []model.TransactionCondition{
					{Field: "memberId", ValueFrom: "user.id"},
				},
				Required: false, // 钱包可以不存在
			},
			{
				Name:  "创建钱包（如果不存在）",
				Type:  "insert",
				Table: "wallets",
				Alias: "newWallet",
				Data: map[string]interface{}{
					"memberId": "${user.id}",
					"balance":  "${params.amount}",
				},
				OnError: "skip", // 如果钱包已存在则跳过
			},
			{
				Name:  "更新钱包余额（如果钱包存在）",
				Type:  "update",
				Table: "wallets",
				Conditions: []model.TransactionCondition{
					{Field: "id", ValueFrom: "wallet.id"},
				},
				Data: map[string]interface{}{
					"balance": "${wallet.balance + params.amount}",
				},
				OnError: "skip", // 如果钱包不存在则跳过
			},
			{
				Name:  "创建交易记录",
				Type:  "insert",
				Table: "transactions",
				Data: map[string]interface{}{
					"memberId":    "${user.id}",
					"type":        "recharge",
					"amount":      "${params.amount}",
					"balance":     "${wallet.balance + params.amount}",
					"description": "账户充值",
				},
			},
		},
	}

	_, err = s.Create(req)
	if err != nil {
		log.Printf("创建充值事务集合失败: %v", err)
	} else {
		log.Println("充值事务集合创建成功")
	}
}

// ptrString 返回字符串指针
func ptrString(s string) *string {
	return &s
}
