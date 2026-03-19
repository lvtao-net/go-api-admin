# 商城系统使用说明

## 概述

这是一个基于现有后台API构建的简单商城系统，使用纯HTML+JavaScript实现前端，通过API与后台交互。

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

## 访问方式

### 后台管理系统
- 地址: http://localhost:8099
- 默认管理员: admin@lvtao.net
- 默认密码: admin123

### 商城前台
- 地址: 直接在浏览器打开 `shop/index.html` 文件
- 或配置静态文件服务指向 shop 目录

## 数据库集合

系统自动创建了以下数据表：

| 表名 | 说明 | 类型 |
|------|------|------|
| members | 会员表 | auth |
| categories | 商品分类 | base |
| products | 商品表 | base |
| orders | 订单表 | base |
| wallets | 钱包表 | base |
| transactions | 交易记录 | base |

## API 接口

### 公开接口

#### 商品相关
```
GET  /api/collections/products/records     # 获取商品列表
GET  /api/collections/products/records/:id # 获取商品详情
GET  /api/collections/categories/records   # 获取分类列表
```

#### 会员认证
```
POST /api/collections/members/register          # 会员注册
POST /api/collections/members/auth-with-password # 会员登录
POST /api/collections/members/request-otp       # 请求验证码
POST /api/collections/members/reset-password    # 重置密码
POST /api/collections/members/auth-refresh      # 刷新Token
```

### 需要认证的接口

#### 订单相关
```
GET  /api/collections/orders/records     # 获取订单列表
GET  /api/collections/orders/records/:id # 获取订单详情
POST /api/collections/orders/records     # 创建订单
PATCH /api/collections/orders/records/:id # 更新订单状态
```

#### 钱包相关
```
GET  /api/collections/wallets/records      # 获取钱包信息
POST /api/collections/wallets/records      # 创建钱包
PATCH /api/collections/wallets/records/:id # 更新钱包余额
GET  /api/collections/transactions/records # 获取交易记录
POST /api/collections/transactions/records # 创建交易记录
```

#### 会员信息
```
GET  /api/collections/members/records/:id # 获取会员信息
PATCH /api/collections/members/records/:id # 更新会员信息
```

## 使用流程

### 1. 启动服务
```bash
cd /Users/lvtao/Documents/trae_projects/go-gin-api-admin
./api-server
```

### 2. 访问商城
在浏览器中打开 `shop/index.html` 文件

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

登录后台管理系统后，可以：
- 管理商品分类
- 管理商品信息（上下架、价格、库存等）
- 查看和管理订单
- 查看会员信息
- 管理系统设置

## 技术栈

- 后端: Go + Gin + GORM + MySQL
- 前端: HTML + CSS + JavaScript (原生)
- API: RESTful API
