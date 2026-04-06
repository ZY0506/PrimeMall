# 电商项目数据库表结构设计文档

## 1. 总体说明

本项目为电商平台，采用微服务架构，数据库按业务域拆分为 7 个独立数据库：

| 数据库 | 服务 | 说明 |
|--------|------|------|
| `shop_user` | 用户服务 | 用户注册登录、个人信息、地址管理、风控 |
| `shop_product` | 商品服务 | 商品分类、SPU、SKU、库存管理、运费模板 |
| `shop_order` | 订单服务 | 购物车、订单、订单商品快照 |
| `shop_payment` | 支付服务 | 支付记录、支付回调、退款记录 |
| `shop_marketing` | 营销服务 | 优惠券、秒杀活动、秒杀预占 |
| `shop_admin` | 后台管理服务 | 管理员、角色权限、操作日志、CMS 内容 |
| `shop_search` | 搜索服务 | 搜索历史、热搜词、ES 同步记录 |

字符集统一使用 `utf8mb4`，排序规则 `utf8mb4_unicode_ci`。

---

## 2. 用户服务 (`shop_user`)

### 2.1 用户主表 (`user`)

存储用户基本信息。

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | bigint unsigned | 主键，自增 |
| `phone` | varchar(20) | 手机号（唯一索引） |
| `password` | varchar(255) | 加密密码 |
| `nickname` | varchar(64) | 昵称 |
| `avatar` | varchar(500) | 头像 URL |
| `gender` | tinyint | 0:未知 1:男 2:女 |
| `birthday` | date | 生日 |
| `status` | tinyint | 1:正常 2:限制下单 3:封禁 |
| `last_login_time` | datetime | 最后登录时间 |
| `last_login_ip` | varchar(64) | 最后登录 IP |
| `ext_info` | json | 扩展信息 |
| `created_at` | datetime | 创建时间 |
| `updated_at` | datetime | 更新时间 |
| `deleted_at` | datetime | 软删除时间 |

索引：`uk_phone`、`idx_status`

### 2.2 用户地址表 (`user_address`)

用户收货地址。

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | bigint unsigned | 主键 |
| `user_id` | bigint unsigned | 用户 ID |
| `tag` | varchar(20) | 标签：HOME, OFFICE, SCHOOL |
| `receiver_name` | varchar(64) | 收货人姓名 |
| `receiver_phone` | varchar(20) | 收货人电话 |
| `is_default` | tinyint(1) | 是否默认地址 |
| `info` | json | 地址信息：{province, city, district, detail, postal_code} |
| `created_at` | datetime | 创建时间 |
| `updated_at` | datetime | 更新时间 |

索引：`idx_user`

### 2.3 用户风控记录表 (`user_punish_log`)

记录用户封禁/限制操作。

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | bigint unsigned | 主键 |
| `user_id` | bigint unsigned | 用户 ID |
| `action_type` | varchar(32) | BAN_LOGIN, BAN_ORDER |
| `reason` | varchar(255) | 原因 |
| `operator` | varchar(64) | 操作人（SYSTEM/管理员） |
| `start_time` | datetime | 开始时间 |
| `end_time` | datetime | 结束时间 |
| `created_at` | datetime | 创建时间 |

索引：`idx_user_id`

### 2.4 用户登录日志表 (`user_login_log`)

记录用户登录行为。

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | bigint unsigned | 主键 |
| `user_id` | bigint unsigned | 用户 ID |
| `login_type` | varchar(32) | password, sms, wechat |
| `login_ip` | varchar(64) | 登录 IP |
| `user_agent` | varchar(500) | UA 信息 |
| `status` | tinyint | 1:成功 2:失败 |
| `fail_reason` | varchar(255) | 失败原因 |
| `created_at` | datetime | 创建时间 |

索引：`idx_user_id`、`idx_created_at`

---

## 3. 商品服务 (`shop_product`)

### 3.1 商品分类表 (`category`)

多级商品分类。

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | bigint unsigned | 主键 |
| `parent_id` | bigint unsigned | 父分类 ID，0 为顶级 |
| `name` | varchar(64) | 分类名称 |
| `icon` | varchar(500) | 图标 URL |
| `sort` | int | 排序值 |
| `level` | tinyint | 层级（1,2,3） |
| `status` | tinyint | 1:启用 2:禁用 |
| `created_at` | datetime | 创建时间 |
| `updated_at` | datetime | 更新时间 |

索引：`idx_parent_id`

### 3.2 商品 SPU 表 (`product_spu`)

标准产品单元（同款商品）。

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | bigint unsigned | 主键 |
| `category_id` | bigint unsigned | 分类 ID |
| `name` | varchar(128) | 商品名称 |
| `brand` | varchar(64) | 品牌 |
| `desc` | varchar(255) | 简短描述 |
| `content` | longtext | 商品详情（富文本） |
| `main_pic` | varchar(500) | 主图 URL |
| `sub_pics` | json | 副图列表 |
| `video_url` | varchar(500) | 视频 URL |
| `status` | tinyint | 1:上架 2:下架 |
| `sales_count` | int | 实际销量 |
| `virtual_sales` | int | 虚拟销量 |
| `weight` | decimal(10,2) | 重量（kg） |
| `freight_template_id` | bigint unsigned | 运费模板 ID |
| `created_at` | datetime | 创建时间 |
| `updated_at` | datetime | 更新时间 |

索引：`idx_category_id`、`idx_status`

### 3.3 商品 SKU 表 (`product_sku`)

库存量单元（具体规格商品）。

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | bigint unsigned | 主键 |
| `spu_id` | bigint unsigned | SPU ID |
| `sku_code` | varchar(64) | SKU 编码（唯一） |
| `price` | decimal(10,2) | 销售价 |
| `market_price` | decimal(10,2) | 市场价 |
| `cost_price` | decimal(10,2) | 成本价 |
| `stock` | int | 库存数量 |
| `version` | int unsigned | 乐观锁版本号 |
| `spec_data` | json | 规格数据（如颜色、尺寸） |
| `images` | json | 规格图片列表 |
| `weight` | decimal(10,2) | 重量（kg） |
| `status` | tinyint | 1:启用 2:禁用 |
| `created_at` | datetime | 创建时间 |
| `updated_at` | datetime | 更新时间 |

索引：`uk_sku_code`、`idx_spu_id`、`idx_price`

### 3.4 库存变动流水表 (`stock_log`)

记录库存变更的详细流水。

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | bigint unsigned | 主键 |
| `sku_id` | bigint unsigned | SKU ID |
| `order_sn` | varchar(64) | 订单号 |
| `change_type` | tinyint | 1:预扣 2:回滚 3:实际扣减 4:退货加回 |
| `quantity` | int | 变动数量 |
| `before_stock` | int | 变动前库存 |
| `after_stock` | int | 变动后库存 |
| `remark` | varchar(255) | 备注 |
| `created_at` | datetime | 创建时间 |

索引：`uk_order_sku_type` (order_sn, sku_id, change_type)、`idx_sku_id`

### 3.5 运费模板表 (`freight_template`)

运费计算规则。

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | bigint unsigned | 主键 |
| `name` | varchar(64) | 模板名称 |
| `type` | tinyint | 1:按件数 2:按重量 |
| `default_fee` | decimal(10,2) | 默认运费 |
| `default_quantity` | int | 默认数量/重量 |
| `extra_fee` | decimal(10,2) | 超出部分每单位运费 |
| `free_threshold_amount` | decimal(10,2) | 包邮金额门槛 |
| `free_threshold_quantity` | int | 包邮数量/重量门槛 |
| `is_default` | tinyint | 是否为默认模板 |
| `status` | tinyint | 状态 |
| `created_at` | datetime | 创建时间 |
| `updated_at` | datetime | 更新时间 |

---

## 4. 订单服务 (`shop_order`)

### 4.1 购物车表 (`cart`)

用户购物车商品。

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | bigint unsigned | 主键 |
| `user_id` | bigint unsigned | 用户 ID |
| `sku_id` | bigint unsigned | SKU ID |
| `count` | int unsigned | 数量 |
| `selected` | tinyint | 是否勾选 |
| `created_at` | datetime | 创建时间 |
| `updated_at` | datetime | 更新时间 |

索引：`uk_user_sku` (user_id, sku_id)

### 4.2 订单主表 (`order_info`)

订单核心信息。

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | bigint unsigned | 主键 |
| `order_sn` | varchar(64) | 订单号（唯一） |
| `order_type` | tinyint | 1:普通订单 2:秒杀订单 3:拼团订单 |
| `idempotency_key` | varchar(64) | 幂等键（唯一） |
| `user_id` | bigint unsigned | 用户 ID |
| `status` | tinyint | 10:待支付 20:已支付 30:已发货 40:已完成 50:已取消 60:售后中 |
| `address_snap` | json | 地址快照 |
| `total_amount` | decimal(10,2) | 商品总金额 |
| `freight_amount` | decimal(10,2) | 运费 |
| `coupon_id` | bigint unsigned | 优惠券 ID |
| `coupon_discount` | decimal(10,2) | 优惠券抵扣金额 |
| `pay_amount` | decimal(10,2) | 实付金额 |
| `remark` | varchar(500) | 订单备注 |
| `pay_time` | datetime | 支付时间 |
| `delivery_time` | datetime | 发货时间 |
| `receive_time` | datetime | 收货时间 |
| `cancel_time` | datetime | 取消时间 |
| `seckill_activity_id` | bigint unsigned | 秒杀活动 ID（普通订单为0） |
| `seckill_price` | decimal(10,2) | 秒杀价格快照 |
| `created_at` | datetime | 创建时间 |
| `updated_at` | datetime | 更新时间 |

索引：`uk_order_sn`、`uk_idempotency_key`、`idx_user_id`、`idx_status`、`idx_order_type`、`idx_seckill_activity_id`、`idx_created_at`

### 4.3 订单商品表 (`order_item`)

订单中的商品快照。

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | bigint unsigned | 主键 |
| `order_id` | bigint unsigned | 订单主表 ID |
| `order_sn` | varchar(64) | 订单号 |
| `sku_id` | bigint unsigned | SKU ID |
| `spu_id` | bigint unsigned | SPU ID |
| `sku_name` | varchar(128) | 商品名称快照 |
| `sku_pic` | varchar(500) | 商品图片快照 |
| `spec_data` | json | 规格快照 |
| `price` | decimal(10,2) | 成交单价 |
| `count` | int | 数量 |
| `total_amount` | decimal(10,2) | 总金额 |
| `is_seckill` | tinyint | 是否秒杀商品 |
| `created_at` | datetime | 创建时间 |

索引：`idx_order_id`、`idx_order_sn`

---

## 5. 支付服务 (`shop_payment`)

### 5.1 支付记录表 (`payment`)

支付交易记录。

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | bigint unsigned | 主键 |
| `payment_sn` | varchar(64) | 支付单号（唯一） |
| `order_sn` | varchar(64) | 订单号（唯一） |
| `user_id` | bigint unsigned | 用户 ID |
| `amount` | decimal(10,2) | 支付金额 |
| `channel` | varchar(32) | wechat, alipay |
| `channel_order_sn` | varchar(128) | 渠道订单号 |
| `transaction_id` | varchar(128) | 渠道交易流水号 |
| `status` | tinyint | 0:待支付 1:成功 2:失败 3:退款中 4:退款成功 |
| `pay_time` | datetime | 支付成功时间 |
| `callback_data` | json | 回调原始数据 |
| `error_msg` | varchar(255) | 错误信息 |
| `created_at` | datetime | 创建时间 |
| `updated_at` | datetime | 更新时间 |

索引：`uk_payment_sn`、`uk_order_sn`、`idx_user_id`、`idx_status`

### 5.2 支付回调日志表 (`payment_callback_log`)

记录支付渠道回调的详细日志。

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | bigint unsigned | 主键 |
| `payment_sn` | varchar(64) | 支付单号 |
| `order_sn` | varchar(64) | 订单号 |
| `channel` | varchar(32) | 渠道 |
| `request_body` | text | 回调请求体 |
| `response_body` | text | 响应内容 |
| `status` | tinyint | 0:处理中 1:成功 2:失败 |
| `error_msg` | varchar(500) | 错误信息 |
| `created_at` | datetime | 创建时间 |

索引：`idx_payment_sn`、`idx_order_sn`

### 5.3 退款记录表 (`refund`)

退款申请与记录。

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | bigint unsigned | 主键 |
| `refund_sn` | varchar(64) | 退款单号（唯一） |
| `payment_sn` | varchar(64) | 支付单号 |
| `order_sn` | varchar(64) | 订单号 |
| `user_id` | bigint unsigned | 用户 ID |
| `amount` | decimal(10,2) | 退款金额 |
| `reason` | varchar(255) | 退款原因 |
| `status` | tinyint | 0:处理中 1:成功 2:失败 |
| `transaction_id` | varchar(128) | 渠道退款流水号 |
| `refund_time` | datetime | 退款成功时间 |
| `error_msg` | varchar(255) | 错误信息 |
| `created_at` | datetime | 创建时间 |
| `updated_at` | datetime | 更新时间 |

索引：`uk_refund_sn`、`idx_payment_sn`、`idx_order_sn`

---

## 6. 营销服务 (`shop_marketing`)

### 6.1 优惠券表 (`coupon`)

优惠券定义。

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | bigint unsigned | 主键 |
| `name` | varchar(128) | 优惠券名称 |
| `type` | tinyint | 1:满减券 2:折扣券 3:无门槛券 |
| `threshold_amount` | decimal(10,2) | 使用门槛金额 |
| `reduce_amount` | decimal(10,2) | 满减金额 |
| `discount_rate` | decimal(3,2) | 折扣率（如 0.85） |
| `max_discount_amount` | decimal(10,2) | 最大抵扣金额（折扣券） |
| `total_quantity` | int | 发行总量 |
| `used_quantity` | int | 已使用数量 |
| `per_user_limit` | int | 每人限领数量 |
| `start_time` | datetime | 生效时间 |
| `end_time` | datetime | 失效时间 |
| `status` | tinyint | 1:启用 2:禁用 |
| `description` | varchar(255) | 描述 |
| `created_at` | datetime | 创建时间 |
| `updated_at` | datetime | 更新时间 |

索引：`idx_time`

### 6.2 用户优惠券表 (`user_coupon`)

用户领取的优惠券。

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | bigint unsigned | 主键 |
| `coupon_id` | bigint unsigned | 优惠券 ID |
| `user_id` | bigint unsigned | 用户 ID |
| `order_sn` | varchar(64) | 使用订单号 |
| `status` | tinyint | 0:未使用 1:已使用 2:已过期 |
| `used_time` | datetime | 使用时间 |
| `source` | varchar(32) | 来源（ADMIN/ACTIVITY） |
| `created_at` | datetime | 领取时间 |
| `expire_time` | datetime | 过期时间 |

索引：`uk_user_coupon` (user_id, coupon_id)、`idx_coupon_id`、`idx_user_id`

### 6.3 秒杀活动表 (`seckill_activity`)

秒杀活动配置。

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | bigint unsigned | 主键 |
| `name` | varchar(128) | 活动名称 |
| `sku_id` | bigint unsigned | 秒杀商品 SKU ID |
| `sku_name` | varchar(128) | 商品名称快照 |
| `sku_pic` | varchar(500) | 商品图片快照 |
| `seckill_price` | decimal(10,2) | 秒杀价 |
| `stock` | int | 活动库存 |
| `sold_count` | int | 已售数量 |
| `per_user_limit` | int | 每人限购数量 |
| `start_time` | datetime | 开始时间 |
| `end_time` | datetime | 结束时间 |
| `status` | tinyint | 1:进行中 2:已结束 3:已取消 |
| `version` | int unsigned | 乐观锁版本号 |
| `created_at` | datetime | 创建时间 |
| `updated_at` | datetime | 更新时间 |

索引：`idx_sku_id`、`idx_time`

### 6.4 秒杀预占记录表 (`seckill_preorder`)

秒杀成功时的预占记录，支付成功后生成正式订单。

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | bigint unsigned | 主键 |
| `preorder_sn` | varchar(64) | 预占单号（唯一） |
| `activity_id` | bigint unsigned | 秒杀活动 ID |
| `sku_id` | bigint unsigned | SKU ID |
| `user_id` | bigint unsigned | 用户 ID |
| `seckill_price` | decimal(10,2) | 秒杀价格 |
| `quantity` | int | 数量 |
| `order_sn` | varchar(64) | 支付成功后关联的正式订单号 |
| `status` | tinyint | 0:预占成功 1:已支付 2:已取消 3:已超时 |
| `expire_time` | datetime | 预占过期时间（通常 5-15 分钟） |
| `created_at` | datetime | 创建时间 |
| `updated_at` | datetime | 更新时间 |

索引：`uk_preorder_sn`、`uk_user_activity` (user_id, activity_id)、`idx_activity_id`、`idx_status`、`idx_expire_time`

---

## 7. 后台管理服务 (`shop_admin`)

### 7.1 管理员表 (`admin`)

后台账号。

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | bigint unsigned | 主键 |
| `username` | varchar(64) | 用户名（唯一） |
| `password` | varchar(255) | 加密密码 |
| `real_name` | varchar(64) | 真实姓名 |
| `avatar` | varchar(500) | 头像 |
| `role_id` | bigint unsigned | 角色 ID |
| `status` | tinyint | 1:启用 2:禁用 |
| `last_login_time` | datetime | 最后登录时间 |
| `last_login_ip` | varchar(64) | 最后登录 IP |
| `created_at` | datetime | 创建时间 |
| `updated_at` | datetime | 更新时间 |

索引：`uk_username`

### 7.2 角色表 (`role`)

权限角色。

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | bigint unsigned | 主键 |
| `name` | varchar(64) | 角色名称 |
| `permissions` | json | 权限列表 |
| `remark` | varchar(255) | 备注 |
| `created_at` | datetime | 创建时间 |
| `updated_at` | datetime | 更新时间 |

### 7.3 操作日志表 (`admin_log`)

管理员操作记录。

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | bigint unsigned | 主键 |
| `admin_id` | bigint unsigned | 管理员 ID |
| `username` | varchar(64) | 管理员用户名 |
| `module` | varchar(64) | 操作模块 |
| `action` | varchar(64) | 操作类型 |
| `request_method` | varchar(10) | 请求方法 |
| `request_url` | varchar(255) | 请求 URL |
| `request_params` | text | 请求参数 |
| `response_result` | text | 响应结果 |
| `ip` | varchar(64) | IP 地址 |
| `duration_ms` | int | 耗时（毫秒） |
| `created_at` | datetime | 创建时间 |

索引：`idx_admin_id`、`idx_created_at`

### 7.4 轮播图表 (`banner`)

首页轮播图配置。

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | bigint unsigned | 主键 |
| `title` | varchar(128) | 标题 |
| `image_url` | varchar(500) | 图片 URL |
| `link_url` | varchar(500) | 跳转链接 |
| `type` | tinyint | 1:商品 2:分类 3:活动页 |
| `target_id` | bigint unsigned | 关联的目标 ID |
| `sort` | int | 排序值 |
| `platform` | varchar(20) | ALL, PC, H5 |
| `status` | tinyint | 1:启用 2:禁用 |
| `start_time` | datetime | 生效开始时间 |
| `end_time` | datetime | 生效结束时间 |
| `created_at` | datetime | 创建时间 |
| `updated_at` | datetime | 更新时间 |

索引：`idx_status`

### 7.5 公告表 (`notice`)

系统公告。

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | bigint unsigned | 主键 |
| `title` | varchar(128) | 标题 |
| `content` | text | 内容 |
| `type` | tinyint | 1:系统公告 2:活动公告 |
| `priority` | tinyint | 1:普通 2:重要 3:置顶 |
| `status` | tinyint | 1:发布 2:草稿 3:已下线 |
| `publish_time` | datetime | 发布时间 |
| `created_by` | varchar(64) | 创建人 |
| `created_at` | datetime | 创建时间 |
| `updated_at` | datetime | 更新时间 |

索引：`idx_status`

### 7.6 帮助中心分类表 (`help_category`)

帮助文档分类。

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | bigint unsigned | 主键 |
| `name` | varchar(64) | 分类名称 |
| `icon` | varchar(255) | 图标 |
| `sort` | int | 排序值 |
| `status` | tinyint | 状态 |
| `created_at` | datetime | 创建时间 |

### 7.7 帮助中心文章表 (`help_article`)

帮助文档文章。

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | bigint unsigned | 主键 |
| `category_id` | bigint unsigned | 分类 ID |
| `title` | varchar(128) | 标题 |
| `content` | longtext | 内容 |
| `keywords` | varchar(255) | SEO 关键词 |
| `view_count` | int | 浏览次数 |
| `sort` | int | 排序值 |
| `status` | tinyint | 1:发布 2:草稿 |
| `created_at` | datetime | 创建时间 |
| `updated_at` | datetime | 更新时间 |

索引：`idx_category_id`

---

## 8. 搜索服务 (`shop_search`)

> 注：实际商品搜索功能建议使用 Elasticsearch，此数据库仅存储辅助数据。

### 8.1 用户搜索历史表 (`search_history`)

记录用户的搜索关键词。

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | bigint unsigned | 主键 |
| `user_id` | bigint unsigned | 用户 ID |
| `keyword` | varchar(128) | 搜索关键词 |
| `result_count` | int | 搜索结果数量 |
| `created_at` | datetime | 创建时间 |

索引：`idx_user_id`、`idx_created_at`

### 8.2 热搜词表 (`hot_keyword`)

统计热门搜索词。

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | bigint unsigned | 主键 |
| `keyword` | varchar(128) | 关键词（唯一） |
| `search_count` | bigint | 搜索次数 |
| `sort` | int | 排序值 |
| `status` | tinyint | 1:显示 2:隐藏 |
| `updated_at` | datetime | 更新时间 |

索引：`uk_keyword`、`idx_search_count`

### 8.3 ES 数据同步记录表 (`es_sync_log`)

记录需要同步到 Elasticsearch 的数据变更。

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | bigint unsigned | 主键 |
| `table_name` | varchar(64) | 表名（product_spu, product_sku） |
| `record_id` | bigint unsigned | 记录 ID |
| `action` | varchar(20) | INSERT, UPDATE, DELETE |
| `status` | tinyint | 0:待同步 1:已同步 2:同步失败 |
| `retry_count` | int | 重试次数 |
| `error_msg` | varchar(500) | 错误信息 |
| `created_at` | datetime | 创建时间 |
| `updated_at` | datetime | 更新时间 |

索引：`idx_status`、`idx_table_record` (table_name, record_id)

---

## 9. 跨库表关系说明

| 关系 | 源表（库） | 目标表（库） | 说明 |
|------|------------|--------------|------|
| 用户地址 | `user_address.user_id` | `user.id` | 外键逻辑关联 |
| 购物车 | `cart.user_id` | `user.id` | 外键逻辑关联 |
| 购物车 | `cart.sku_id` | `product_sku.id` | 外键逻辑关联 |
| 订单 | `order_info.user_id` | `user.id` | 外键逻辑关联 |
| 订单 | `order_info.address_snap` | `user_address` | 快照，无外键 |
| 订单商品 | `order_item.sku_id` | `product_sku.id` | 外键逻辑关联 |
| 支付 | `payment.order_sn` | `order_info.order_sn` | 外键逻辑关联 |
| 退款 | `refund.payment_sn` | `payment.payment_sn` | 外键逻辑关联 |
| 秒杀预占 | `seckill_preorder.activity_id` | `seckill_activity.id` | 外键逻辑关联 |
| 秒杀预占 | `seckill_preorder.order_sn` | `order_info.order_sn` | 支付成功后关联 |
| 用户优惠券 | `user_coupon.coupon_id` | `coupon.id` | 外键逻辑关联 |
| 管理员 | `admin.role_id` | `role.id` | 外键逻辑关联 |
| 帮助文章 | `help_article.category_id` | `help_category.id` | 外键逻辑关联 |

> 由于微服务架构，各数据库之间不建立物理外键，通过应用层保证数据一致性。

---

## 10. 版本历史

| 日期 | 版本 | 说明 |
|------|------|------|
| 2026-04-06 | 1.0 | 初始版本，包含用户、商品、订单、支付、营销、后台、搜索七大服务数据库表结构 |