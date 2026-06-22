# PrimeMall 商城 API 文档

## 概述

本文档定义了 PrimeMall 商城**前台用户端 (Shop)** 和**管理后台 (Admin)** 的所有 API 接口。系统采用 **go-zero 微服务架构**，统一网关负责路由、鉴权、限流。

- **字符编码**：UTF-8
- **请求/响应格式**：JSON
- **时间格式**：RFC3339，例如 `2026-04-09T15:04:05+08:00`
- **金额单位**：**分**（int64），所有金额字段均为整数分
- **密码规则**：8-20 位，至少包含字母和数字

### 统一响应格式

所有接口响应均遵循以下结构：

```json
{
  "code": 0,
  "msg": "success",
  "data": { ... }
}
```

| 字段   | 类型                | 说明                       |
| ---- | ----------------- | ------------------------ |
| code | int               | 业务状态码，`0` 表示成功，其他值表示错误   |
| msg  | string            | 提示信息                     |
| data | object/array/null | 响应数据，成功时返回具体内容，失败时为 null |

**注意**：管理员登录 (`POST /api/v1/admin/login`) 例外，其响应直接返回 `AdminLoginResp` 对象，未包裹在标准格式中。

错误响应格式相同，`code` 为非 0 值，`msg` 包含错误描述。

---

### 认证说明

系统使用**双令牌**机制：

- **Access Token**：有效期 2 小时，放在 `Authorization: Bearer <access_token>` 头中访问业务接口
- **Refresh Token**：有效期 7 天，仅用于调用刷新接口获取新的 Access Token

当 Access Token 过期时，API 返回 `{"code":12003,"msg":"Token 已过期","data":null}`。前端应使用 Refresh Token 调用刷新接口，获得新令牌后重试原请求。

### 公共中间件

- **JWT 认证**：需认证的接口统一使用 `rest.WithJwt` 中间件，要求请求头携带 `Authorization: Bearer <access_token>`
- **Casbin 权限**：需认证的接口额外经过 Casbin 权限中间件，根据用户角色进行细粒度权限控制
- **Token 黑名单**：退出登录或被拉黑的用户 Token 将被加入黑名单，后续请求被拦截

---

## 模块一：用户服务（前台）

### 1.1 获取短信验证码

- **幂等**：是（通过 `idempotency_key`）
- **请求**：`POST /api/v1/user/captcha`

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
| phone           | string | 是   | 手机号（11 位）                                                         |
| scene           | string | 是   | 场景：`login`, `register`, `reset_pwd`, `update_pwd`, `update_phone` |
| idempotency_key | string | 是   | 幂等键（客户端 UUID，防止重复发送）                                              |

**响应体**：`{"code":0,"msg":"success","data":null}`

**注意事项**：

- 同一手机号 1 分钟内最多 1 次，同一 IP 每小时最多 10 次
- 验证码存储于 Redis：`sms:code:{scene}:{phone}`，有效期 5 分钟

---

### 1.2 手机号密码登录

- **请求**：`POST /api/v1/user/login`

**请求体**：

```json
{
  "phone": "13800138000",
  "password": "明文密码"
}
```

| 字段       | 类型     | 必填  | 说明         |
| -------- | ------ | --- | ---------- |
| phone    | string | 是   | 手机号（11 位）  |
| password | string | 是   | 密码（8-20 位） |

**响应体**：

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIs...",
    "access_expire": 1744189445,
    "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
    "refresh_expire": 1744707845
  }
}
```

| data 字段        | 类型     | 说明                            |
| -------------- | ------ | ----------------------------- |
| access_token   | string | 业务 API 访问令牌                   |
| access_expire  | int64  | Access Token 过期时间（Unix 时间戳秒）  |
| refresh_token  | string | 刷新令牌                          |
| refresh_expire | int64  | Refresh Token 过期时间（Unix 时间戳秒） |

---

### 1.3 手机号验证码登录

- **请求**：`POST /api/v1/user/login/mobile`

**请求体**：

```json
{
  "phone": "13800138000",
  "code": "123456"
}
```

| 字段    | 类型     | 必填  | 说明        |
| ----- | ------ | --- | --------- |
| phone | string | 是   | 手机号（11 位） |
| code  | string | 是   | 验证码（6 位）  |

**响应体**：同 1.2。

---

### 1.4 用户注册

- **幂等**：是（通过 `idempotency_key`）
- **请求**：`POST /api/v1/user/register`

**请求体**：

```json
{
  "phone": "13800138000",
  "password": "密码",
  "confirm_password": "密码",
  "code": "123456",
  "nickname": "昵称",
  "idempotency_key": "uuid-xxxx"
}
```

| 字段               | 类型     | 必填  | 说明                   |
| ---------------- | ------ | --- | -------------------- |
| phone            | string | 是   | 手机号（11 位）            |
| password         | string | 是   | 密码（8-20 位）           |
| confirm_password | string | 是   | 确认密码（需与 password 一致） |
| code             | string | 是   | 验证码（6 位）             |
| nickname         | string | 否   | 昵称                   |
| idempotency_key  | string | 是   | 幂等键                  |

**响应体**：同 1.2（注册后自动登录）。

---

### 1.5 刷新 Access Token

- **请求**：`POST /api/v1/user/token/refresh`

**请求头**：`Authorization: Bearer <refresh_token>`

**请求体**：

```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
}
```

| 字段            | 类型     | 必填  | 说明   |
| ------------- | ------ | --- | ---- |
| refresh_token | string | 是   | 刷新令牌 |

**响应体**：同 1.2（返回新的一对令牌）。

---

### 1.6 忘记密码重置

- **幂等**：是（通过 `idempotency_key`）
- **请求**：`POST /api/v1/user/password/reset`

**请求体**：

```json
{
  "phone": "13800138000",
  "code": "123456",
  "new_password": "新密码",
  "idempotency_key": "uuid-xxxx"
}
```

| 字段              | 类型     | 必填  | 说明          |
| --------------- | ------ | --- | ----------- |
| phone           | string | 是   | 手机号（11 位）   |
| code            | string | 是   | 验证码（6 位）    |
| new_password    | string | 是   | 新密码（8-20 位） |
| idempotency_key | string | 是   | 幂等键         |

**响应体**：`{"code":0,"msg":"success","data":null}`

---

### 1.7 OSS 上传回调

- **请求**：`POST /api/v1/user/upload/avatar/callback`
- **说明**：阿里云 OSS 上传成功后的服务端回调接口（服务端自行保证幂等）
- **请求体**：

```json
{
  "bucket": "primemall",
  "object": "avatar/xxx.jpg",
  "etag": "etag-value",
  "size": 102400,
  "mimeType": "image/jpeg"
}
```

| 字段       | 类型     | 必填  | 说明             |
| -------- | ------ | --- | -------------- |
| bucket   | string | 是   | 阿里云 Bucket     |
| object   | string | 是   | 阿里云 OSS Object |
| etag     | string | 是   | 阿里云 OSS Etag   |
| size     | int64  | 是   | 文件大小（字节）       |
| mimeType | string | 是   | 文件 MIME 类型     |

**响应体**：`{"code":0,"msg":"success","data":null}`

---

### 1.8 获取个人资料

- **认证**：需要 JWT
- **请求**：`GET /api/v1/user/info`
- **请求头**：`Authorization: Bearer <access_token>`

**响应体**：

```json
{
  "code": 0,
  "msg": "success",
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
| id          | uint64 | 用户 ID                 |
| phone       | string | 手机号                   |
| nickname    | string | 昵称                    |
| avatar      | string | 头像 URL                |
| gender      | int64  | 性别：0-未知，1-男，2-女       |
| birthday    | string | 生日（日期格式 `2006-01-02`） |
| status      | int64  | 账号状态：1-正常，2-限制下单，3-封禁 |
| status_desc | string | 状态描述                  |

---

### 1.9 更新个人资料

- **认证**：需要 JWT
- **请求**：`PUT /api/v1/user/info`
- **请求头**：`Authorization: Bearer <access_token>`

**请求体**（全部可选，仅传需要修改的字段）：

```json
{
  "nickname": "新昵称",
  "avatar": "https://...",
  "gender": 2,
  "birthday": "1995-05-05"
}
```

| 字段       | 类型     | 必填  | 说明                 |
| -------- | ------ | --- | ------------------ |
| nickname | string | 否   | 昵称（1-30 字符）        |
| avatar   | string | 否   | 头像 URL（1-255 字符）   |
| gender   | int64  | 否   | 性别：0-未知，1-男，2-女    |
| birthday | string | 否   | 生日，格式 `2006-01-02` |

**响应体**：`{"code":0,"msg":"success","data":null}`

---

### 1.10 修改登录密码

- **认证**：需要 JWT
- **请求**：`PUT /api/v1/user/password/update`
- **请求头**：`Authorization: Bearer <access_token>`

**请求体**：

```json
{
  "old_password": "旧密码",
  "new_password": "新密码",
  "confirm_password": "新密码"
}
```

| 字段               | 类型     | 必填  | 说明                       |
| ---------------- | ------ | --- | ------------------------ |
| old_password     | string | 是   | 旧密码（8-20 位）              |
| new_password     | string | 是   | 新密码（8-20 位，不能与旧密码一致）     |
| confirm_password | string | 是   | 确认密码（需与 new_password 一致） |

**响应体**：`{"code":0,"msg":"success","data":null}`

---

### 1.11 修改手机号

- **认证**：需要 JWT
- **幂等**：是（通过 `idempotency_key`）
- **请求**：`PUT /api/v1/user/phone`
- **请求头**：`Authorization: Bearer <access_token>`

**请求体**：

```json
{
  "old_phone_code": "123456",
  "new_phone": "13900139000",
  "new_phone_code": "654321",
  "idempotency_key": "uuid-xxxx"
}
```

| 字段              | 类型     | 必填  | 说明             |
| --------------- | ------ | --- | -------------- |
| old_phone_code  | string | 否   | 原手机验证码（已登录时可选） |
| new_phone       | string | 是   | 新手机号（11 位）     |
| new_phone_code  | string | 是   | 新手机验证码（6 位）    |
| idempotency_key | string | 是   | 幂等键            |

**响应体**：`{"code":0,"msg":"success","data":null}`

---

### 1.12 注销账号

- **认证**：需要 JWT
- **请求**：`DELETE /api/v1/user/account`
- **请求头**：`Authorization: Bearer <access_token>`

**请求体**：

```json
{
  "password": "密码",
  "refresh_token": "refresh_token内容"
}
```

| 字段            | 类型     | 必填  | 说明         |
| ------------- | ------ | --- | ---------- |
| password      | string | 是   | 密码（8-20 位） |
| refresh_token | string | 是   | 刷新令牌       |

**响应体**：`{"code":0,"msg":"success","data":null}`

---

### 1.13 退出登录

- **认证**：需要 JWT
- **请求**：`POST /api/v1/user/logout`
- **请求头**：`Authorization: Bearer <access_token>`

**响应体**：`{"code":0,"msg":"success","data":null}`

**注意事项**：

- 将 Token 加入黑名单，Access Token 立即失效

---

### 1.14 获取 OSS 上传凭证

- **认证**：需要 JWT
- **请求**：`POST /api/v1/user/upload/avatar/token`
- **请求头**：`Authorization: Bearer <access_token>`

**请求体**：

```json
{
  "file_name": "avatar.jpg"
}
```

| 字段        | 类型     | 必填  | 说明             |
| --------- | ------ | --- | -------------- |
| file_name | string | 是   | 原始文件名（用于提取扩展名） |

**响应体**：

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "accessid": "LTAI5t...",
    "host": "https://primemall.oss-cn-shenzhen.aliyuncs.com",
    "policy": "eyJleHBpcmF0aW9uIjoi...",
    "signature": "eRMV...",
    "expire": 1744189445,
    "dir": "avatar/",
    "object_key": "avatar/xxx.jpg",
    "callback": "..."
  }
}
```

| data 字段    | 类型     | 说明                   |
| ---------- | ------ | -------------------- |
| accessid   | string | 阿里云 AccessKeyId      |
| host       | string | 阿里云 OSS Host         |
| policy     | string | 阿里云 OSS 策略           |
| signature  | string | 阿里云 OSS 签名           |
| expire     | int64  | OSS 签名有效期（Unix 时间戳秒） |
| dir        | string | OSS 存储目录（`avatar/`）  |
| object_key | string | OSS 存储对象名            |
| callback   | string | OSS 回调参数             |

---

## 模块二：地址管理

### 2.1 添加收货地址

- **认证**：需要 JWT
- **幂等**：是
- **请求**：`POST /api/v1/address/`
- **请求头**：`Authorization: Bearer <access_token>`

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

| 字段                    | 类型     | 必填  | 说明                                      |
| --------------------- | ------ | --- | --------------------------------------- |
| receiver_name         | string | 是   | 收货人姓名（1-50 字符）                          |
| receiver_phone        | string | 是   | 收货人电话（11 位）                             |
| is_default            | int64  | 是   | 是否默认地址：0-否，1-是                          |
| tag                   | string | 否   | 地址标签：`HOME`-家, `OFFICE`-公司, `SCHOOL`-学校 |
| detail.province       | string | 是   | 省份（1-50 字符）                             |
| detail.city           | string | 是   | 城市（1-50 字符）                             |
| detail.district       | string | 是   | 区/县（1-50 字符）                            |
| detail.detail_address | string | 是   | 详细地址（1-200 字符）                          |
| detail.postal_code    | string | 否   | 邮政编码（4-10 字符）                           |

**响应体**：`{"code":0,"msg":"success","data":null}`

**注意事项**：

- 每个用户最多添加 20 条地址
- 将已存在的默认地址取消，再将新地址设为默认

---

### 2.2 获取地址列表

- **认证**：需要 JWT
- **请求**：`GET /api/v1/address/list`
- **请求头**：`Authorization: Bearer <access_token>`
- **请求参数**（Query）：

| 参数   | 类型    | 必填  | 说明                |
| ---- | ----- | --- | ----------------- |
| page | int64 | 否   | 页码，默认 1           |
| size | int64 | 否   | 每页数量，默认 10，最大 100 |

**响应体**：

```json
{
  "code": 0,
  "msg": "success",
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

### 2.3 更新收货地址

- **认证**：需要 JWT
- **请求**：`PUT /api/v1/address/`
- **请求头**：`Authorization: Bearer <access_token>`

**请求体**：

```json
{
  "id": 101,
  "receiver_name": "李四",
  "receiver_phone": "13900139000",
  "is_default": 0,
  "tag": "OFFICE",
  "detail": {
    "province": "广东省",
    "city": "深圳市",
    "district": "南山区",
    "detail_address": "科技园 xxx 大厦",
    "postal_code": "518000"
  }
}
```

**响应体**：`{"code":0,"msg":"success","data":null}`

---

### 2.4 删除收货地址

- **认证**：需要 JWT
- **请求**：`DELETE /api/v1/address/:id`
- **请求头**：`Authorization: Bearer <access_token>`

**路径参数**：

| 参数  | 类型     | 说明    |
| --- | ------ | ----- |
| id  | uint64 | 地址 ID |

**响应体**：`{"code":0,"msg":"success","data":null}`

**注意事项**：

- 默认地址不可直接删除，需先设置其他地址为默认

---

### 2.5 设置默认收货地址

- **认证**：需要 JWT
- **请求**：`PATCH /api/v1/address/default/:id`
- **请求头**：`Authorization: Bearer <access_token>`

**路径参数**：

| 参数  | 类型     | 说明    |
| --- | ------ | ----- |
| id  | uint64 | 地址 ID |

**响应体**：`{"code":0,"msg":"success","data":null}`

---

## 模块三：商品服务（前台公开）

### 3.1 获取全部分类树

- **请求**：`GET /api/v1/product/category/tree`

**响应体**：

```json
{
  "code": 0,
  "msg": "success",
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

| data 字段   | 类型       | 说明                  |
| --------- | -------- | ------------------- |
| id        | uint64   | 分类 ID               |
| parent_id | uint64   | 父分类 ID（顶级为 0）       |
| name      | string   | 分类名称                |
| icon      | string   | 分类图标 URL            |
| level     | int64    | 分类层级：1-一级，2-二级，3-三级 |
| children  | []object | 子分类列表（递归结构）         |

---

### 3.2 获取单个分类详情

- **请求**：`GET /api/v1/product/category/:id`

**路径参数**：

| 参数  | 类型     | 说明    |
| --- | ------ | ----- |
| id  | uint64 | 分类 ID |

**响应体**：

```json
{
  "code": 0,
  "msg": "success",
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

### 3.3 商品列表与筛选

- **请求**：`GET /api/v1/product/list`
- **请求参数**（Query）：

| 参数          | 类型       | 必填  | 说明                                       |
| ----------- | -------- | --- | ---------------------------------------- |
| category_id | uint64   | 否   | 分类 ID                                    |
| brand       | string   | 否   | 品牌模糊匹配                                   |
| min_price   | int64    | 否   | 最低价格（分）                                  |
| max_price   | int64    | 否   | 最高价格（分）                                  |
| keyword     | string   | 否   | 商品名称关键词搜索                                |
| attrs       | []object | 否   | 属性过滤，如 `[{"name":"颜色","values":["黑色"]}]` |
| is_new      | bool     | 否   | 是否新品（创建时间 < 7 天）                         |
| sort_by     | string   | 否   | 排序字段：`price`, `sales`, `created_at`      |
| sort_type   | string   | 否   | 排序方向：`asc`, `desc`                       |
| page        | int64    | 否   | 默认 1                                     |
| size        | int64    | 否   | 默认 10，最大 100                             |

**响应体**：

```json
{
  "code": 0,
  "msg": "success",
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

| data 列表字段   | 类型     | 说明                  |
| ----------- | ------ | ------------------- |
| id          | uint64 | 商品 SPU ID           |
| name        | string | 商品名称                |
| brand       | string | 品牌                  |
| description | string | 商品简介                |
| default_pic | string | 默认主图 URL            |
| price       | int64  | 最低 SKU 价格（分）        |
| sales       | int64  | 实际销量                |
| show_sales  | int64  | 前台展示销量（=sales+虚拟销量） |

---

### 3.4 商品详情

- **请求**：`GET /api/v1/product/detail/:id`

**路径参数**：

| 参数  | 类型     | 说明        |
| --- | ------ | --------- |
| id  | uint64 | 商品 SPU ID |

**响应体**：

```json
{
  "code": 0,
  "msg": "success",
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
        "code": "SKU001",
        "pic": ["https://..."],
        "price": 799900,
        "stock": 50,
        "weight": 221,
        "spec_data": [{"name": "颜色", "values": "黑色"}, {"name": "容量", "values": "256G"}]
      }
    ],
    "attributes": [
      {"name": "颜色", "values": ["黑色", "白色"]},
      {"name": "容量", "values": ["128G", "256G"]}
    ]
  }
}
```

| data 字段          | 类型       | 说明             |
| ---------------- | -------- | -------------- |
| id               | uint64   | 商品 SPU ID      |
| status           | int64    | 1-上架，2-下架      |
| status_desc      | string   | 状态描述           |
| price            | int64    | 最低 SKU 价格（分）   |
| sales            | int64    | 实际销量           |
| show_sales       | int64    | 前台展示销量         |
| skus[].spec_data | []object | SKU 规格数据       |
| attributes       | []object | 聚合属性（用于前端规格筛选） |

---

### 3.5 个性化推荐

- **请求**：`GET /api/v1/product/recommend`
- **请求参数**（Query）：

| 参数         | 类型     | 必填  | 说明                                        |
| ---------- | ------ | --- | ----------------------------------------- |
| scene      | string | 否   | 场景：`home`（首页）, `cart`（购物车）, `detail`（详情页） |
| limit      | int64  | 否   | 返回数量，默认 10，最大 50                          |
| current_id | uint64 | 否   | 当前商品 ID（详情页推荐时使用）                         |

**响应体**：同 3.3 的 data 结构。

---

### 3.6 批量获取 SKU 价格与库存

- **请求**：`POST /api/v1/product/skus/batch`

**请求体**：

```json
{
  "sku_ids": [2001, 2002]
}
```

| 字段      | 类型       | 必填  | 说明                 |
| ------- | -------- | --- | ------------------ |
| sku_ids | []uint64 | 是   | SKU ID 列表（1-100 个） |

**响应体**：

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "list": [
      {"id": 2001, "price": 799900, "stock": 50},
      {"id": 2002, "price": 899900, "stock": 0}
    ]
  }
}
```

---

## 模块四：搜索服务

### 4.1 ES 商品搜索

- **请求**：`GET /api/v1/search/products`
- **底层**：基于 Elasticsearch 搜索引擎

**请求参数**（Query）：

| 参数          | 类型     | 必填  | 说明                    |
| ----------- | ------ | --- | --------------------- |
| keyword     | string | 否   | 搜索关键词                 |
| category_id | int64  | 否   | 分类 ID                 |
| brand       | string | 否   | 品牌                    |
| min_price   | int64  | 否   | 最低价格（分）               |
| max_price   | int64  | 否   | 最高价格（分）               |
| sort_by     | string | 否   | 排序字段：`price`, `sales` |
| sort_type   | string | 否   | 排序方向：`asc`, `desc`    |
| page        | int64  | 否   | 默认 1                  |
| size        | int64  | 否   | 默认 10                 |

**响应体**：

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "total": 200,
    "list": [
      {
        "id": 1001,
        "name": "iPhone 15 Pro",
        "brand": "Apple",
        "cover": "https://...",
        "price": 799900,
        "sales": 1234,
        "category_id": 1,
        "category_name": "手机",
        "description": "钛金属，A17 Pro 芯片"
      }
    ]
  }
}
```

---

### 4.2 ES 搜索建议

- **请求**：`GET /api/v1/search/suggest`

**请求参数**（Query）：

| 参数      | 类型     | 必填  | 说明              |
| ------- | ------ | --- | --------------- |
| keyword | string | 是   | 搜索关键词           |
| size    | int64  | 否   | 建议数量，默认 5，最大 10 |

**响应体**：

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "suggestions": ["iPhone 15", "iPhone 14", "iPhone 13"]
  }
}
```

---

### 4.3 热搜词列表

- **请求**：`GET /api/v1/search/hot-keywords`

**响应体**：

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "list": [
      {"id": 1, "keyword": "iPhone", "search_count": 9999, "sort": 1},
      {"id": 2, "keyword": "华为", "search_count": 8888, "sort": 2}
    ]
  }
}
```

| data 列表字段    | 类型     | 说明     |
| ------------ | ------ | ------ |
| id           | uint64 | 热搜词 ID |
| keyword      | string | 关键词    |
| search_count | int64  | 搜索次数   |
| sort         | int32  | 排序值    |

---

### 4.4 搜索历史

- **认证**：需要 JWT
- **请求**：`GET /api/v1/search/history`
- **请求头**：`Authorization: Bearer <access_token>`

**响应体**：

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "list": [
      {"keyword": "iPhone", "created_at": 1744189445},
      {"keyword": "耳机", "created_at": 1744189000}
    ]
  }
}
```

| data 列表字段  | 类型     | 说明             |
| ---------- | ------ | -------------- |
| keyword    | string | 搜索关键词          |
| created_at | int64  | 搜索时间（Unix 时间戳） |

---

### 4.5 清除搜索历史

- **认证**：需要 JWT
- **请求**：`DELETE /api/v1/search/history`
- **请求头**：`Authorization: Bearer <access_token>`

**响应体**：`{"code":0,"msg":"success","data":null}`

---

## 模块五：购物车

### 5.1 添加/更新购物车商品

- **认证**：需要 JWT
- **幂等**：是（通过 `idempotency_key`）
- **请求**：`POST /api/v1/order/cart/update`
- **请求头**：`Authorization: Bearer <access_token>`

**请求体**：

```json
{
  "sku_id": 2001,
  "quantity": 2,
  "idempotency_key": "uuid-xxxx"
}
```

| 字段              | 类型     | 必填  | 说明                       |
| --------------- | ------ | --- | ------------------------ |
| sku_id          | uint64 | 是   | SKU ID                   |
| quantity        | int64  | 是   | 数量（大于 0 为添加/更新，等于 0 为删除） |
| idempotency_key | string | 是   | 幂等键                      |

**响应体**：`{"code":0,"msg":"success","data":null}`

---

### 5.2 修改购物车勾选状态

- **认证**：需要 JWT
- **请求**：`POST /api/v1/order/cart/select`
- **请求头**：`Authorization: Bearer <access_token>`

**请求体**：

```json
{
  "sku_id": 2001,
  "selected": true
}
```

| 字段       | 类型     | 必填  | 说明     |
| -------- | ------ | --- | ------ |
| sku_id   | uint64 | 是   | SKU ID |
| selected | bool   | 是   | 是否勾选   |

**响应体**：`{"code":0,"msg":"success","data":null}`

---

### 5.3 获取购物车列表

- **认证**：需要 JWT
- **请求**：`GET /api/v1/order/cart/list`
- **请求头**：`Authorization: Bearer <access_token>`

**响应体**：

```json
{
  "code": 0,
  "msg": "success",
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

| data 字段        | 类型    | 说明                |
| -------------- | ----- | ----------------- |
| items[].price  | int64 | 实时单价（分）           |
| items[].stock  | int64 | 实时库存              |
| items[].status | int64 | 商品状态：1-正常，2-下架/失效 |
| total_amount   | int64 | 勾选商品总金额（分）        |

---

### 5.4 批量删除购物车商品

- **认证**：需要 JWT
- **请求**：`DELETE /api/v1/order/cart/remove`
- **请求头**：`Authorization: Bearer <access_token>`

**请求体**：

```json
{
  "sku_ids": [2001, 2002]
}
```

| 字段      | 类型       | 必填  | 说明                 |
| ------- | -------- | --- | ------------------ |
| sku_ids | []uint64 | 是   | SKU ID 列表（1-100 个） |

**响应体**：`{"code":0,"msg":"success","data":null}`

---

### 5.5 清空购物车

- **认证**：需要 JWT
- **请求**：`DELETE /api/v1/order/cart/clear`
- **请求头**：`Authorization: Bearer <access_token>`

**响应体**：`{"code":0,"msg":"success","data":null}`

---

## 模块六：订单核心流程

### 6.1 预下单（确认订单页）

- **认证**：需要 JWT
- **请求**：`POST /api/v1/order/pre-order`
- **请求头**：`Authorization: Bearer <access_token>`

**请求体**：

```json
{
  "items": [
    {"sku_id": 2001, "quantity": 1}
  ],
  "address_id": 101,
  "coupon_id": 0
}
```

| 字段               | 类型       | 必填  | 说明                |
| ---------------- | -------- | --- | ----------------- |
| items            | []object | 是   | 下单商品列表（1-50 个）    |
| items[].sku_id   | uint64   | 是   | SKU ID            |
| items[].quantity | int64    | 是   | 购买数量              |
| address_id       | uint64   | 是   | 收货地址 ID           |
| coupon_id        | uint64   | 是   | 优惠券 ID（传 0 表示不使用） |

**响应体**：

```json
{
  "code": 0,
  "msg": "success",
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
    "total_amount": 799900,
    "freight_amount": 0,
    "coupon_amount": 10000,
    "pay_amount": 789900
  }
}
```

| data 字段          | 类型     | 说明              |
| ---------------- | ------ | --------------- |
| settlement_token | string | 结算令牌（有效期 15 分钟） |
| total_amount     | int64  | 商品总价（分）         |
| freight_amount   | int64  | 运费（分）           |
| coupon_amount    | int64  | 优惠券减免（分）        |
| pay_amount       | int64  | 实际支付金额（分）       |

---

### 6.2 正式提交订单

- **认证**：需要 JWT
- **幂等**：是（通过 `idempotency_key`）
- **请求**：`POST /api/v1/order/create`
- **请求头**：`Authorization: Bearer <access_token>`

**请求体**：

```json
{
  "settlement_token": "abc123...",
  "pay_type": 1,
  "remark": "请放快递柜",
  "idempotency_key": "uuid-xxxx",
  "cart_sku_ids": [2001, 2002]
}
```

| 字段               | 类型       | 必填  | 说明                       |
| ---------------- | -------- | --- | ------------------------ |
| settlement_token | string   | 是   | 结算令牌（预下单返回，含地址与优惠券信息）    |
| pay_type         | int64    | 否   | 支付方式：1-微信，2-支付宝          |
| remark           | string   | 否   | 订单备注（最多 200 字符）          |
| idempotency_key  | string   | 是   | 幂等键（前端生成）                |
| cart_sku_ids     | []uint64 | 否   | 购物车商品 SKU ID 列表，直购时传空或不传 |

**注意**：收货地址和优惠券在预下单（6.1）时选定，结算令牌已包含这些信息，正式提交时无需重复传入。

**响应体**：

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "order_sn": "202604091234567890",
    "pay_amount": 789900
  }
}
```

---

### 6.3 订单列表

- **认证**：需要 JWT
- **请求**：`GET /api/v1/order/list`
- **请求头**：`Authorization: Bearer <access_token>`
- **请求参数**（Query）：

| 参数     | 类型    | 必填  | 说明                                                  |
| ------ | ----- | --- | --------------------------------------------------- |
| status | int64 | 否   | 订单状态：0-全部，10-待付款，20-待发货，30-待收货，40-已完成，50-已取消，60-售后中 |
| page   | int64 | 否   | 默认 1                                                |
| size   | int64 | 否   | 默认 10                                               |

**响应体**：

```json
{
  "code": 0,
  "msg": "success",
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
            "full_name": "iPhone 15 Pro 黑色 256G",
            "pic": "https://...",
            "price": 799900,
            "quantity": 1
          }
        ]
      }
    ]
  }
}
```

| data 列表字段         | 类型     | 说明        |
| ----------------- | ------ | --------- |
| items[].full_name | string | 商品名称（含规格） |
| items[].price     | int64  | 下单时单价（分）  |

---

### 6.4 订单详情

- **认证**：需要 JWT
- **请求**：`GET /api/v1/order/detail/:order_sn`
- **请求头**：`Authorization: Bearer <access_token>`

**路径参数**：

| 参数       | 类型     | 说明  |
| -------- | ------ | --- |
| order_sn | string | 订单号 |

**响应体**：

```json
{
  "code": 0,
  "msg": "success",
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
    "expire_time": 1744188345,
    "pay_time": "",
    "delivery_time": "",
    "finish_time": "",
    "cancel_time": ""
  }
}
```

| data 字段       | 类型     | 说明                                             |
| ------------- | ------ | ---------------------------------------------- |
| status        | int64  | 10-待付款, 20-待发货, 30-待收货, 40-已完成, 50-已取消, 60-售后中 |
| expire_time   | int64  | 待付款截止时间（Unix 时间戳秒）                             |
| delivery_sn   | string | 物流单号（未发货为空）                                    |
| delivery_corp | string | 物流公司（未发货为空）                                    |

---

### 6.5 取消订单

- **认证**：需要 JWT
- **请求**：`POST /api/v1/order/cancel/:order_sn`
- **请求头**：`Authorization: Bearer <access_token>`

**路径参数**：

| 参数       | 类型     | 说明  |
| -------- | ------ | --- |
| order_sn | string | 订单号 |

**请求体**（可选）：

```json
{
  "cancel_reason": "不想要了",
  "reason_type": 1
}
```

| 字段            | 类型     | 必填  | 说明              |
| ------------- | ------ | --- | --------------- |
| cancel_reason | string | 否   | 取消原因（最多 200 字符） |
| reason_type   | int64  | 否   | 取消原因类型          |

**响应体**：`{"code":0,"msg":"success","data":null}`

**注意事项**：

- **仅待支付（10）状态**可取消
- 取消后释放库存、解锁优惠券

---

### 6.6 确认收货

- **认证**：需要 JWT
- **请求**：`POST /api/v1/order/confirm/:order_sn`
- **请求头**：`Authorization: Bearer <access_token>`

**路径参数**：

| 参数       | 类型     | 说明  |
| -------- | ------ | --- |
| order_sn | string | 订单号 |

**响应体**：`{"code":0,"msg":"success","data":null}`

---

### 6.7 查询物流轨迹

- **认证**：需要 JWT
- **请求**：`GET /api/v1/order/delivery/track/:order_sn`
- **请求头**：`Authorization: Bearer <access_token>`

**路径参数**：

| 参数       | 类型     | 说明  |
| -------- | ------ | --- |
| order_sn | string | 订单号 |

**响应体**：

```json
{
  "code": 0,
  "msg": "success",
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

| data 字段       | 类型     | 说明                                           |
| ------------- | ------ | -------------------------------------------- |
| status        | int64  | 物流主状态：100-已揽收，200-运输中，300-派送中，400-已签收，500-异常 |
| tracks[].time | string | 轨迹时间（RFC3339 格式）                             |

---

## 模块七：售后服务

### 7.1 申请售后

- **认证**：需要 JWT
- **幂等**：是（通过 `idempotency_key`）
- **请求**：`POST /api/v1/order/aftersale/apply`
- **请求头**：`Authorization: Bearer <access_token>`

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

| 字段              | 类型       | 必填  | 说明                  |
| --------------- | -------- | --- | ------------------- |
| order_sn        | string   | 是   | 订单号                 |
| sku_id          | uint64   | 是   | SKU ID              |
| quantity        | int64    | 是   | 退款数量（支持部分退款）        |
| type            | int64    | 是   | 售后类型：1-仅退款，2-退货退款   |
| reason          | string   | 是   | 售后原因（1-200 字符）      |
| apply_amount    | int64    | 是   | 申请退款金额（分）           |
| images          | []string | 否   | 凭证图片 URL 列表（最多 9 张） |
| idempotency_key | string   | 是   | 幂等键                 |

**响应体**：

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "msg": "申请成功"
  }
}
```

**注意事项**：

- 同一商品只能申请一次售后
- 超出售后申请时限（通常为确认收货后 7 天）不可申请
- 退货退款（type=2）需后续提交物流信息

---

### 7.2 查询售后详情

- **认证**：需要 JWT
- **请求**：`GET /api/v1/order/aftersale/detail/:id`
- **请求头**：`Authorization: Bearer <access_token>`

**路径参数**：

| 参数  | 类型     | 说明     |
| --- | ------ | ------ |
| id  | uint64 | 售后单 ID |

**响应体**：

```json
{
  "code": 0,
  "msg": "success",
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

| data 字段              | 类型     | 说明                            |
| -------------------- | ------ | ----------------------------- |
| after_sale_type      | int64  | 1-仅退款，2-退货退款                  |
| status               | int64  | 1-待审核，2-待退货，3-退款中，4-已完成，5-已拒绝 |
| return_tracking_corp | string | 退货物流公司（退货退款时）                 |
| return_tracking_sn   | string | 退货物流单号（退货退款时）                 |

---

### 7.3 取消售后申请

- **认证**：需要 JWT
- **请求**：`POST /api/v1/order/aftersale/cancel/:id`
- **请求头**：`Authorization: Bearer <access_token>`

**路径参数**：

| 参数  | 类型     | 说明     |
| --- | ------ | ------ |
| id  | uint64 | 售后单 ID |

**响应体**：`{"code":0,"msg":"success","data":null}`

**注意事项**：

- **仅待审核（1）状态**可取消

---

### 7.4 获取售后单列表

- **认证**：需要 JWT
- **请求**：`GET /api/v1/order/aftersale/list`
- **请求头**：`Authorization: Bearer <access_token>`
- **请求参数**（Query）：

| 参数     | 类型    | 必填  | 说明                                         |
| ------ | ----- | --- | ------------------------------------------ |
| status | int64 | 否   | 售后单状态：1-待审核，2-待退货，3-退款中，4-已完成，5-已拒绝，不传则查全部 |
| page   | int64 | 否   | 默认 1                                       |
| size   | int64 | 否   | 默认 10                                      |

**响应体**：

```json
{
  "code": 0,
  "msg": "success",
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

## 模块八：支付服务

### 8.1 创建支付单

- **认证**：需要 JWT
- **幂等**：是（通过 `idempotency_key`）
- **请求**：`POST /api/v1/payment/create`
- **请求头**：`Authorization: Bearer <access_token>`

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
| openid          | string | 条件必填 | 微信 JSAPI 支付时必传                  |
| idempotency_key | string | 是    | 幂等键                             |

**响应体**：

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "payment_sn": "PAY202604090001",
    "pay_data": "prepay_id=wx123..."
  }
}
```

| data 字段    | 类型     | 说明                     |
| ---------- | ------ | ---------------------- |
| payment_sn | string | 支付流水号                  |
| pay_data   | string | 支付参数（JSON 字符串或表单 HTML） |

---

### 8.2 根据支付流水号查询支付状态

- **认证**：需要 JWT
- **请求**：`GET /api/v1/payment/status/payment/:payment_sn`
- **请求头**：`Authorization: Bearer <access_token>`

**路径参数**：

| 参数         | 类型     | 说明    |
| ---------- | ------ | ----- |
| payment_sn | string | 支付流水号 |

**响应体**：

```json
{
  "code": 0,
  "msg": "success",
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

| data 字段          | 类型     | 说明                               |
| ---------------- | ------ | -------------------------------- |
| status           | int64  | 0-待支付, 1-成功, 2-失败, 3-退款中, 4-退款成功 |
| channel_order_sn | string | 渠道订单号                            |
| transaction_id   | string | 第三方交易号                           |

---

### 8.3 根据订单号查询支付状态

- **认证**：需要 JWT
- **请求**：`GET /api/v1/payment/status/order/:order_sn`
- **请求头**：`Authorization: Bearer <access_token>`

**路径参数**：

| 参数       | 类型     | 说明  |
| -------- | ------ | --- |
| order_sn | string | 订单号 |

**响应体**：同 8.2。

---

### 8.4 关闭支付单

- **认证**：需要 JWT
- **请求**：`POST /api/v1/payment/close/:payment_sn`
- **请求头**：`Authorization: Bearer <access_token>`

**路径参数**：

| 参数         | 类型     | 说明    |
| ---------- | ------ | ----- |
| payment_sn | string | 支付流水号 |

**响应体**：`{"code":0,"msg":"success","data":null}`

**注意事项**：

- **仅未支付（status=0）** 状态可关闭

---

### 8.5 申请退款

- **认证**：需要 JWT
- **幂等**：是（通过 `idempotency_key`）
- **请求**：`POST /api/v1/payment/refund/apply`
- **请求头**：`Authorization: Bearer <access_token>`

**请求体**：

```json
{
  "order_sn": "202604091234567890",
  "payment_sn": "PAY202604090001",
  "amount": 789900,
  "reason": "质量问题",
  "type": 1,
  "images": ["https://..."],
  "idempotency_key": "uuid-xxxx"
}
```

| 字段              | 类型       | 必填  | 说明             |
| --------------- | -------- | --- | -------------- |
| order_sn        | string   | 是   | 订单号            |
| payment_sn      | string   | 否   | 支付流水号（不传则自动查找） |
| amount          | int64    | 是   | 退款金额（分）        |
| reason          | string   | 是   | 退款原因（1-200 字符） |
| type            | int64    | 是   | 1-仅退款，2-退货退款   |
| images          | []string | 否   | 退款凭证图片（最多 9 张） |
| idempotency_key | string   | 是   | 幂等键            |

**响应体**：

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "refund_sn": "REF202604090001",
    "status": 0,
    "status_desc": "处理中"
  }
}
```

| data 字段   | 类型     | 说明              |
| --------- | ------ | --------------- |
| refund_sn | string | 退款单号            |
| status    | int64  | 0-处理中，1-成功，2-失败 |

---

### 8.6 查询退款详情

- **认证**：需要 JWT
- **请求**：`GET /api/v1/payment/refund/detail/:order_sn`
- **请求头**：`Authorization: Bearer <access_token>`

**路径参数**：

| 参数       | 类型     | 说明  |
| -------- | ------ | --- |
| order_sn | string | 订单号 |

**响应体**：

```json
{
  "code": 0,
  "msg": "success",
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

### 8.7 支付宝异步回调

- **请求**：`POST /api/v1/payment/callback/alipay`
- **说明**：支付宝服务器异步通知回调（服务端自行保证幂等）

**响应体**：纯文本 `success`

---

### 8.8 微信支付异步回调

- **请求**：`POST /api/v1/payment/callback/wechat`
- **说明**：微信支付服务器异步通知回调（服务端自行保证幂等）

**响应体**：

```json
{
  "code": "SUCCESS",
  "message": "OK"
}
```

---

## 模块九：营销服务（优惠券）

### 9.1 领券中心列表

- **认证**：需要 JWT（已登录用户可查看 `is_claimed` 状态）
- **请求**：`GET /api/v1/marketing/coupon/list`
- **请求头**：`Authorization: Bearer <access_token>`

**请求参数**（Query）：

| 参数   | 类型    | 必填  | 说明    |
| ---- | ----- | --- | ----- |
| page | int64 | 否   | 默认 1  |
| size | int64 | 否   | 默认 10 |

**响应体**：

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "total": 50,
    "list": [
      {
        "id": 101,
        "name": "新人专享券",
        "type": 1,
        "threshold_amount": 10000,
        "reduce_amount": 1000,
        "discount_rate": 80,
        "max_discount_amount": 5000,
        "start_time": 1746000000,
        "end_time": 1748591999,
        "description": "满100减10",
        "is_claimed": false
      }
    ]
  }
}
```

| data 列表字段           | 类型    | 说明                   |
| ------------------- | ----- | -------------------- |
| type                | int64 | 1-满减券，2-折扣券，3-无门槛券   |
| threshold_amount    | int64 | 使用门槛（分）              |
| reduce_amount       | int64 | 满减金额（分），仅满减券有效       |
| discount_rate       | int64 | 折扣率：80 表示 8 折，仅折扣券有效 |
| max_discount_amount | int64 | 折扣封顶金额（分），仅折扣券有效     |
| start_time          | int64 | 生效开始时间（Unix 时间戳秒）    |
| end_time            | int64 | 生效结束时间（Unix 时间戳秒）    |
| is_claimed          | bool  | 当前登录用户是否已领取（需认证后）    |

---

### 9.2 领取优惠券

- **认证**：需要 JWT
- **幂等**：是（通过 `idempotency_key`）
- **请求**：`POST /api/v1/marketing/coupon/claim`
- **请求头**：`Authorization: Bearer <access_token>`

**请求体**：

```json
{
  "coupon_id": 101,
  "idempotency_key": "uuid-xxxx"
}
```

| 字段              | 类型     | 必填  | 说明     |
| --------------- | ------ | --- | ------ |
| coupon_id       | uint64 | 是   | 优惠券 ID |
| idempotency_key | string | 是   | 幂等键    |

**响应体**：

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "user_coupon_id": 1001,
    "expire_time": 1748591999
  }
}
```

| data 字段        | 类型     | 说明                  |
| -------------- | ------ | ------------------- |
| user_coupon_id | uint64 | 用户优惠券记录 ID          |
| expire_time    | int64  | 领取后的过期时间（Unix 时间戳秒） |

---

### 9.3 我的优惠券列表

- **认证**：需要 JWT
- **请求**：`GET /api/v1/marketing/coupon/mine`
- **请求头**：`Authorization: Bearer <access_token>`
- **请求参数**（Query）：

| 参数     | 类型    | 必填  | 说明                    |
| ------ | ----- | --- | --------------------- |
| status | int64 | 否   | 0-未使用（默认），1-已使用，2-已过期 |
| page   | int64 | 否   | 默认 1                  |
| size   | int64 | 否   | 默认 10                 |

**响应体**：

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "total": 10,
    "list": [
      {
        "id": 1001,
        "coupon_id": 101,
        "name": "新人专享券",
        "type": 1,
        "status": 0,
        "expire_time": 1748591999,
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

| data 列表字段          | 类型     | 说明                   |
| ------------------ | ------ | -------------------- |
| status             | int64  | 0-未使用，1-已使用，2-已过期    |
| expire_time        | int64  | 用户持有的过期时间（Unix 时间戳秒） |
| order_sn           | string | 使用该券的订单号（已使用时）       |
| unavailable_reason | string | 不可用原因（结算页用）          |

---

### 9.4 结算页可用优惠券

- **认证**：需要 JWT
- **请求**：`POST /api/v1/marketing/coupon/available`
- **请求头**：`Authorization: Bearer <access_token>`

**请求体**：

```json
{
  "order_amount": 799900
}
```

| 字段           | 类型    | 必填  | 说明         |
| ------------ | ----- | --- | ---------- |
| order_amount | int64 | 是   | 当前订单总金额（分） |

**响应体**：

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "available": [
      {
        "id": 1001,
        "coupon_id": 101,
        "name": "新人专享券",
        "type": 1,
        "status": 0,
        "expire_time": 1748591999,
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
        "expire_time": 1748591999,
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

### 9.5 使用优惠券（内部接口）

- **认证**：需要 JWT
- **请求**：`POST /internal/v1/coupon/use`
- **说明**：订单服务调用，锁定并扣减优惠券（幂等）

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
  "code": 0,
  "msg": "success",
  "data": {
    "success": true,
    "discount_amount": 1000,
    "error_msg": ""
  }
}
```

---

### 9.6 解锁优惠券（内部接口）

- **认证**：需要 JWT
- **请求**：`POST /internal/v1/coupon/unlock`
- **说明**：订单取消时回滚优惠券（幂等）

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
  "code": 0,
  "msg": "success",
  "data": {
    "success": true,
    "error_msg": ""
  }
}
```

---

## 模块十：管理后台

### 10.1 管理员登录

- **请求**：`POST /api/v1/admin/login`

**注意**：此接口未使用统一响应包装，直接返回 `AdminLoginResp` 对象。

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
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "access_expire": 1744189445
}
```

| data 字段       | 类型     | 说明                |
| ------------- | ------ | ----------------- |
| access_token  | string | JWT 访问令牌          |
| access_expire | int64  | 令牌过期时间（Unix 时间戳秒） |

---

### 10.2 获取管理员信息

- **认证**：需要 JWT
- **请求**：`GET /api/v1/admin/info`
- **请求头**：`Authorization: Bearer <access_token>`

**响应体**：

```json
{
  "code": 0,
  "msg": "success",
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

### 10.3 修改密码

- **认证**：需要 JWT
- **请求**：`PUT /api/v1/admin/password`
- **请求头**：`Authorization: Bearer <access_token>`

**请求体**：

```json
{
  "old_password": "123456",
  "new_password": "654321"
}
```

| 字段           | 类型     | 必填  | 说明          |
| ------------ | ------ | --- | ----------- |
| old_password | string | 是   | 原密码         |
| new_password | string | 是   | 新密码（6-20 位） |

**响应体**：`{"code":0,"msg":"success","data":null}`

---

### 10.4 用户管理 - 用户列表

- **认证**：需要 JWT
- **请求**：`GET /api/v1/admin/user/list`
- **请求头**：`Authorization: Bearer <access_token>`
- **请求参数**（Query）：

| 参数      | 类型     | 必填  | 说明                    |
| ------- | ------ | --- | --------------------- |
| keyword | string | 否   | 搜索关键词（手机号/昵称）         |
| status  | int64  | 否   | 用户状态：1-正常，2-限制下单，3-封禁 |
| page    | int64  | 否   | 默认 1                  |
| size    | int64  | 否   | 默认 10                 |

**响应体**：

```json
{
  "code": 0,
  "msg": "success",
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

### 10.5 用户管理 - 用户详情

- **认证**：需要 JWT
- **请求**：`GET /api/v1/admin/user/detail/:id`
- **请求头**：`Authorization: Bearer <access_token>`

**路径参数**：

| 参数  | 类型     | 说明    |
| --- | ------ | ----- |
| id  | uint64 | 用户 ID |

**响应体**：

```json
{
  "code": 0,
  "msg": "success",
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

### 10.6 用户管理 - 拉黑用户

- **认证**：需要 JWT
- **请求**：`POST /api/v1/admin/user/blacklist`
- **请求头**：`Authorization: Bearer <access_token>`

**请求体**：

```json
{
  "id": 1,
  "reason": "恶意下单"
}
```

| 字段     | 类型     | 必填  | 说明    |
| ------ | ------ | --- | ----- |
| id     | uint64 | 是   | 用户 ID |
| reason | string | 否   | 拉黑原因  |

**响应体**：`{"code":0,"msg":"success","data":null}`

---

### 10.7 用户管理 - 恢复用户

- **认证**：需要 JWT
- **请求**：`POST /api/v1/admin/user/recover`
- **请求头**：`Authorization: Bearer <access_token>`

**请求体**：

```json
{
  "id": 1
}
```

| 字段  | 类型     | 必填  | 说明    |
| --- | ------ | --- | ----- |
| id  | uint64 | 是   | 用户 ID |

**响应体**：`{"code":0,"msg":"success","data":null}`

---

### 10.8 分类管理 - 分类列表

- **认证**：需要 JWT
- **请求**：`GET /api/v1/admin/category/list`
- **请求头**：`Authorization: Bearer <access_token>`

**响应体**：

```json
{
  "code": 0,
  "msg": "success",
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

### 10.9 分类管理 - 保存分类

- **认证**：需要 JWT
- **请求**：`POST /api/v1/admin/category/save`
- **请求头**：`Authorization: Bearer <access_token>`

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

| 字段        | 类型     | 必填  | 说明                |
| --------- | ------ | --- | ----------------- |
| id        | uint64 | 否   | 分类 ID（新增时为 0 或不传） |
| parent_id | uint64 | 是   | 父分类 ID（顶级为 0）     |
| name      | string | 是   | 分类名称              |
| icon      | string | 否   | 分类图标 URL          |
| sort      | int64  | 是   | 排序值               |

**响应体**：`{"code":0,"msg":"success","data":null}`

---

### 10.10 分类管理 - 删除分类

- **认证**：需要 JWT
- **请求**：`DELETE /api/v1/admin/category/:id`
- **请求头**：`Authorization: Bearer <access_token>`

**路径参数**：

| 参数  | 类型     | 说明    |
| --- | ------ | ----- |
| id  | uint64 | 分类 ID |

**响应体**：`{"code":0,"msg":"success","data":null}`

**注意事项**：

- 含有子分类的分类无法删除，需先删除子分类

---

### 10.11 商品管理 - 商品列表

- **认证**：需要 JWT
- **请求**：`GET /api/v1/admin/product/list`
- **请求头**：`Authorization: Bearer <access_token>`
- **请求参数**（Query）：

| 参数          | 类型     | 必填  | 说明           |
| ----------- | ------ | --- | ------------ |
| keyword     | string | 否   | 搜索关键词（商品名称）  |
| category_id | uint64 | 否   | 分类 ID        |
| status      | int64  | 否   | 状态：1-上架，2-下架 |
| page        | int64  | 否   | 默认 1         |
| size        | int64  | 否   | 默认 10        |

**响应体**：

```json
{
  "code": 0,
  "msg": "success",
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

### 10.12 商品管理 - 商品详情

- **认证**：需要 JWT
- **请求**：`GET /api/v1/admin/product/detail/:id`
- **请求头**：`Authorization: Bearer <access_token>`

**路径参数**：

| 参数  | 类型     | 说明        |
| --- | ------ | --------- |
| id  | uint64 | 商品 SPU ID |

**响应体**：

```json
{
  "code": 0,
  "msg": "success",
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

### 10.13 商品管理 - 保存商品

- **认证**：需要 JWT
- **请求**：`POST /api/v1/admin/product/save`
- **请求头**：`Authorization: Bearer <access_token>`

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

| 字段                  | 类型       | 必填  | 说明            |
| ------------------- | -------- | --- | ------------- |
| id                  | uint64   | 否   | 商品 ID（新增时为 0） |
| category_id         | uint64   | 是   | 分类 ID         |
| brand               | string   | 是   | 品牌名称          |
| name                | string   | 是   | 商品名称          |
| description         | string   | 是   | 商品描述          |
| content             | string   | 是   | 商品详情 HTML     |
| default_pic         | string   | 是   | 默认图片 URL      |
| banner_pics         | []string | 否   | 轮播图列表         |
| video_url           | string   | 否   | 视频链接          |
| weight              | int64    | 是   | 重量（克）         |
| freight_template_id | uint64   | 是   | 运费模板 ID       |
| status              | int64    | 是   | 1-上架，2-下架     |
| skus                | []object | 是   | SKU 列表        |

**响应体**：`{"code":0,"msg":"success","data":null}`

---

### 10.14 商品管理 - 更新商品状态

- **认证**：需要 JWT
- **请求**：`POST /api/v1/admin/product/status`
- **请求头**：`Authorization: Bearer <access_token>`

**请求体**：

```json
{
  "id": 1,
  "status": 2
}
```

| 字段     | 类型     | 必填  | 说明        |
| ------ | ------ | --- | --------- |
| id     | uint64 | 是   | 商品 ID     |
| status | int64  | 是   | 1-上架，2-下架 |

**响应体**：`{"code":0,"msg":"success","data":null}`

---

### 10.15 商品管理 - 删除商品

- **认证**：需要 JWT
- **请求**：`DELETE /api/v1/admin/product/:id`
- **请求头**：`Authorization: Bearer <access_token>`

**路径参数**：

| 参数  | 类型     | 说明        |
| --- | ------ | --------- |
| id  | uint64 | 商品 SPU ID |

**响应体**：`{"code":0,"msg":"success","data":null}`

---

### 10.16 订单管理 - 订单列表

- **认证**：需要 JWT
- **请求**：`GET /api/v1/admin/order/list`
- **请求头**：`Authorization: Bearer <access_token>`
- **请求参数**（Query）：

| 参数         | 类型     | 必填  | 说明                                             |
| ---------- | ------ | --- | ---------------------------------------------- |
| status     | int64  | 否   | 订单状态：10-待付款,20-待发货,30-待收货,40-已完成,50-已取消,60-售后中 |
| order_sn   | string | 否   | 订单号精确搜索                                        |
| start_time | string | 否   | 开始时间（格式：`2026-04-01 00:00:00`）                 |
| end_time   | string | 否   | 结束时间（格式：`2026-04-30 23:59:59`）                 |
| page       | int64  | 否   | 默认 1                                           |
| size       | int64  | 否   | 默认 10                                          |

**响应体**：

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "total": 100,
    "list": [
      {
        "order_sn": "202604091234567890",
        "user_id": 1,
        "user_phone": "13800138000",
        "pay_amount": 799900,
        "status": 20,
        "status_desc": "待发货",
        "created_at": "2026-04-09T15:00:00+08:00"
      }
    ]
  }
}
```

---

### 10.17 订单管理 - 订单详情

- **认证**：需要 JWT
- **请求**：`GET /api/v1/admin/order/detail/:order_sn`
- **请求头**：`Authorization: Bearer <access_token>`

**路径参数**：

| 参数       | 类型     | 说明  |
| -------- | ------ | --- |
| order_sn | string | 订单号 |

**响应体**：

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "order_sn": "202604091234567890",
    "user_id": 1,
    "user_phone": "13800138000",
    "status": 20,
    "status_desc": "待发货",
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

| data 字段          | 类型     | 说明          |
| ---------------- | ------ | ----------- |
| pay_type         | int64  | 1-微信，2-支付宝  |
| pay_type_desc    | string | 支付方式描述      |
| delivery_company | string | 物流公司（发货后非空） |
| payment_sn       | string | 支付流水号       |

---

### 10.18 订单管理 - 发货

- **认证**：需要 JWT
- **请求**：`POST /api/v1/admin/order/ship`
- **请求头**：`Authorization: Bearer <access_token>`

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

**响应体**：`{"code":0,"msg":"success","data":null}`

---

### 10.19 售后管理 - 退款列表

- **认证**：需要 JWT
- **请求**：`GET /api/v1/admin/refund/list`
- **请求头**：`Authorization: Bearer <access_token>`
- **请求参数**（Query）：

| 参数     | 类型    | 必填  | 说明                                     |
| ------ | ----- | --- | -------------------------------------- |
| status | int64 | 否   | 状态：1-待审核，2-待退货，3-退款中，4-已完成，5-已拒绝，不传查全部 |
| page   | int64 | 否   | 默认 1                                   |
| size   | int64 | 否   | 默认 10                                  |

**响应体**：

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "total": 10,
    "list": [
      {
        "id": 1,
        "refund_sn": "RF202604091234567890",
        "order_sn": "202604091234567890",
        "user_id": 1,
        "refund_amount": 799900,
        "status": 1,
        "status_desc": "待审核",
        "created_at": "2026-04-09T16:00:00+08:00"
      }
    ]
  }
}
```

---

### 10.20 售后管理 - 处理退款

- **认证**：需要 JWT
- **请求**：`POST /api/v1/admin/refund/handle`
- **请求头**：`Authorization: Bearer <access_token>`

**请求体**：

```json
{
  "refund_sn": "RF202604091234567890",
  "action": 1,
  "remark": "同意退款"
}
```

| 字段        | 类型     | 必填  | 说明        |
| --------- | ------ | --- | --------- |
| refund_sn | string | 是   | 退款单号      |
| action    | int64  | 是   | 1-通过，2-拒绝 |
| remark    | string | 否   | 处理备注      |

**响应体**：`{"code":0,"msg":"success","data":null}`

---

### 10.21 优惠券管理 - 优惠券列表

- **认证**：需要 JWT
- **请求**：`GET /api/v1/admin/coupon/list`
- **请求头**：`Authorization: Bearer <access_token>`
- **请求参数**（Query）：

| 参数     | 类型     | 必填  | 说明           |
| ------ | ------ | --- | ------------ |
| status | int64  | 否   | 状态：1-启用，2-禁用 |
| name   | string | 否   | 优惠券名称搜索      |
| page   | int64  | 否   | 默认 1         |
| size   | int64  | 否   | 默认 10        |

**响应体**：

```json
{
  "code": 0,
  "msg": "success",
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
        "start_time": 1746000000,
        "end_time": 1748591999
      }
    ]
  }
}
```

| data 列表字段  | 类型    | 说明                 |
| ---------- | ----- | ------------------ |
| type       | int64 | 1-满减券，2-折扣券，3-无门槛券 |
| start_time | int64 | 生效开始时间（Unix 时间戳秒）  |
| end_time   | int64 | 失效结束时间（Unix 时间戳秒）  |

---

### 10.22 优惠券管理 - 保存优惠券

- **认证**：需要 JWT
- **请求**：`POST /api/v1/admin/coupon/save`
- **请求头**：`Authorization: Bearer <access_token>`

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
  "start_time": 1746000000,
  "end_time": 1748591999,
  "status": 1,
  "description": "新用户专享"
}
```

| 字段                  | 类型     | 必填  | 说明                 |
| ------------------- | ------ | --- | ------------------ |
| id                  | uint64 | 否   | 优惠券 ID（新增时为 0）     |
| name                | string | 是   | 优惠券名称              |
| type                | int64  | 是   | 1-满减券，2-折扣券，3-无门槛券 |
| threshold_amount    | int64  | 是   | 使用门槛（分），0 表示无门槛    |
| reduce_amount       | int64  | 否   | 满减金额（分），仅满减券有效     |
| discount_rate       | int64  | 否   | 折扣率（万比分比），仅折扣券有效   |
| max_discount_amount | int64  | 否   | 折扣封顶金额（分），仅折扣券有效   |
| total_quantity      | int64  | 是   | 发行总量               |
| per_user_limit      | int64  | 是   | 每人限领数量             |
| start_time          | int64  | 是   | 生效开始时间（Unix 时间戳秒）  |
| end_time            | int64  | 是   | 失效结束时间（Unix 时间戳秒）  |
| status              | int64  | 是   | 1-启用，2-禁用          |
| description         | string | 否   | 使用说明               |

**响应体**：`{"code":0,"msg":"success","data":null}`

---

### 10.23 优惠券管理 - 更新优惠券状态

- **认证**：需要 JWT
- **请求**：`POST /api/v1/admin/coupon/status`
- **请求头**：`Authorization: Bearer <access_token>`

**请求体**：

```json
{
  "id": 1,
  "status": 2
}
```

| 字段     | 类型     | 必填  | 说明        |
| ------ | ------ | --- | --------- |
| id     | uint64 | 是   | 优惠券 ID    |
| status | int64  | 是   | 1-启用，2-禁用 |

**响应体**：`{"code":0,"msg":"success","data":null}`

---

### 10.24 操作日志

- **认证**：需要 JWT
- **请求**：`GET /api/v1/admin/operate/logs`
- **请求头**：`Authorization: Bearer <access_token>`
- **请求参数**（Query）：

| 参数   | 类型    | 必填  | 说明    |
| ---- | ----- | --- | ----- |
| page | int64 | 否   | 默认 1  |
| size | int64 | 否   | 默认 10 |

**响应体**：

```json
{
  "code": 0,
  "msg": "success",
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

## 附录：订单状态流转

```
┌──────┐  支付    ┌──────┐  发货    ┌──────┐  确认收货  ┌──────┐
│待付款│ ──────→ │待发货│ ──────→ │待收货│ ────────→ │已完成│
│ (10) │         │ (20) │         │ (30) │           │ (40) │
└──┬───┘         └──────┘         └──────┘           └──────┘
   │                                                      │
   │ 取消                ┌──────┐   申请售后               │
   └──────────────────→ │已取消│ ←────────────────────────┘
                        │ (50) │
                        └──────┘        ┌──────┐
                                        │售后中│
                                        │ (60) │
                                        └──────┘
```

- **待付款（10）**：可取消；超时未支付自动取消
- **待发货（20）**：支付成功后进入；管理员后台发货
- **待收货（30）**：发货后进入；用户确认收货或系统超时自动确认
- **已完成（40）**：可申请售后
- **已取消（50）**：仅限待付款状态取消
- **售后中（60）**：已完成订单中存在进行中的售后单

---

## 附录：售后状态流转

```
       ┌──────┐
       │待审核│ ──── 管理员审核 ────→ ┌──通过──→ ┌──────┐
       │ (1)  │                       │         │退款中│
       └──┬───┘                       │         │ (3)  │
          │ 取消                      │         └──┬───┘
          ↓                           │            │
       ┌──────┐                       │            ↓
       │已取消│                       └──拒绝──→ ┌──────┐
       └──────┘                                   │已拒绝│
                                                  │ (5)  │
                                                  └──────┘
```

- **仅退款（type=1）**：审核通过 → 退款中 → 已完成
- **退货退款（type=2）**：审核通过 → 待退货（用户寄回）→ 商家收货 → 退款中 → 已完成

---

## 附录：错误码

### 通用错误

| 错误码 | 含义      |
| --- | ------- |
| 0   | 成功      |
| 400 | 请求错误    |
| 500 | 服务器内部错误 |

### 参数与请求错误 (11xxx)

| 错误码   | 含义          |
| ----- | ----------- |
| 11001 | 请求参数无效      |
| 11002 | 缺少必要参数      |
| 11003 | 参数类型错误      |
| 11004 | 参数超出范围      |
| 11005 | 分页参数无效      |
| 11006 | 时间格式无效      |
| 11007 | object 格式错误 |

### 认证与授权错误 (12xxx)

| 错误码   | 含义        |
| ----- | --------- |
| 12001 | 未认证，请先登录  |
| 12002 | Token 无效  |
| 12005 | 权限不足      |
| 12006 | 账号或密码错误   |
| 12007 | 验证码错误     |
| 12010 | 手机号已注册    |
| 12011 | 访问过于频繁    |
| 12012 | 两次输入密码不一致 |
| 12013 | 新密码与旧密码一致 |

### 用户模块 (2xxxx)

| 错误码   | 含义          |
| ----- | ----------- |
| 20001 | 用户不存在       |
| 20002 | 账号已被禁用      |
| 20004 | 账号已注销       |
| 20005 | 账号受限（限制下单等） |
| 20009 | 账号已解封       |
| 20010 | 风控日志不存在     |

### 商品模块 (3xxxx)

| 错误码   | 含义           |
| ----- | ------------ |
| 30001 | 商品不存在        |
| 30002 | 商品已下架        |
| 30004 | SKU 不存在      |
| 30007 | 分类不存在        |
| 30008 | 分类含有子分类，无法删除 |
| 30011 | 分类被禁用        |
| 30012 | 数量非法         |

### 订单模块 (4xxxx)

| 错误码   | 含义             |
| ----- | -------------- |
| 40001 | 订单不存在          |
| 40002 | 订单状态不允许当前操作    |
| 40003 | 订单取消失败（非待支付状态） |
| 40005 | 订单已过期          |
| 40008 | 重复提交订单（幂等冲突）   |
| 40010 | 地址不属于当前用户      |
| 40011 | 运费模板不存在        |
| 40012 | 运费计算失败         |
| 40013 | 预下单失败          |

### 支付模块 (5xxxx)

| 错误码   | 含义         |
| ----- | ---------- |
| 50001 | 支付单不存在     |
| 50007 | 退款单不存在     |
| 50008 | 退款状态不允许操作  |
| 50009 | 退款金额超过可退金额 |

### 营销/优惠券模块 (6xxxx)

| 错误码   | 含义             |
| ----- | -------------- |
| 60001 | 优惠券不存在         |
| 60002 | 优惠券已过期         |
| 60003 | 优惠券未开始         |
| 60004 | 优惠券已领取过        |
| 60005 | 优惠券库存不足        |
| 60006 | 优惠券不可用（不满足门槛等） |

### 地址模块 (7xxxx)

| 错误码   | 含义                    |
| ----- | --------------------- |
| 70001 | 地址不存在                 |
| 70002 | 地址数量超出限制（最多 20 条）     |
| 70003 | 默认地址不可直接删除，请先设置其他默认地址 |

### 售后模块 (9xxxx)

| 错误码   | 含义             |
| ----- | -------------- |
| 90001 | 售后单不存在         |
| 90002 | 售后单状态不允许操作     |
| 90003 | 该商品已申请过售后      |
| 90005 | 售后原因无效         |
| 90007 | 金额无效           |
| 90008 | 数量无效           |
| 90009 | 仅退货退款订单可提交物流信息 |

### 管理员模块 (10xxx)

| 错误码   | 含义        |
| ----- | --------- |
| 10001 | 管理员不存在    |
| 10002 | 管理员已被禁用   |
| 10003 | 管理员用户名已存在 |
| 10004 | 角色不存在     |

---

## 附录：公共分页参数

所有支持分页的查询接口均使用以下 query 参数：

| 参数   | 类型    | 必填  | 说明                |
| ---- | ----- | --- | ----------------- |
| page | int64 | 否   | 页码，默认 1，最小 1      |
| size | int64 | 否   | 每页数量，默认 10，最大 100 |

分页响应结构：

```json
{
  "total": 100,
  "list": [ ... ]
}
```

---

## 附录：API 路径汇总

### 前台用户端

| 模块  | 方法     | 路径                                           | 认证  | 幂等  |
| --- | ------ | -------------------------------------------- | --- | --- |
| 用户  | POST   | `/api/v1/user/captcha`                       | 否   | 是   |
| 用户  | POST   | `/api/v1/user/login`                         | 否   | 否   |
| 用户  | POST   | `/api/v1/user/login/mobile`                  | 否   | 否   |
| 用户  | POST   | `/api/v1/user/register`                      | 否   | 是   |
| 用户  | POST   | `/api/v1/user/token/refresh`                 | 否   | 否   |
| 用户  | POST   | `/api/v1/user/password/reset`                | 否   | 是   |
| 用户  | POST   | `/api/v1/user/upload/avatar/callback`        | 否   | 否*  |
| 用户  | GET    | `/api/v1/user/info`                          | 是   | 否   |
| 用户  | PUT    | `/api/v1/user/info`                          | 是   | 否   |
| 用户  | PUT    | `/api/v1/user/password/update`               | 是   | 否   |
| 用户  | PUT    | `/api/v1/user/phone`                         | 是   | 是   |
| 用户  | DELETE | `/api/v1/user/account`                       | 是   | 否   |
| 用户  | POST   | `/api/v1/user/logout`                        | 是   | 否   |
| 用户  | POST   | `/api/v1/user/upload/avatar/token`           | 是   | 否   |
| 地址  | POST   | `/api/v1/address/`                           | 是   | 是   |
| 地址  | GET    | `/api/v1/address/list`                       | 是   | 否   |
| 地址  | PUT    | `/api/v1/address/`                           | 是   | 否   |
| 地址  | DELETE | `/api/v1/address/:id`                        | 是   | 否   |
| 地址  | PATCH  | `/api/v1/address/default/:id`                | 是   | 否   |
| 商品  | GET    | `/api/v1/product/category/tree`              | 否   | 否   |
| 商品  | GET    | `/api/v1/product/category/:id`               | 否   | 否   |
| 商品  | GET    | `/api/v1/product/list`                       | 否   | 否   |
| 商品  | GET    | `/api/v1/product/detail/:id`                 | 否   | 否   |
| 商品  | GET    | `/api/v1/product/recommend`                  | 否   | 否   |
| 商品  | POST   | `/api/v1/product/skus/batch`                 | 否   | 否   |
| 搜索  | GET    | `/api/v1/search/products`                    | 否   | 否   |
| 搜索  | GET    | `/api/v1/search/suggest`                     | 否   | 否   |
| 搜索  | GET    | `/api/v1/search/hot-keywords`                | 否   | 否   |
| 搜索  | GET    | `/api/v1/search/history`                     | 是   | 否   |
| 搜索  | DELETE | `/api/v1/search/history`                     | 是   | 否   |
| 购物车 | POST   | `/api/v1/order/cart/update`                  | 是   | 是   |
| 购物车 | POST   | `/api/v1/order/cart/select`                  | 是   | 否   |
| 购物车 | GET    | `/api/v1/order/cart/list`                    | 是   | 否   |
| 购物车 | DELETE | `/api/v1/order/cart/remove`                  | 是   | 否   |
| 购物车 | DELETE | `/api/v1/order/cart/clear`                   | 是   | 否   |
| 订单  | POST   | `/api/v1/order/pre-order`                    | 是   | 否   |
| 订单  | POST   | `/api/v1/order/create`                       | 是   | 是   |
| 订单  | GET    | `/api/v1/order/list`                         | 是   | 否   |
| 订单  | GET    | `/api/v1/order/detail/:order_sn`             | 是   | 否   |
| 订单  | POST   | `/api/v1/order/cancel/:order_sn`             | 是   | 否   |
| 订单  | POST   | `/api/v1/order/confirm/:order_sn`            | 是   | 否   |
| 物流  | GET    | `/api/v1/order/delivery/track/:order_sn`     | 是   | 否   |
| 售后  | POST   | `/api/v1/order/aftersale/apply`              | 是   | 是   |
| 售后  | GET    | `/api/v1/order/aftersale/detail/:id`         | 是   | 否   |
| 售后  | POST   | `/api/v1/order/aftersale/cancel/:id`         | 是   | 否   |
| 售后  | GET    | `/api/v1/order/aftersale/list`               | 是   | 否   |
| 支付  | POST   | `/api/v1/payment/create`                     | 是   | 是   |
| 支付  | GET    | `/api/v1/payment/status/payment/:payment_sn` | 是   | 否   |
| 支付  | GET    | `/api/v1/payment/status/order/:order_sn`     | 是   | 否   |
| 支付  | POST   | `/api/v1/payment/close/:payment_sn`          | 是   | 否   |
| 支付  | POST   | `/api/v1/payment/refund/apply`               | 是   | 是   |
| 支付  | GET    | `/api/v1/payment/refund/detail/:order_sn`    | 是   | 否   |
| 支付  | POST   | `/api/v1/payment/callback/alipay`            | 否   | 否*  |
| 支付  | POST   | `/api/v1/payment/callback/wechat`            | 否   | 否*  |
| 营销  | GET    | `/api/v1/marketing/coupon/list`              | 是   | 否   |
| 营销  | POST   | `/api/v1/marketing/coupon/claim`             | 是   | 是   |
| 营销  | GET    | `/api/v1/marketing/coupon/mine`              | 是   | 否   |
| 营销  | POST   | `/api/v1/marketing/coupon/available`         | 是   | 否   |
| 营销  | POST   | `/internal/v1/coupon/use`                    | 是   | 是   |
| 营销  | POST   | `/internal/v1/coupon/unlock`                 | 是   | 是   |

> `*` = 服务端自行保证幂等

### 管理后台

| 模块   | 方法     | 路径                                     | 认证  | 幂等  |
| ---- | ------ | -------------------------------------- | --- | --- |
| 管理员  | POST   | `/api/v1/admin/login`                  | 否   | 否   |
| 管理员  | GET    | `/api/v1/admin/info`                   | 是   | 否   |
| 管理员  | PUT    | `/api/v1/admin/password`               | 是   | 否   |
| 用户管理 | GET    | `/api/v1/admin/user/list`              | 是   | 否   |
| 用户管理 | GET    | `/api/v1/admin/user/detail/:id`        | 是   | 否   |
| 用户管理 | POST   | `/api/v1/admin/user/blacklist`         | 是   | 否   |
| 用户管理 | POST   | `/api/v1/admin/user/recover`           | 是   | 否   |
| 分类管理 | GET    | `/api/v1/admin/category/list`          | 是   | 否   |
| 分类管理 | POST   | `/api/v1/admin/category/save`          | 是   | 否   |
| 分类管理 | DELETE | `/api/v1/admin/category/:id`           | 是   | 否   |
| 商品管理 | GET    | `/api/v1/admin/product/list`           | 是   | 否   |
| 商品管理 | GET    | `/api/v1/admin/product/detail/:id`     | 是   | 否   |
| 商品管理 | POST   | `/api/v1/admin/product/save`           | 是   | 否   |
| 商品管理 | POST   | `/api/v1/admin/product/status`         | 是   | 否   |
| 商品管理 | DELETE | `/api/v1/admin/product/:id`            | 是   | 否   |
| 订单管理 | GET    | `/api/v1/admin/order/list`             | 是   | 否   |
| 订单管理 | GET    | `/api/v1/admin/order/detail/:order_sn` | 是   | 否   |
| 订单管理 | POST   | `/api/v1/admin/order/ship`             | 是   | 否   |
| 售后管理 | GET    | `/api/v1/admin/refund/list`            | 是   | 否   |
| 售后管理 | POST   | `/api/v1/admin/refund/handle`          | 是   | 否   |
| 优惠券  | GET    | `/api/v1/admin/coupon/list`            | 是   | 否   |
| 优惠券  | POST   | `/api/v1/admin/coupon/save`            | 是   | 否   |
| 优惠券  | POST   | `/api/v1/admin/coupon/status`          | 是   | 否   |
| 操作日志 | GET    | `/api/v1/admin/operate/logs`           | 是   | 否   |
