# PrimeMall 商城 API 文档（统一响应格式）

## 概述

本文档定义了 PrimeMall 商城**前台用户**和**管理后台**的所有 API 接口。系统采用**微服务架构**，统一网关负责路由、鉴权、限流、跨域。

- **前台基础路径**：`https://api.primemall.com`
- **后台基础路径**：`https://admin.primemall.com`（或同域名 `/api/v1/admin`）
- **字符编码**：UTF-8
- **请求/响应格式**：JSON
- **时间格式**：RFC3339，例如 `2026-04-09T15:04:05+08:00`
- **金额单位**：**分**（int64），所有金额字段均为整数分。

### 统一响应格式

所有接口响应均遵循以下结构：

```json
{
  "code": 200,
  "message": "success",
  "data": { ... }
}
```

| 字段      | 类型                | 说明                       |
| ------- | ----------------- | ------------------------ |
| code    | int               | 业务状态码，200 表示成功，其他值表示错误   |
| message | string            | 提示信息                     |
| data    | object/array/null | 响应数据，成功时返回具体内容，失败时为 null |

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

**请求头**：`Content-Type: application/json`

**请求体**：

```json
{
  "phone": "13800138000",
  "scene": "login",
  "idempotency_key": "uuid-xxxx"
}
```

| 字段              | 类型     | 必填  | 说明                                                                |
| --------------- | ------ | --- | ----------------------------------------------------------------- |
| phone           | string | 是   | 手机号（11位）                                                          |
| scene           | string | 是   | 场景：`login`, `register`, `reset_pwd`, `update_pwd`, `update_phone` |
| idempotency_key | string | 是   | 幂等键（客户端UUID，防止重复发送）                                               |

**响应体**：

```json
{
  "code": 200,
  "message": "success",
  "data": null
}
```

**注意事项**：

- 同一手机号 1 分钟内最多 1 次，同一 IP 每小时最多 10 次
- 验证码存储于 Redis，key：`sms:code:{scene}:{phone}`，有效期 5 分钟

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
    "access_expire": "2026-04-09T17:00:00+08:00",
    "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
    "refresh_expire": "2026-04-16T10:00:00+08:00"
  }
}
```

| data 字段        | 类型     | 说明                            |
| -------------- | ------ | ----------------------------- |
| access_token   | string | 业务 API 访问令牌                   |
| access_expire  | string | Access Token 过期时间（RFC3339格式）  |
| refresh_token  | string | 刷新令牌                          |
| refresh_expire | string | Refresh Token 过期时间（RFC3339格式） |

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
  "idempotency_key": "uuid-xxxx"
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
    "access_expire": "2026-04-09T17:00:00+08:00",
    "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
    "refresh_expire": "2026-04-16T10:00:00+08:00"
  }
}
```

---

### 1.6 忘记密码重置

**请求路径**：`POST /api/v1/user/password/reset`

**请求体**：

```json
{
  "phone": "13800138000",
  "code": "123456",
  "new_password": "新密码",
  "idempotency_key": "uuid-xxxx"
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

### 1.7 OSS 上传回调

**请求路径**：`POST /api/v1/user/upload/avatar/callback`

**说明**：阿里云 OSS 上传成功后的服务端回调接口（幂等）。

---

### 1.8 获取个人资料（需认证）

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

| data 字段     | 类型     | 说明                    |
| ----------- | ------ | --------------------- |
| id          | uint64 | 用户ID                  |
| phone       | string | 手机号                   |
| nickname    | string | 昵称                    |
| avatar      | string | 头像URL                 |
| gender      | int64  | 性别：0-未知，1-男，2-女       |
| birthday    | string | 生日（日期格式）              |
| status      | int64  | 账号状态：1-正常，2-限制下单，3-封禁 |
| status_desc | string | 状态描述                  |

---

### 1.9 更新个人资料（需认证）

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

### 1.10 修改登录密码（需认证）

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

### 1.11 修改手机号（需认证）

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

### 1.12 注销账号（需认证）

**请求路径**：`DELETE /api/v1/user/account`

**请求体**：

```json
{
  "password": "密码",
  "refresh_token": "refresh_token内容"
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

### 1.13 退出登录（需认证）

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

**注意事项**：

- 将 token 加入黑名单，Access Token 失效

---

### 1.14 获取 OSS 上传凭证（需认证）

**请求路径**：`POST /api/v1/user/upload/avatar/token`

**请求体**：

```json
{
  "file_name": "avatar.jpg"
}
```

**响应体**：

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "accessid": "LTAI5t...",
    "host": "https://primemall.oss-cn-shenzhen.aliyuncs.com",
    "policy": "eyJleHBpcmF0aW9uIjoi...",
    "signature": "eRMV...",
    "expire": "2026-04-09T17:00:00+08:00",
    "dir": "avatar/",
    "object_key": "avatar/xxx.jpg",
    "callback": "..."
  }
}
```

---

### 1.15 添加收货地址（需认证）

**请求路径**：`POST /api/v1/address/`

**请求体**：

```json
{
  "receiver_name": "张三",
  "receiver_phone": "13800138000",
  "is_default": 1,
  "tag": "HOME",
  "detail": {
    "province": "广东省",
    "city": "深圳市",
    "district": "南山区",
    "detail_address": "科技园 xxx 大厦",
    "postal_code": "518000"
  }
}
```

| 字段                    | 类型     | 必填  | 说明                              |
| --------------------- | ------ | --- | ------------------------------- |
| receiver_name         | string | 是   | 收货人姓名（1-50字符）                   |
| receiver_phone        | string | 是   | 收货人电话（11位）                      |
| is_default            | int64  | 是   | 是否默认地址：0-否，1-是                  |
| tag                   | string | 否   | 地址标签：`HOME`, `OFFICE`, `SCHOOL` |
| detail.province       | string | 是   | 省份                              |
| detail.city           | string | 是   | 城市                              |
| detail.district       | string | 是   | 区/县                             |
| detail.detail_address | string | 是   | 详细地址（1-200字符）                   |
| detail.postal_code    | string | 否   | 邮政编码                            |

**响应体**：

```json
{
  "code": 200,
  "message": "success",
  "data": null
}
```

---

### 1.16 获取地址列表（需认证）

**请求路径**：`GET /api/v1/address/list`

**请求参数**（Query）：

| 参数   | 类型    | 必填  | 说明              |
| ---- | ----- | --- | --------------- |
| page | int64 | 否   | 页码，默认1          |
| size | int64 | 否   | 每页数量，默认10，最大100 |

**响应体**：

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "total": 5,
    "list": [
      {
        "id": 101,
        "receiver_name": "张三",
        "receiver_phone": "13800138000",
        "is_default": 1,
        "tag": "HOME",
        "detail": {
          "province": "广东省",
          "city": "深圳市",
          "district": "南山区",
          "detail_address": "科技园 xxx 大厦",
          "postal_code": "518000"
        }
      }
    ]
  }
}
```

---

### 1.17 更新收货地址（需认证）

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

### 1.18 删除收货地址（需认证）

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

### 1.19 设置默认收货地址（需认证）

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
        "children": [
          {
            "id": 2,
            "parent_id": 1,
            "name": "手机",
            "icon": "https://...",
            "level": 2,
            "children": []
          }
        ]
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
    "id": 1,
    "parent_id": 0,
    "name": "手机通讯",
    "icon": "https://...",
    "level": 1
  }
}
```

---

### 2.3 商品列表与筛选（公开）

**请求路径**：`GET /api/v1/product/list`

**请求参数**（Query）：

| 参数          | 类型     | 必填  | 说明                                                     |
| ----------- | ------ | --- | ------------------------------------------------------ |
| category_id | int64  | 否   | 分类 ID                                                  |
| brand       | string | 否   | 品牌模糊匹配                                                 |
| min_price   | int64  | 否   | 最低价格（分）                                                |
| max_price   | int64  | 否   | 最高价格（分）                                                |
| keyword     | string | 否   | 商品名称关键词                                                |
| is_new      | bool   | 否   | 是否新品（7天内上架）                                            |
| sort        | string | 否   | 排序：`price_asc`, `price_desc`, `sales_desc`, `new_desc` |
| page        | int64  | 否   | 默认 1                                                   |
| size        | int64  | 否   | 默认 10                                                  |

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
        "desc": "钛金属，A17 Pro 芯片",
        "main_pic": "https://...",
        "price": 799900,
        "sales_count": 1234,
        "virtual_sales": 11234
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
    "desc": "钛金属，A17 Pro 芯片",
    "content": "<p>详细描述...</p>",
    "main_pic": "https://...",
    "sub_pics": ["https://...", "https://..."],
    "video_url": "https://...",
    "status": 1,
    "status_desc": "上架",
    "sales_count": 1234,
    "virtual_sales": 11234,
    "freight_template_id": 10,
    "skus": [
      {
        "id": 2001,
        "spu_id": 1001,
        "sku_code": "SKU001",
        "spec_data": "{\"颜色\":\"黑色\",\"容量\":\"256G\"}",
        "images": ["https://..."],
        "price": 799900,
        "market_price": 899900,
        "stock": 50,
        "locked_stock": 5,
        "weight": 221,
        "status": 1
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

| 参数         | 类型     | 必填  | 说明                                    |
| ---------- | ------ | --- | ------------------------------------- |
| scene      | string | 否   | 场景：`home`, `cart`, `detail`，默认 `home` |
| limit      | int64  | 否   | 返回数量，默认 10                            |
| current_id | int64  | 否   | 当前商品ID（详情页推荐时）                        |

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
      {
        "id": 2001,
        "price": 799900,
        "market_price": 899900,
        "stock": 50,
        "locked_stock": 5,
        "status": 1
      },
      {
        "id": 2002,
        "price": 899900,
        "market_price": 999900,
        "stock": 0,
        "locked_stock": 0,
        "status": 2
      }
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

| 字段       | 类型     | 必填  | 说明                   |
| -------- | ------ | --- | -------------------- |
| sku_id   | uint64 | 是   | SKU ID               |
| quantity | int64  | 是   | 数量（大于0为添加/更新，等于0为删除） |

**响应体**：

```json
{
  "code": 200,
  "message": "success",
  "data": null
}
```

**注意事项**：

- 数量为0时删除该SKU，幂等操作

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
  "items": [
    {"sku_id": 2001, "quantity": 1}
  ],
  "address_id": 101,
  "coupon_id": 301
}
```

| 字段               | 类型     | 必填  | 说明            |
| ---------------- | ------ | --- | ------------- |
| items            | array  | 是   | 下单商品列表（1-50个） |
| items[].sku_id   | uint64 | 是   | SKU ID        |
| items[].quantity | int64  | 是   | 购买数量          |
| address_id       | uint64 | 否   | 地址ID（可选）      |
| coupon_id        | uint64 | 否   | 优惠券ID（可选）     |

**响应体**：

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "settlement_token": "abc123...",
    "items": [
      {
        "sku_id": 2001,
        "spu_id": 1001,
        "product_name": "iPhone 15 Pro",
        "sku_name": "黑色 256G",
        "pic": "https://...",
        "price": 799900,
        "quantity": 1,
        "total_amount": 799900,
        "status": 1
      }
    ],
    "address": {
      "receiver_name": "张三",
      "receiver_phone": "138...",
      "detail": {
        "province": "广东省",
        "city": "深圳市",
        "district": "南山区",
        "detail_address": "科技园 xxx 大厦"
      }
    },
    "total_amount": 799900,
    "freight_amount": 0,
    "coupon_amount": 10000,
    "pay_amount": 789900
  }
}
```

**注意事项**：

- 生成结算令牌，有效期 15 分钟
- 计算运费、优惠券折扣等

---

### 3.7 正式提交订单（需认证）

**请求路径**：`POST /api/v1/order/create`

**请求体**：

```json
{
  "settlement_token": "abc123...",
  "address_id": 101,
  "coupon_id": 301,
  "pay_type": 1,
  "remark": "请放快递柜",
  "idempotency_key": "uuid-xxxx",
  "cart_sku_ids": [2001, 2002]
}
```

| 字段               | 类型     | 必填  | 说明               |
| ---------------- | ------ | --- | ---------------- |
| settlement_token | string | 是   | 结算令牌             |
| address_id       | uint64 | 是   | 地址ID             |
| coupon_id        | uint64 | 否   | 优惠券ID            |
| pay_type         | int64  | 否   | 支付方式：1-微信，2-支付宝  |
| remark           | string | 否   | 订单备注（最多200字符）    |
| idempotency_key  | string | 是   | 幂等键（前端生成）        |
| cart_sku_ids     | array  | 否   | 购物车商品ID列表（直购时使用） |

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

| 参数     | 类型    | 必填  | 说明                                                  |
| ------ | ----- | --- | --------------------------------------------------- |
| status | int64 | 否   | 订单状态：0=全部，10=待付款，20=已付款，30=已发货，40=已完成，50=已取消，60=售后中 |
| page   | int64 | 否   | 默认1                                                 |
| size   | int64 | 否   | 默认10                                                |

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
    ],
    "address": {
      "receiver_name": "张三",
      "receiver_phone": "138...",
      "detail": {
        "province": "广东省",
        "city": "深圳市",
        "district": "南山区",
        "detail_address": "科技园 xxx 大厦"
      }
    },
    "freight_amount": 0,
    "coupon_amount": 10000,
    "remark": "请放快递柜",
    "delivery_sn": "",
    "delivery_corp": "",
    "expire_time": "2026-04-09T10:45:00+08:00",
    "pay_time": "",
    "delivery_time": "",
    "finish_time": "",
    "cancel_time": ""
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

**注意事项**：

- 仅待支付状态可取消
- 取消后释放库存、解锁优惠券

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
  "images": ["https://..."],
  "idempotency_key": "uuid-xxxx"
}
```

| 字段              | 类型     | 必填  | 说明                |
| --------------- | ------ | --- | ----------------- |
| order_sn        | string | 是   | 订单号               |
| sku_id          | uint64 | 是   | SKU ID            |
| quantity        | int64  | 是   | 退款数量（支持部分退款）      |
| type            | int64  | 是   | 售后类型：1-仅退款，2-退货退款 |
| reason          | string | 是   | 售后原因（1-200字符）     |
| apply_amount    | int64  | 是   | 申请退款金额（分）         |
| images          | array  | 否   | 凭证图片URL列表（最多9张）   |
| idempotency_key | string | 是   | 幂等键               |

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
    "create_time": "2026-04-09T11:00:00+08:00"
  }
}
```

---

### 3.14 查询售后详情（需认证）

**请求路径**：`GET /api/v1/order/aftersale/detail/:id`

**响应体**：

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "after_sale_id": 5001,
    "order_sn": "202604091234567890",
    "after_sale_type": 1,
    "status": 1,
    "status_desc": "待审核",
    "apply_amount": 799900,
    "reason": "质量问题",
    "images": ["https://..."],
    "audit_remark": "",
    "create_time": "2026-04-09T11:00:00+08:00",
    "audit_time": "",
    "return_tracking_sn": "",
    "return_tracking_corp": "",
    "receive_time": "",
    "refund_amount": 0,
    "refund_time": "",
    "order_item": {
      "sku_id": 2001,
      "spu_id": 1001,
      "product_name": "iPhone 15 Pro",
      "sku_name": "黑色 256G",
      "pic": "https://...",
      "price": 799900,
      "quantity": 1,
      "total_amount": 799900
    }
  }
}
```

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

**注意事项**：

- 仅待审核状态可取消

---

### 3.16 获取售后单列表（需认证）

**请求路径**：`GET /api/v1/order/aftersale/list`

**请求参数**（Query）：

| 参数     | 类型    | 必填  | 说明                                         |
| ------ | ----- | --- | ------------------------------------------ |
| status | int64 | 否   | 售后单状态：1-待审核，2-待退货，3-退款中，4-已完成，5-已拒绝，不传则查全部 |
| page   | int64 | 否   | 默认1                                        |
| size   | int64 | 否   | 默认10                                       |

**响应体**：

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "total": 5,
    "list": [
      {
        "after_sale_id": 5001,
        "order_sn": "202604091234567890",
        "type": 1,
        "status": 1,
        "status_desc": "待审核",
        "refund_amount": 799900,
        "product_name": "iPhone 15 Pro",
        "sku_name": "黑色 256G",
        "pic": "https://..."
      }
    ]
  }
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
  "openid": "oUpF8uMuAJO_M2pxb1Q9zNjWeS6o",
  "idempotency_key": "uuid-xxxx"
}
```

| 字段              | 类型     | 必填   | 说明                              |
| --------------- | ------ | ---- | ------------------------------- |
| order_sn        | string | 是    | 订单号                             |
| channel         | string | 是    | 支付渠道：`alipay`, `wechat`         |
| pay_type        | string | 是    | 支付场景：`APP`, `H5`, `JSAPI`, `QR` |
| amount          | int64  | 是    | 应付金额（分），用于后端校验                  |
| openid          | string | 条件必填 | 微信JSAPI支付时必传                    |
| idempotency_key | string | 是    | 幂等键                             |

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

| status | 说明   |
| ------ | ---- |
| 0      | 待支付  |
| 1      | 支付成功 |
| 2      | 支付失败 |
| 3      | 退款中  |
| 4      | 退款成功 |

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

**注意事项**：

- 仅未支付时可关闭

---

### 4.5 申请退款（需认证）

**请求路径**：`POST /api/v1/payment/refund/apply`

**请求体**：

```json
{
  "order_sn": "202604091234567890",
  "amount": 789900,
  "reason": "质量问题",
  "idempotency_key": "uuid-xxxx"
}
```

| 字段              | 类型     | 必填  | 说明      |
| --------------- | ------ | --- | ------- |
| order_sn        | string | 是   | 订单号     |
| amount          | int64  | 是   | 退款金额（分） |
| reason          | string | 是   | 退款原因    |
| idempotency_key | string | 是   | 幂等键     |

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

**说明**：支付宝服务器异步通知回调（幂等）

**响应体**：纯文本 `success`

---

### 4.8 微信支付异步回调（公开）

**请求路径**：`POST /api/v1/payment/callback/wechat`

**说明**：微信支付服务器异步通知回调（幂等）

**响应体**：`{"code":"SUCCESS"}`

---

## 5. 营销服务模块（优惠券）

### 5.1 领券中心列表（公开）

**请求路径**：`GET /api/v1/marketing/coupon/list`

**请求参数**（Query）：

| 参数   | 类型    | 必填  | 说明   |
| ---- | ----- | --- | ---- |
| page | int64 | 否   | 默认1  |
| size | int64 | 否   | 默认10 |

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

| type | 说明   |
| ---- | ---- |
| 1    | 满减券  |
| 2    | 折扣券  |
| 3    | 无门槛券 |

---

### 5.2 领取优惠券（需认证）

**请求路径**：`POST /api/v1/marketing/coupon/claim`

**请求体**：

```json
{
  "coupon_id": 101,
  "idempotency_key": "uuid-xxxx"
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

| 参数     | 类型    | 必填  | 说明                       |
| ------ | ----- | --- | ------------------------ |
| status | int64 | 否   | 状态：0-未使用，1-已使用，2-已过期，默认0 |
| page   | int64 | 否   | 默认1                      |
| size   | int64 | 否   | 默认10                     |

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
    "available": [
      {
        "id": 1001,
        "coupon_id": 101,
        "name": "新人专享券",
        "type": 1,
        "status": 0,
        "expire_time": "2026-05-01T23:59:59+08:00",
        "threshold_amount": 10000,
        "reduce_amount": 1000,
        "discount_rate": 0,
        "max_discount_amount": 0
      }
    ],
    "unavailable": [
      {
        "id": 1002,
        "coupon_id": 102,
        "name": "满200减20",
        "type": 1,
        "status": 0,
        "expire_time": "2026-05-01T23:59:59+08:00",
        "threshold_amount": 20000,
        "reduce_amount": 2000,
        "discount_rate": 0,
        "max_discount_amount": 0,
        "unavailable_reason": "订单金额不满足"
      }
    ]
  }
}
```

---

### 5.5 使用优惠券（内部接口）

**请求路径**：`POST /internal/v1/coupon/use`

**说明**：订单服务调用，锁定并扣减优惠券（幂等）

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

**说明**：订单取消时回滚优惠券（幂等）

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

| 错误码   | 含义          |
| ----- | ----------- |
| 200   | 成功          |
| 10001 | 参数错误        |
| 10002 | Token 过期    |
| 10003 | Token 无效    |
| 20001 | 订单不存在       |
| 20002 | 订单状态不允许此操作  |
| 20003 | 库存不足        |
| 20004 | 订单已支付       |
| 30001 | 支付单不存在      |
| 30002 | 支付单状态不允许此操作 |
| 40001 | 优惠券不存在      |
| 40002 | 优惠券已领完      |
| 40003 | 优惠券已使用      |
| 40004 | 优惠券已过期      |
| 50001 | 商品不存在       |
| 50002 | 商品已下架       |
| 60001 | 用户不存在       |
| 60002 | 用户已被封禁      |

错误响应示例：

```json
{
  "code": 10001,
  "message": "参数错误",
  "data": null
}
```

---

## 6. 管理后台模块

### 6.1 管理员登录

**请求路径**：`POST /api/v1/admin/login`

**请求体**：

```json
{
  "username": "admin",
  "password": "123456"
}
```

| 字段       | 类型     | 必填  | 说明     |
| -------- | ------ | --- | ------ |
| username | string | 是   | 管理员用户名 |
| password | string | 是   | 管理员密码  |

**响应体**：

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "access_expire": "2026-04-10T15:04:05+08:00"
  }
}
```

| 字段            | 类型     | 说明       |
| ------------- | ------ | -------- |
| access_token  | string | JWT 访问令牌 |
| access_expire | string | 令牌过期时间   |

---

### 6.2 获取管理员信息（需认证）

**请求路径**：`GET /api/v1/admin/info`

**请求头**：`Authorization: Bearer <access_token>`

**响应体**：

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": 1,
    "username": "admin",
    "nickname": "管理员",
    "avatar": "https://example.com/avatar.png",
    "role": "超级管理员"
  }
}
```

---

### 6.3 修改密码（需认证）

**请求路径**：`PUT /api/v1/admin/password`

**请求头**：`Authorization: Bearer <access_token>`

**请求体**：

```json
{
  "old_password": "123456",
  "new_password": "654321"
}
```

| 字段           | 类型     | 必填  | 说明         |
| ------------ | ------ | --- | ---------- |
| old_password | string | 是   | 原密码        |
| new_password | string | 是   | 新密码（6-20位） |

**响应体**：

```json
{
  "code": 200,
  "message": "success",
  "data": null
}
```

---

### 6.4 用户管理 - 用户列表（需认证）

**请求路径**：`GET /api/v1/admin/user/list`

**请求头**：`Authorization: Bearer <access_token>`

**请求参数**（Query）：

| 参数      | 类型     | 必填  | 说明                    |
| ------- | ------ | --- | --------------------- |
| keyword | string | 否   | 搜索关键词（手机号/昵称）         |
| status  | int64  | 否   | 用户状态：1-正常，2-限制下单，3-封禁 |
| page    | int64  | 否   | 页码，默认1                |
| size    | int64  | 否   | 每页数量，默认10             |

**响应体**：

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "total": 100,
    "list": [
      {
        "id": 1,
        "phone": "13800138000",
        "nickname": "张三",
        "avatar": "https://example.com/avatar.png",
        "status": 1,
        "status_desc": "正常",
        "created_at": "2026-04-01T10:00:00+08:00"
      }
    ]
  }
}
```

---

### 6.5 用户管理 - 用户详情（需认证）

**请求路径**：`GET /api/v1/admin/user/detail/:id`

**请求头**：`Authorization: Bearer <access_token>`

**路径参数**：

| 参数  | 类型     | 说明   |
| --- | ------ | ---- |
| id  | uint64 | 用户ID |

**响应体**：

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": 1,
    "phone": "13800138000",
    "nickname": "张三",
    "avatar": "https://example.com/avatar.png",
    "gender": 1,
    "birthday": "1990-01-01",
    "status": 1,
    "status_desc": "正常",
    "order_count": 10,
    "total_amount": 99900,
    "last_login_time": "2026-04-09T15:00:00+08:00",
    "created_at": "2026-04-01T10:00:00+08:00"
  }
}
```

---

### 6.6 用户管理 - 拉黑用户（需认证）

**请求路径**：`POST /api/v1/admin/user/blacklist`

**请求头**：`Authorization: Bearer <access_token>`

**请求体**：

```json
{
  "id": 1,
  "reason": "恶意下单"
}
```

| 字段     | 类型     | 必填  | 说明   |
| ------ | ------ | --- | ---- |
| id     | uint64 | 是   | 用户ID |
| reason | string | 否   | 拉黑原因 |

**响应体**：

```json
{
  "code": 200,
  "message": "success",
  "data": null
}
```

---

### 6.7 用户管理 - 恢复用户（需认证）

**请求路径**：`POST /api/v1/admin/user/recover`

**请求头**：`Authorization: Bearer <access_token>`

**请求体**：

```json
{
  "id": 1
}
```

| 字段  | 类型     | 必填  | 说明   |
| --- | ------ | --- | ---- |
| id  | uint64 | 是   | 用户ID |

**响应体**：

```json
{
  "code": 200,
  "message": "success",
  "data": null
}
```

---

### 6.8 分类管理 - 分类列表（需认证）

**请求路径**：`GET /api/v1/admin/category/list`

**请求头**：`Authorization: Bearer <access_token>`

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
        "name": "电子产品",
        "icon": "https://example.com/icon.png",
        "level": 1,
        "sort": 1
      }
    ]
  }
}
```

---

### 6.9 分类管理 - 保存分类（需认证）

**请求路径**：`POST /api/v1/admin/category/save`

**请求头**：`Authorization: Bearer <access_token>`

**请求体**：

```json
{
  "id": 0,
  "parent_id": 0,
  "name": "电子产品",
  "icon": "https://example.com/icon.png",
  "sort": 1
}
```

| 字段        | 类型     | 必填  | 说明             |
| --------- | ------ | --- | -------------- |
| id        | uint64 | 否   | 分类ID（新增时为0或不传） |
| parent_id | uint64 | 是   | 父分类ID（顶级为0）    |
| name      | string | 是   | 分类名称           |
| icon      | string | 否   | 分类图标           |
| sort      | int64  | 是   | 排序值            |

**响应体**：

```json
{
  "code": 200,
  "message": "success",
  "data": null
}
```

---

### 6.10 分类管理 - 删除分类（需认证）

**请求路径**：`DELETE /api/v1/admin/category/:id`

**请求头**：`Authorization: Bearer <access_token>`

**路径参数**：

| 参数  | 类型     | 说明   |
| --- | ------ | ---- |
| id  | uint64 | 分类ID |

**响应体**：

```json
{
  "code": 200,
  "message": "success",
  "data": null
}
```

---

### 6.11 商品管理 - 商品列表（需认证）

**请求路径**：`GET /api/v1/admin/product/list`

**请求头**：`Authorization: Bearer <access_token>`

**请求参数**（Query）：

| 参数          | 类型     | 必填  | 说明           |
| ----------- | ------ | --- | ------------ |
| keyword     | string | 否   | 搜索关键词（商品名称）  |
| category_id | uint64 | 否   | 分类ID         |
| status      | int64  | 否   | 状态：1-启用，2-禁用 |
| page        | int64  | 否   | 页码，默认1       |
| size        | int64  | 否   | 每页数量，默认10    |

**响应体**：

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "total": 100,
    "list": [
      {
        "id": 1,
        "name": "iPhone 15 Pro",
        "brand": "Apple",
        "default_pic": "https://example.com/product.png",
        "price": 799900,
        "sales": 1000,
        "status": 1,
        "created_at": "2026-04-01T10:00:00+08:00"
      }
    ]
  }
}
```

---

### 6.12 商品管理 - 商品详情（需认证）

**请求路径**：`GET /api/v1/admin/product/detail/:id`

**请求头**：`Authorization: Bearer <access_token>`

**路径参数**：

| 参数  | 类型     | 说明       |
| --- | ------ | -------- |
| id  | uint64 | 商品SPU ID |

**响应体**：

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": 1,
    "category_id": 1,
    "brand": "Apple",
    "name": "iPhone 15 Pro",
    "description": "苹果手机",
    "content": "<p>商品详情...</p>",
    "default_pic": "https://example.com/product.png",
    "banner_pics": ["https://example.com/banner1.png"],
    "video_url": "https://example.com/video.mp4",
    "weight": 200,
    "freight_template_id": 1,
    "status": 1,
    "skus": [
      {
        "id": 1,
        "code": "SKU001",
        "pic": ["https://example.com/sku.png"],
        "price": 799900,
        "stock": 100,
        "weight": 200,
        "spec_data": [
          {"name": "颜色", "value": "黑色"},
          {"name": "存储", "value": "256GB"}
        ]
      }
    ],
    "created_at": "2026-04-01T10:00:00+08:00"
  }
}
```

---

### 6.13 商品管理 - 保存商品（需认证）

**请求路径**：`POST /api/v1/admin/product/save`

**请求头**：`Authorization: Bearer <access_token>`

**请求体**：

```json
{
  "id": 0,
  "category_id": 1,
  "brand": "Apple",
  "name": "iPhone 15 Pro",
  "description": "苹果手机",
  "content": "<p>商品详情...</p>",
  "default_pic": "https://example.com/product.png",
  "banner_pics": ["https://example.com/banner1.png"],
  "video_url": "https://example.com/video.mp4",
  "weight": 200,
  "freight_template_id": 1,
  "status": 1,
  "skus": [
    {
      "id": 0,
      "code": "SKU001",
      "pic": ["https://example.com/sku.png"],
      "price": 799900,
      "stock": 100,
      "weight": 200,
      "spec_data": [
        {"name": "颜色", "value": "黑色"},
        {"name": "存储", "value": "256GB"}
      ]
    }
  ]
}
```

| 字段                  | 类型       | 必填  | 说明           |
| ------------------- | -------- | --- | ------------ |
| id                  | uint64   | 否   | 商品ID（新增时为0）  |
| category_id         | uint64   | 是   | 分类ID         |
| brand               | string   | 是   | 品牌名称         |
| name                | string   | 是   | 商品名称         |
| description         | string   | 是   | 商品描述         |
| content             | string   | 是   | 商品详情HTML     |
| default_pic         | string   | 是   | 默认图片         |
| banner_pics         | []string | 否   | 轮播图列表        |
| video_url           | string   | 否   | 视频链接         |
| weight              | int64    | 是   | 重量（克）        |
| freight_template_id | uint64   | 是   | 运费模板ID       |
| status              | int64    | 是   | 状态：1-启用，2-禁用 |
| skus                | []object | 是   | SKU列表        |

**响应体**：

```json
{
  "code": 200,
  "message": "success",
  "data": null
}
```

---

### 6.14 商品管理 - 更新商品状态（需认证）

**请求路径**：`POST /api/v1/admin/product/status`

**请求头**：`Authorization: Bearer <access_token>`

**请求体**：

```json
{
  "id": 1,
  "status": 2
}
```

| 字段     | 类型     | 必填  | 说明           |
| ------ | ------ | --- | ------------ |
| id     | uint64 | 是   | 商品ID         |
| status | int64  | 是   | 状态：1-启用，2-禁用 |

**响应体**：

```json
{
  "code": 200,
  "message": "success",
  "data": null
}
```

---

### 6.15 商品管理 - 删除商品（需认证）

**请求路径**：`DELETE /api/v1/admin/product/:id`

**请求头**：`Authorization: Bearer <access_token>`

**路径参数**：

| 参数  | 类型     | 说明       |
| --- | ------ | -------- |
| id  | uint64 | 商品SPU ID |

**响应体**：

```json
{
  "code": 200,
  "message": "success",
  "data": null
}
```

---

### 6.16 订单管理 - 订单列表（需认证）

**请求路径**：`GET /api/v1/admin/order/list`

**请求头**：`Authorization: Bearer <access_token>`

**请求参数**（Query）：

| 参数         | 类型     | 必填  | 说明                                             |
| ---------- | ------ | --- | ---------------------------------------------- |
| status     | int64  | 否   | 订单状态：10-待支付，20-已支付，30-已发货，40-已完成，50-已取消，60-售后中 |
| order_sn   | string | 否   | 订单号                                            |
| start_time | string | 否   | 开始时间（格式：2026-04-01 00:00:00）                   |
| end_time   | string | 否   | 结束时间（格式：2026-04-30 23:59:59）                   |
| page       | int64  | 否   | 页码，默认1                                         |
| size       | int64  | 否   | 每页数量，默认10                                      |

**响应体**：

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "total": 100,
    "list": [
      {
        "order_sn": "202604091234567890",
        "user_id": 1,
        "user_phone": "13800138000",
        "pay_amount": 799900,
        "status": 20,
        "status_desc": "已支付",
        "created_at": "2026-04-09T15:00:00+08:00"
      }
    ]
  }
}
```

---

### 6.17 订单管理 - 订单详情（需认证）

**请求路径**：`GET /api/v1/admin/order/detail/:order_sn`

**请求头**：`Authorization: Bearer <access_token>`

**路径参数**：

| 参数       | 类型     | 说明  |
| -------- | ------ | --- |
| order_sn | string | 订单号 |

**响应体**：

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "order_sn": "202604091234567890",
    "user_id": 1,
    "user_phone": "13800138000",
    "status": 20,
    "status_desc": "已支付",
    "pay_type": 1,
    "pay_type_desc": "微信支付",
    "pay_amount": 799900,
    "total_amount": 799900,
    "freight_amount": 0,
    "coupon_amount": 0,
    "remark": "请尽快发货",
    "address": {
      "receiver_name": "张三",
      "receiver_phone": "13800138000",
      "province": "广东省",
      "city": "深圳市",
      "district": "南山区",
      "detail_address": "科技园路88号"
    },
    "items": [
      {
        "sku_id": 1,
        "spu_id": 1,
        "product_name": "iPhone 15 Pro",
        "sku_name": "黑色 256GB",
        "pic": "https://example.com/sku.png",
        "price": 799900,
        "quantity": 1,
        "total_amount": 799900
      }
    ],
    "delivery_sn": "",
    "delivery_company": "",
    "payment_sn": "PAY202604091234567890",
    "create_time": "2026-04-09T15:00:00+08:00",
    "pay_time": "2026-04-09T15:05:00+08:00",
    "delivery_time": "",
    "finish_time": "",
    "cancel_time": ""
  }
}
```

---

### 6.18 订单管理 - 发货（需认证）

**请求路径**：`POST /api/v1/admin/order/ship`

**请求头**：`Authorization: Bearer <access_token>`

**请求体**：

```json
{
  "order_sn": "202604091234567890",
  "delivery_company": "顺丰速运",
  "delivery_sn": "SF1234567890",
  "remark": "已安排发货"
}
```

| 字段               | 类型     | 必填  | 说明     |
| ---------------- | ------ | --- | ------ |
| order_sn         | string | 是   | 订单号    |
| delivery_company | string | 是   | 物流公司名称 |
| delivery_sn      | string | 是   | 物流单号   |
| remark           | string | 否   | 备注     |

**响应体**：

```json
{
  "code": 200,
  "message": "success",
  "data": null
}
```

---

### 6.19 售后管理 - 退款列表（需认证）

**请求路径**：`GET /api/v1/admin/refund/list`

**请求头**：`Authorization: Bearer <access_token>`

**请求参数**（Query）：

| 参数     | 类型    | 必填  | 说明                                      |
| ------ | ----- | --- | --------------------------------------- |
| status | int64 | 否   | 状态：10-待审核，20-待商家收货，30-退款中，40-已完成，50-已拒绝 |
| page   | int64 | 否   | 页码，默认1                                  |
| size   | int64 | 否   | 每页数量，默认10                               |

**响应体**：

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "total": 10,
    "list": [
      {
        "id": 1,
        "refund_sn": "RF202604091234567890",
        "order_sn": "202604091234567890",
        "user_id": 1,
        "refund_amount": 799900,
        "status": 10,
        "status_desc": "待审核",
        "created_at": "2026-04-09T16:00:00+08:00"
      }
    ]
  }
}
```

---

### 6.20 售后管理 - 处理退款（需认证）

**请求路径**：`POST /api/v1/admin/refund/handle`

**请求头**：`Authorization: Bearer <access_token>`

**请求体**：

```json
{
  "refund_sn": "RF202604091234567890",
  "action": 1,
  "remark": "同意退款"
}
```

| 字段        | 类型     | 必填  | 说明           |
| --------- | ------ | --- | ------------ |
| refund_sn | string | 是   | 退款单号         |
| action    | int64  | 是   | 操作：1-通过，2-拒绝 |
| remark    | string | 否   | 处理备注         |

**响应体**：

```json
{
  "code": 200,
  "message": "success",
  "data": null
}
```

---

### 6.21 优惠券管理 - 优惠券列表（需认证）

**请求路径**：`GET /api/v1/admin/coupon/list`

**请求头**：`Authorization: Bearer <access_token>`

**请求参数**（Query）：

| 参数     | 类型     | 必填  | 说明           |
| ------ | ------ | --- | ------------ |
| status | int64  | 否   | 状态：1-启用，2-禁用 |
| name   | string | 否   | 优惠券名称        |
| page   | int64  | 否   | 页码，默认1       |
| size   | int64  | 否   | 每页数量，默认10    |

**响应体**：

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "total": 10,
    "list": [
      {
        "id": 1,
        "name": "新用户专享券",
        "type": 1,
        "total_quantity": 1000,
        "used_quantity": 500,
        "status": 1,
        "start_time": "2026-04-01T00:00:00+08:00",
        "end_time": "2026-05-01T23:59:59+08:00"
      }
    ]
  }
}
```

---

### 6.22 优惠券管理 - 保存优惠券（需认证）

**请求路径**：`POST /api/v1/admin/coupon/save`

**请求头**：`Authorization: Bearer <access_token>`

**请求体**：

```json
{
  "id": 0,
  "name": "新用户专享券",
  "type": 1,
  "threshold_amount": 10000,
  "reduce_amount": 1000,
  "discount_rate": 0,
  "max_discount_amount": 0,
  "total_quantity": 1000,
  "per_user_limit": 1,
  "start_time": "2026-04-01T00:00:00+08:00",
  "end_time": "2026-05-01T23:59:59+08:00",
  "status": 1,
  "description": "新用户专享"
}
```

| 字段                  | 类型     | 必填  | 说明                    |
| ------------------- | ------ | --- | --------------------- |
| id                  | uint64 | 否   | 优惠券ID（新增时为0）          |
| name                | string | 是   | 优惠券名称                 |
| type                | int64  | 是   | 类型：1-满减券，2-折扣券，3-无门槛券 |
| threshold_amount    | int64  | 是   | 使用门槛金额（分），0表示无门槛      |
| reduce_amount       | int64  | 否   | 满减金额（分），仅满减券有效        |
| discount_rate       | int64  | 否   | 折扣率（万比分比），仅折扣券有效      |
| max_discount_amount | int64  | 否   | 折扣封顶金额（分），仅折扣券有效      |
| total_quantity      | int64  | 是   | 发行总量                  |
| per_user_limit      | int64  | 是   | 每人限领数量                |
| start_time          | string | 是   | 生效开始时间                |
| end_time            | string | 是   | 失效结束时间                |
| status              | int64  | 是   | 状态：1-启用，2-禁用          |
| description         | string | 否   | 使用说明                  |

**响应体**：

```json
{
  "code": 200,
  "message": "success",
  "data": null
}
```

---

### 6.23 优惠券管理 - 更新优惠券状态（需认证）

**请求路径**：`POST /api/v1/admin/coupon/status`

**请求头**：`Authorization: Bearer <access_token>`

**请求体**：

```json
{
  "id": 1,
  "status": 2
}
```

| 字段     | 类型     | 必填  | 说明           |
| ------ | ------ | --- | ------------ |
| id     | uint64 | 是   | 优惠券ID        |
| status | int64  | 是   | 状态：1-启用，2-禁用 |

**响应体**：

```json
{
  "code": 200,
  "message": "success",
  "data": null
}
```

---

### 6.24 操作日志 - 日志列表（需认证）

**请求路径**：`GET /api/v1/admin/operate/logs`

**请求头**：`Authorization: Bearer <access_token>`

**请求参数**（Query）：

| 参数   | 类型    | 必填  | 说明        |
| ---- | ----- | --- | --------- |
| page | int64 | 否   | 页码，默认1    |
| size | int64 | 否   | 每页数量，默认10 |

**响应体**：

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "total": 100,
    "list": [
      {
        "id": 1,
        "admin_id": 1,
        "module": "订单",
        "action": "发货",
        "content": "订单号：202604091234567890",
        "ip": "192.168.1.100",
        "created_at": "2026-04-09T16:00:00+08:00"
      }
    ]
  }
}
```

---
