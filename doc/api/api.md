# PrimeMall 商城 API 文档（统一响应格式）

## 概述

本文档定义了 PrimeMall 商城前台所有 API 接口。系统采用**微服务架构**，统一网关负责路由、鉴权、限流、跨域。

- **基础路径**：`https://api.primemall.com`
- **字符编码**：UTF-8
- **请求/响应格式**：JSON
- **时间格式**：RFC3339，例如 `2026-04-09T15:04:05+08:00`
- **金额单位**：**分**（int64），所有金额字段均为整数分。
- **上传图像**：使用服务端签名（Post Policy）方案。

### 统一响应格式

所有接口响应均遵循以下结构：

```json
{
  "code": 200,
  "message": "success",
  "data": { ... }
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| code | int | 业务状态码，200 表示成功，其他值表示错误 |
| message | string | 提示信息 |
| data | object/array/null | 响应数据，成功时返回具体内容，失败时为 null |

错误响应格式相同，`code` 为非 200 值，`message` 包含错误描述。

---

## 认证说明

系统使用**双令牌**机制：

- **Access Token**：有效期 2 小时，放在 `Authorization: Bearer <access_token>` 头中访问业务接口。
- **Refresh Token**：有效期 7 天，仅用于调用刷新接口获取新的 Access Token。

当 Access Token 过期时，API 返回 `401` + `{"code":10002,"message":"token expired"}`。前端应使用 Refresh Token 调用刷新接口，获得新令牌后重试原请求。

---

## 1. 用户服务模块

### 1.1 获取短信验证码

**请求路径**：`POST /api/v1/user/captcha`

**请求体**：
```json
{
  "phone": "13800138000",
  "scene": "login"
}
```

**响应体**：
```json
{
  "code": 200,
  "message": "success",
  "data": null
}
```

**注意事项/实现细节**：
- 同一手机号 1 分钟内最多 1 次，同一 IP 每小时最多 10 次。
- 需先通过图形验证码获取 `captcha_token`（防短信轰炸）。
- 验证码存储于 Redis，key：`sms:code:{scene}:{phone}`，有效期 5 分钟。

---

### 1.2 手机号密码登录

**请求路径**：`POST /api/v1/user/login`

**请求体**：
```json
{
  "phone": "13800138000",
  "password": "明文密码"
}
```

**响应体**：
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIs...",
    "access_expire": 1744291200,
    "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
    "refresh_expire": 1744886400
  }
}
```

| data 字段 | 类型 | 说明 |
|-----------|------|------|
| access_token | string | 业务 API 访问令牌 |
| access_expire | int64 | Access Token 过期 Unix 时间戳（秒） |
| refresh_token | string | 刷新令牌 |
| refresh_expire | int64 | Refresh Token 过期时间戳（秒） |

**注意事项/实现细节**：
- 密码使用 bcrypt 加密存储和比对。
- JWT 签名使用 HMAC-SHA256，payload 中增加 `type: "access"` 或 `"refresh"` 字段区分。

---

### 1.3 手机号验证码登录

**请求路径**：`POST /api/v1/user/login/mobile`

**请求体**：
```json
{
  "phone": "13800138000",
  "code": "123456"
}
```

**响应体**：同 1.2。

---

### 1.4 用户注册

**请求路径**：`POST /api/v1/user/register`

**请求体**：
```json
{
  "phone": "13800138000",
  "password": "密码",
  "code": "123456",
  "nickname": "可选昵称"
}
```

**响应体**：同 1.2（注册后自动登录）。

---

### 1.5 刷新 Access Token

**请求路径**：`POST /api/v1/user/token/refresh`

**请求头**：`Authorization: Bearer <refresh_token>`

**请求体**：无

**响应体**：
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIs...",
    "access_expire": 1744294800,
    "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
    "refresh_expire": 1744886400
  }
}
```

**注意事项/实现细节**：
- 验证 Refresh Token 签名、有效期及 `type: "refresh"`。
- 若采用滚动刷新，同时颁发新的 Refresh Token。

---

### 1.6 忘记密码重置

**请求路径**：`POST /api/v1/user/password/reset`

**请求体**：
```json
{
  "phone": "13800138000",
  "code": "123456",
  "new_password": "新密码"
}
```

**响应体**：
```json
{
  "code": 200,
  "message": "success",
  "data": null
}
```

### 刷新token

**请求路径**：`POST /api/v1/user/token/refresh`

**响应体**：
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIs...",
    "access_expire": 1744294800,
    "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
    "refresh_expire": 1744886400
  }
}
```

---

### 1.7 获取个人资料（需认证）

**请求路径**：`GET /api/v1/user/info`

**请求头**：`Authorization: Bearer <access_token>`

**响应体**：
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": 10001,
    "phone": "13800138000",
    "nickname": "烤冷面",
    "avatar": "https://cdn.primemall.com/avatar/xxx.jpg",
    "gender": 1,
    "birthday": "1990-01-01",
    "status": 1,
    "status_desc": "正常"
  }
}
```

---

### 1.8 更新个人资料（需认证）

**请求路径**：`PUT /api/v1/user/info`

**请求体**（部分字段）：
```json
{
  "nickname": "新昵称",
  "avatar": "https://...",
  "gender": 2,
  "birthday": "1995-05-05"
}
```

**响应体**：
```json
{
  "code": 200,
  "message": "success",
  "data": null
}
```

---

### 1.9 修改登录密码（需认证）

**请求路径**：`PUT /api/v1/user/password/update`

**请求体**：
```json
{
  "old_password": "旧密码",
  "new_password": "新密码"
}
```

**响应体**：
```json
{
  "code": 200,
  "message": "success",
  "data": null
}
```

---

### 1.10 修改手机号（需认证）

**请求路径**：`PUT /api/v1/user/phone`

**请求体**：
```json
{
  "old_phone_code": "123456",
  "new_phone": "13900139000",
  "new_phone_code": "654321"
}
```

**响应体**：
```json
{
  "code": 200,
  "message": "success",
  "data": null
}
```

---

### 1.11 注销账号（需认证）

**请求路径**：`DELETE /api/v1/user/account`

**请求体**：
```json
{
  "password": "密码",
  "code": "验证码（可选）"
}
```

**响应体**：
```json
{
  "code": 200,
  "message": "success",
  "data": null
}
```

---

### 1.12 退出登录（需认证）

**请求路径**：`POST /api/v1/user/logout`

**请求头**：`Authorization: Bearer <access_token>`

**响应体**：
```json
{
  "code": 200,
  "message": "success",
  "data": null
}
```

---

### 1.13 添加收货地址（需认证）

**请求路径**：`POST /api/v1/address/`

**请求体**：
```json
{
  "receiver_name": "张三",
  "receiver_phone": "13800138000",
  "province": "广东省",
  "city": "深圳市",
  "district": "南山区",
  "detail_address": "科技园 xxx 大厦",
  "is_default": 1,
  "tag": "HOME",
  "postal_code": "518000"
}
```

**响应体**：
```json
{
  "code": 200,
  "message": "success",
  "data": null
}
```

---

### 1.14 获取地址列表（需认证）

**请求路径**：`GET /api/v1/address/list`

**响应体**：
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "list": [
      {
        "id": 101,
        "receiver_name": "张三",
        "receiver_phone": "13800138000",
        "province": "广东省",
        "city": "深圳市",
        "district": "南山区",
        "detail_address": "科技园 xxx 大厦",
        "is_default": 1,
        "tag": "HOME",
        "postal_code": "518000"
      }
    ]
  }
}
```

---

### 1.15 更新收货地址（需认证）

**请求路径**：`PUT /api/v1/address/`

**请求体**：
```json
{
  "id": 101,
  "receiver_name": "李四",
  "is_default": 0
}
```

**响应体**：
```json
{
  "code": 200,
  "message": "success",
  "data": null
}
```

---

### 1.16 删除收货地址（需认证）

**请求路径**：`DELETE /api/v1/address/:id`

**响应体**：
```json
{
  "code": 200,
  "message": "success",
  "data": null
}
```

---

### 1.17 设置默认收货地址（需认证）

**请求路径**：`PATCH /api/v1/address/default/:id`

**响应体**：
```json
{
  "code": 200,
  "message": "success",
  "data": null
}
```

---

## 2. 商品服务模块

### 2.1 获取全部分类树（公开）

**请求路径**：`GET /api/v1/product/category/tree`

**响应体**：
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "list": [
      {
        "id": 1,
        "parent_id": 0,
        "name": "手机通讯",
        "icon": "https://...",
        "level": 1,
        "children": [...]
      }
    ]
  }
}
```

---

### 2.2 获取单个分类详情（公开）

**请求路径**：`GET /api/v1/product/category/:id`

**响应体**：
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "category": {
      "id": 1,
      "parent_id": 0,
      "name": "手机通讯",
      "icon": "https://...",
      "level": 1
    }
  }
}
```

---

### 2.3 商品列表与筛选（公开）

**请求路径**：`GET /api/v1/product/list`

**请求参数**（Query）：

| 参数 | 类型 | 说明 |
|------|------|------|
| category_id | int64 | 分类 ID |
| brand | string | 品牌模糊匹配 |
| min_price | int64 | 最低价格（分） |
| max_price | int64 | 最高价格（分） |
| keyword | string | 商品名称关键词 |
| attrs | string | 属性过滤，如 `颜色:红色,内存:64G` |
| is_new | bool | 是否新品（7天内上架） |
| sort | string | `price_asc`, `price_desc`, `sales_desc`, `new_desc` |
| page | int64 | 默认 1 |
| size | int64 | 默认 10 |

**响应体**：
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "total": 100,
    "list": [
      {
        "id": 1001,
        "name": "iPhone 15 Pro",
        "brand": "Apple",
        "description": "钛金属，A17 Pro 芯片",
        "default_pic": "https://...",
        "price": 799900,
        "sales": 1234,
        "show_sales": 11234
      }
    ]
  }
}
```

---

### 2.4 商品详情（公开）

**请求路径**：`GET /api/v1/product/detail/:id`

**响应体**：
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": 1001,
    "category_id": 1,
    "brand": "Apple",
    "name": "iPhone 15 Pro",
    "description": "钛金属，A17 Pro 芯片",
    "content": "<p>详细描述...</p>",
    "default_pic": "https://...",
    "banner_pics": ["https://...", "https://..."],
    "video_url": "https://...",
    "weight": 221,
    "freight_template_id": 10,
    "price": 799900,
    "sales": 1234,
    "show_sales": 11234,
    "status": 1,
    "status_desc": "上架",
    "skus": [
      {
        "id": 2001,
        "name": "黑色 256G",
        "pic": "https://...",
        "price": 799900,
        "stock": 50,
        "spec_data": "{\"颜色\":\"黑色\",\"容量\":\"256G\"}"
      }
    ],
    "attributes": [
      { "name": "颜色", "values": ["黑色", "白色"] },
      { "name": "容量", "values": ["128G", "256G"] }
    ]
  }
}
```

---

### 2.5 个性化推荐（公开）

**请求路径**：`GET /api/v1/product/recommend`

**请求参数**（Query）：

| 参数 | 类型 | 说明 |
|------|------|------|
| scene | string | `home`、`cart`、`detail` |
| limit | int64 | 默认 10 |
| current_id | int64 | 当前商品ID（详情页推荐时） |

**响应体**：同 2.3 的 data 结构。

---

### 2.6 批量获取SKU价格与库存（公开）

**请求路径**：`POST /api/v1/product/skus/batch`

**请求体**：
```json
{
  "sku_ids": [2001, 2002]
}
```

**响应体**：
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "list": [
      { "id": 2001, "price": 799900, "stock": 50 },
      { "id": 2002, "price": 899900, "stock": 0 }
    ]
  }
}
```

---

## 3. 订单服务模块

### 3.1 添加/更新购物车（需认证）

**请求路径**：`POST /api/v1/order/cart/update`

**请求体**：
```json
{
  "sku_id": 2001,
  "quantity": 2
}
```

**响应体**：
```json
{
  "code": 200,
  "message": "success",
  "data": null
}
```

---

### 3.2 修改购物车勾选状态（需认证）

**请求路径**：`POST /api/v1/order/cart/select`

**请求体**：
```json
{
  "sku_id": 2001,
  "selected": true
}
```

**响应体**：
```json
{
  "code": 200,
  "message": "success",
  "data": null
}
```

---

### 3.3 获取购物车列表（需认证）

**请求路径**：`GET /api/v1/order/cart/list`

**响应体**：
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "items": [
      {
        "id": 1001,
        "sku_id": 2001,
        "spu_id": 1001,
        "product_name": "iPhone 15 Pro",
        "sku_name": "黑色 256G",
        "pic": "https://...",
        "price": 799900,
        "quantity": 1,
        "selected": true,
        "stock": 50,
        "status": 1
      }
    ],
    "total_amount": 799900
  }
}
```

---

### 3.4 批量删除购物车商品（需认证）

**请求路径**：`DELETE /api/v1/order/cart/remove`

**请求体**：
```json
{
  "sku_ids": [2001, 2002]
}
```

**响应体**：
```json
{
  "code": 200,
  "message": "success",
  "data": null
}
```

---

### 3.5 清空购物车（需认证）

**请求路径**：`DELETE /api/v1/order/cart/clear`

**响应体**：
```json
{
  "code": 200,
  "message": "success",
  "data": null
}
```

---

### 3.6 预下单（需认证）

**请求路径**：`POST /api/v1/order/pre-order`

**请求体**：
```json
{
  "items": [{"sku_id": 2001, "quantity": 1}],
  "address_id": 101,
  "coupon_id": 301
}
```

**响应体**：
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "settlement_token": "abc123",
    "items": [...],
    "address": {
      "receiver_name": "张三",
      "receiver_phone": "138...",
      "province": "广东省",
      "city": "深圳市",
      "district": "南山区",
      "detail_address": "..."
    },
    "total_amount": 799900,
    "freight_amount": 0,
    "coupon_amount": 10000,
    "pay_amount": 789900
  }
}
```

---

### 3.7 正式提交订单（需认证）

**请求路径**：`POST /api/v1/order/create`

**请求体**：
```json
{
  "settlement_token": "abc123",
  "address_id": 101,
  "coupon_id": 301,
  "pay_type": 1,
  "pay_amount": 789900,
  "remark": "请放快递柜",
  "idempotency_key": "uuid-xxxx"
}
```

**响应体**：
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "order_sn": "202604091234567890",
    "pay_amount": 789900
  }
}
```

---

### 3.8 订单列表（需认证）

**请求路径**：`GET /api/v1/order/list`

**请求参数**（Query）：

| 参数 | 类型 | 说明 |
|------|------|------|
| status | int64 | 0=全部，10=待付款，20=待发货，30=待收货，40=已完成，50=已取消，60=售后中 |
| page | int64 | 默认1 |
| size | int64 | 默认10 |

**响应体**：
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "total": 20,
    "list": [
      {
        "order_sn": "202604091234567890",
        "status": 10,
        "status_desc": "待付款",
        "pay_amount": 789900,
        "create_time": "2026-04-09T10:30:00+08:00",
        "items": [
          {
            "sku_id": 2001,
            "spu_id": 1001,
            "product_name": "iPhone 15 Pro",
            "sku_name": "黑色 256G",
            "pic": "https://...",
            "price": 799900,
            "quantity": 1,
            "total_amount": 799900
          }
        ]
      }
    ]
  }
}
```

---

### 3.9 订单详情（需认证）

**请求路径**：`GET /api/v1/order/detail/:order_sn`

**响应体**：
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "order_sn": "202604091234567890",
    "status": 10,
    "status_desc": "待付款",
    "pay_amount": 789900,
    "freight_amount": 0,
    "coupon_amount": 10000,
    "remark": "请放快递柜",
    "address": {
      "receiver_name": "张三",
      "receiver_phone": "138...",
      "province": "广东省",
      "city": "深圳市",
      "district": "南山区",
      "detail_address": "..."
    },
    "items": [...],
    "delivery_sn": "",
    "delivery_corp": "",
    "expire_time": 1744291200,
    "create_time": "2026-04-09T10:30:00+08:00"
  }
}
```

---

### 3.10 取消订单（需认证）

**请求路径**：`POST /api/v1/order/cancel/:order_sn`

**请求体**（可选）：
```json
{
  "cancel_reason": "不想要了"
}
```

**响应体**：
```json
{
  "code": 200,
  "message": "success",
  "data": null
}
```

---

### 3.11 确认收货（需认证）

**请求路径**：`POST /api/v1/order/confirm/:order_sn`

**响应体**：
```json
{
  "code": 200,
  "message": "success",
  "data": null
}
```

---

### 3.12 查询物流轨迹（需认证）

**请求路径**：`GET /api/v1/order/delivery/track/:order_sn`

**响应体**：
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "order_sn": "202604091234567890",
    "delivery_sn": "SF1234567890",
    "delivery_corp": "顺丰速运",
    "status": 300,
    "status_desc": "派送中",
    "tracks": [
      {
        "time": "2026-04-10T08:00:00+08:00",
        "station": "深圳科技园营业点",
        "status": "已揽收",
        "description": "快递员已取件"
      }
    ]
  }
}
```

---

### 3.13 申请售后（需认证）

**请求路径**：`POST /api/v1/order/aftersale/apply`

**请求体**：
```json
{
  "order_sn": "202604091234567890",
  "sku_id": 2001,
  "quantity": 1,
  "type": 1,
  "reason": "质量问题",
  "apply_amount": 799900,
  "images": ["https://..."]
}
```

**响应体**：
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "after_sale_id": 5001,
    "status": 1,
    "status_desc": "待审核",
    "apply_amount": 799900,
    "audit_remark": "",
    "created_at": "2026-04-09T11:00:00+08:00"
  }
}
```

---

### 3.14 查询售后详情（需认证）

**请求路径**：`GET /api/v1/order/aftersale/detail/:id`

**响应体**：同 3.13 的 data 结构。

---

### 3.15 取消售后申请（需认证）

**请求路径**：`POST /api/v1/order/aftersale/cancel/:id`

**响应体**：
```json
{
  "code": 200,
  "message": "success",
  "data": null
}
```

---

## 4. 支付服务模块

### 4.1 创建支付单（需认证）

**请求路径**：`POST /api/v1/payment/create`

**请求体**：
```json
{
  "order_sn": "202604091234567890",
  "channel": "wechat",
  "pay_type": "JSAPI",
  "amount": 789900,
  "openid": "oUpF8uMuAJO_M2pxb1Q9zNjWeS6o"
}
```

**响应体**：
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "payment_sn": "PAY202604090001",
    "pay_data": "prepay_id=wx123..."
  }
}
```

---

### 4.2 根据支付流水号查询支付状态（需认证）

**请求路径**：`GET /api/v1/payment/status/payment/:payment_sn`

**响应体**：
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "payment_sn": "PAY202604090001",
    "order_sn": "202604091234567890",
    "channel": "wechat",
    "channel_order_sn": "4200001234567890",
    "transaction_id": "4200001234567890",
    "status": 1,
    "status_desc": "成功",
    "amount": 789900,
    "pay_time": "2026-04-09T10:35:00+08:00",
    "error_msg": ""
  }
}
```

---

### 4.3 根据订单号查询支付状态（需认证）

**请求路径**：`GET /api/v1/payment/status/order/:order_sn`

**响应体**：同 4.2。

---

### 4.4 关闭支付单（需认证）

**请求路径**：`POST /api/v1/payment/close/:payment_sn`

**响应体**：
```json
{
  "code": 200,
  "message": "success",
  "data": null
}
```

---

### 4.5 申请退款（需认证）

**请求路径**：`POST /api/v1/payment/refund/apply`

**请求体**：
```json
{
  "order_sn": "202604091234567890",
  "payment_sn": "PAY202604090001",
  "amount": 789900,
  "reason": "质量问题",
  "type": 1,
  "images": ["https://..."]
}
```

**响应体**：
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "refund_sn": "REF202604090001",
    "status": 0,
    "status_desc": "处理中"
  }
}
```

---

### 4.6 查询退款详情（需认证）

**请求路径**：`GET /api/v1/payment/refund/detail/:order_sn`

**响应体**：
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "refund_sn": "REF202604090001",
    "payment_sn": "PAY202604090001",
    "order_sn": "202604091234567890",
    "amount": 789900,
    "status": 1,
    "status_desc": "退款成功",
    "refund_time": "2026-04-10T09:00:00+08:00",
    "transaction_id": "5000001234567890",
    "error_msg": ""
  }
}
```

---

### 4.7 支付宝异步回调（公开）

**请求路径**：`POST /api/v1/payment/callback/alipay`

**请求体**：支付宝 POST 表单数据

**响应体**：纯文本 `success`（非 JSON）

---

### 4.8 微信支付异步回调（公开）

**请求路径**：`POST /api/v1/payment/callback/wechat`

**请求体**：微信 XML 数据

**响应体**：`{"code":"SUCCESS"}`

---

## 5. 营销服务模块（优惠券）

### 5.1 领券中心列表（需认证）

**请求路径**：`GET /api/v1/marketing/coupon/list`

**请求参数**（Query）：`page`、`size`

**响应体**：
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "total": 50,
    "list": [
      {
        "id": 101,
        "name": "新人专享券",
        "type": 1,
        "threshold_amount": 10000,
        "reduce_amount": 1000,
        "discount_rate": 0,
        "max_discount_amount": 0,
        "start_time": "2026-04-01T00:00:00+08:00",
        "end_time": "2026-05-01T23:59:59+08:00",
        "description": "满100减10",
        "is_claimed": false
      }
    ]
  }
}
```

---

### 5.2 领取优惠券（需认证）

**请求路径**：`POST /api/v1/marketing/coupon/claim`

**请求体**：
```json
{
  "coupon_id": 101
}
```

**响应体**：
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "user_coupon_id": 1001,
    "expire_time": "2026-05-01T23:59:59+08:00"
  }
}
```

---

### 5.3 我的优惠券列表（需认证）

**请求路径**：`GET /api/v1/marketing/coupon/mine`

**请求参数**（Query）：
- `status`：0=未使用，1=已使用，2=已过期
- `page`、`size`

**响应体**：
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "total": 10,
    "list": [
      {
        "id": 1001,
        "coupon_id": 101,
        "name": "新人专享券",
        "type": 1,
        "status": 0,
        "expire_time": "2026-05-01T23:59:59+08:00",
        "order_sn": "",
        "unavailable_reason": "",
        "threshold_amount": 10000,
        "reduce_amount": 1000,
        "discount_rate": 0,
        "max_discount_amount": 0
      }
    ]
  }
}
```

---

### 5.4 结算页可用优惠券（需认证）

**请求路径**：`POST /api/v1/marketing/coupon/available`

**请求体**：
```json
{
  "order_amount": 799900
}
```

**响应体**：
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "available": [...],
    "unavailable": [...]
  }
}
```

---

### 5.5 使用优惠券（内部接口）

**请求路径**：`POST /internal/v1/coupon/use`

**请求体**：
```json
{
  "user_coupon_id": 1001,
  "order_sn": "202604091234567890"
}
```

**响应体**：
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "success": true,
    "discount_amount": 1000,
    "error_msg": ""
  }
}
```

---

### 5.6 解锁优惠券（内部接口）

**请求路径**：`POST /internal/v1/coupon/unlock`

**请求体**：
```json
{
  "user_coupon_id": 1001,
  "order_sn": "202604091234567890"
}
```

**响应体**：
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "success": true,
    "error_msg": ""
  }
}
```

---

## 附录：错误码说明

| 错误码 | 含义 |
|--------|------|
| 200 | 成功 |
| 其他 | 业务错误（具体定义由开发者自定义） |

错误响应示例：
```json
{
  "code": 10001,
  "message": "参数错误",
  "data": null
}
```