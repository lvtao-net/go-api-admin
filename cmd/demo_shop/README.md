# 商城演示系统

这是一个基于现有后台API构建的简单商城演示系统，使用纯HTML+JavaScript实现前端。

## 目录结构

```
cmd/demo_shop/
├── main.go           # 商城服务入口
├── seed/             # 数据初始化
│   ├── shop_collections.go  # 创建商城集合
│   └── shop_data.go         # 创建示例数据
├── shop/             # 商城前端
│   ├── index.html    # 商品列表页
│   ├── login.html    # 登录页
│   ├── register.html # 注册页
│   ├── product.html  # 商品详情页
│   ├── orders.html   # 订单列表页
│   ├── wallet.html   # 钱包页面
│   ├── profile.html  # 个人中心
│   ├── css/          # 样式文件
│   └── js/           # JavaScript文件
└── README.md         # 本文件
```

## 功能特性

### 1. 会员功能
- ✅ 会员注册（支持邮箱/手机号）
- ✅ 会员登录
- ✅ 找回密码（通过验证码）
- ✅ 个人信息管理

### 2. 商品功能
- ✅ 商品列表展示
- ✅ 商品分类筛选
- ✅ 商品搜索
- ✅ 商品详情查看

### 3. 订单功能
- ✅ 商品下单
- ✅ 余额支付订单
- ✅ 确认收货
- ✅ 订单列表查看
- ✅ 订单详情查看
- ✅ 取消订单

### 4. 钱包功能
- ✅ 查看余额
- ✅ 账户充值
- ✅ 交易记录查看

## 启动方式

### 方式一：使用 go run

```bash
cd /Users/lvtao/Documents/trae_projects/go-gin-api-admin
go run ./cmd/demo_shop
```

### 方式二：编译后运行

```bash
cd /Users/lvtao/Documents/trae_projects/go-gin-api-admin
go build -o demo_shop ./cmd/demo_shop
./demo_shop
```

### 自定义端口

```bash
./demo_shop -port 3000
```

## 访问地址

启动后访问：

- **商城首页**: http://localhost:8080/shop
- **API接口**: http://localhost:8080/api

## 与后台管理系统的区别

| 项目 | 后台管理系统 | 商城演示系统 |
|------|-------------|-------------|
| 入口 | `cmd/server/main.go` | `cmd/demo_shop/main.go` |
| 端口 | 8099 | 8080（可配置） |
| 功能 | 后台管理界面 | 商城前台界面 |
| 静态文件 | web/dist | cmd/demo_shop/shop |

两个系统共用同一个数据库，可以同时运行。

## 数据库集合

系统自动创建以下数据表：

| 表名 | 说明 | 类型 |
|------|------|------|
| members | 会员表 | auth |
| categories | 商品分类 | base |
| products | 商品表 | base |
| orders | 订单表 | base |
| wallets | 钱包表 | base |
| transactions | 交易记录 | base |

## 使用流程

### 1. 启动服务
```bash
./demo_shop
```

### 2. 访问商城
浏览器打开 http://localhost:8080/shop

### 3. 注册会员
1. 点击"登录"按钮
2. 点击"注册账号"
3. 填写邮箱、密码、昵称
4. 点击"注册"按钮

### 4. 购物流程
1. 浏览商品列表
2. 点击商品查看详情
3. 选择数量，点击"立即购买"
4. 填写收货地址，确认下单
5. 在"我的订单"中支付订单

### 5. 充值余额
1. 点击"我的钱包"
2. 点击"充值"按钮
3. 选择或输入充值金额
4. 确认充值

## 注意事项

1. 验证码功能目前会在控制台打印，实际使用时需要配置邮件/短信服务
2. 图片使用 picsum.photos 提供的占位图
3. 充值功能为模拟充值，实际使用时需要对接支付接口

## 后台管理

如需管理商品、订单等数据，请启动后台管理系统：

```bash
go run ./cmd/server
```

访问 http://localhost:8099
- 账号：admin@lvtao.net
- 密码：admin123
