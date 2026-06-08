# 电商项目数据库表结构设计文档

## 1. 总体说明

本项目为电商平台，采用微服务架构，数据库按业务域拆分为 7 个独立数据库：

| 数据库 | 服务 | 说明 |
|--------|------|------|
| `shop_user` | 用户服务 | 用户注册登录、个人信息、地址管理、风控 |
| `shop_product` | 商品服务 | 商品分类、SPU、SKU、库存管理、运费模板 |
| `shop_order` | 订单服务 | 购物车、订单、订单商品快照、售后 |
| `shop_payment` | 支付服务 | 支付记录、支付回调、退款记录 |
| `shop_marketing` | 营销服务 | 优惠券、秒杀活动、秒杀预占 |
| `shop_admin` | 后台管理服务 | 管理员、角色权限、操作日志、CMS 内容 |
| `shop_search` | 搜索服务 | 搜索历史、热搜词、ES 同步记录 |

字符集统一使用 `utf8mb4`，排序规则 `utf8mb4_unicode_ci`。

**金额单位说明**：除特别注明外，所有金额字段单位为**分**（int64），即 `100` 表示 `1.00` 元。

---

## 2. 用户服务 (`shop_user`)

### 2.1 用户主表 (`user`)

存储用户基本信息。

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | bigint unsigned | 主键 |
| `phone` | varchar(20) | 手机号（唯一索引） |
| `password` | varchar(255) | 加密密码（bcrypt） |
| `nickname` | varchar(64) | 昵称 |
| `avatar` | varchar(500) | 头像 URL |
| `gender` | int64 | 性别：0-未知，1-男，2-女 |
| `birthday` | date | 生日（可空） |
| `status` | int64 | 账号状态：1-正常，2-限制下单，3-封禁 |
| `last_login_time` | datetime | 最后登录时间（可空） |
| `last_login_ip` | varchar(64) | 最后登录 IP |
| `ext_info` | json | 扩展信息（JSON格式，可空） |
| `created_at` | datetime | 创建时间 |
| `updated_at` | datetime | 更新时间 |
| `deleted_at` | datetime | 软删除时间（可空） |

索引：`uk_phone`（唯一索引）、`idx_status`、`idx_nickname`、`idx_created_at`

### 2.2 用户地址表 (`user_address`)

用户收货地址。

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | bigint unsigned | 主键 |
| `user_id` | bigint unsigned | 用户 ID |
| `tag` | varchar(20) | 地址标签：`HOME`-家，`OFFICE`-公司，`SCHOOL`-学校 |
| `receiver_name` | varchar(64) | 收货人姓名 |
| `receiver_phone` | varchar(20) | 收货人电话 |
| `is_default` | int64 | 是否默认地址：0-否，1-是 |
| `info` | json | 地址详细信息：{province, city, district, detail, postal_code} |
| `created_at` | datetime | 创建时间 |
| `updated_at` | datetime | 更新时间 |

索引：`idx_user`、`idx_user_default`（user_id, is_default）

### 2.3 用户风控记录表 (`user_punish_log`)

记录用户封禁/限制操作。

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | bigint unsigned | 主键 |
| `user_id` | bigint unsigned | 用户 ID |
| `phone` | varchar(20) | 手机号码 |
| `action_type` | int64 | 处罚类型：1-禁止下单，2-禁止登录 |
| `reason` | varchar(255) | 处罚原因 |
| `banned_by` | varchar(64) | 操作人：`SYSTEM`-系统自动，`ADMIN`-管理员（记录管理员账号） |
| `unbanned_by` | varchar(64) | 解除操作人（可空） |
| `start_time` | datetime | 处罚开始时间 |
| `end_time` | datetime | 处罚结束时间（可空） |
| `created_at` | datetime | 创建时间 |

索引：`uk_phone`（唯一索引）、`idx_user_id`

### 2.4 用户登录日志表 (`user_login_log`)

记录用户登录行为。

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | bigint unsigned | 主键 |
| `user_id` | bigint unsigned | 用户 ID（未登录为0） |
| `login_type` | varchar(32) | 登录方式：`password`-密码登录，`captcha`-短信登录 |
| `login_ip` | varchar(64) | 登录 IP |
| `user_agent` | varchar(500) | UA 信息 |
| `status` | int64 | 登录状态：1-成功，2-失败 |
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
| `parent_id` | bigint unsigned | 父分类 ID：0-顶级分类 |
| `name` | varchar(64) | 分类名称 |
| `icon` | varchar(500) | 分类图标 URL |
| `sort` | int64 | 排序值：数值越小越靠前 |
| `level` | int64 | 分类层级：1-一级分类，2-二级分类，3-三级分类 |
| `status` | int64 | 状态：1-启用，2-禁用 |
| `created_at` | datetime | 创建时间 |
| `updated_at` | datetime | 更新时间 |
| `delete_at` | datetime | 删除时间（可空，软删除） |

索引：`idx_parent_id`

### 3.2 商品 SPU 表 (`product_spu`)

标准产品单元（同款商品）。

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | bigint unsigned | 主键 |
| `category_id` | bigint unsigned | 所属分类 ID |
| `name` | varchar(128) | 商品名称 |
| `brand` | varchar(64) | 品牌名称 |
| `desc` | varchar(255) | 商品简短描述 |
| `content` | text | 商品详情内容（HTML，可空） |
| `main_pic` | varchar(500) | 主图 URL |
| `sub_pics` | text | 副图列表（JSON数组，可空） |
| `video_url` | varchar(500) | 商品视频 URL |
| `status` | int64 | 上架状态：1-上架，2-下架 |
| `sales_count` | int64 | 实际销量 |
| `virtual_sales` | int64 | 虚拟销量（用于展示） |
| `freight_template_id` | bigint unsigned | 运费模板 ID：0-使用默认模板 |
| `created_at` | datetime | 创建时间 |
| `updated_at` | datetime | 更新时间 |
| `delete_at` | datetime | 删除时间（可空，软删除） |

索引：`idx_category_id`、`idx_status`

### 3.3 商品 SKU 表 (`product_sku`)

库存量单元（具体规格商品）。

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | bigint unsigned | 主键 |
| `spu_id` | bigint unsigned | 所属 SPU ID |
| `spu_name` | varchar(128) | SPU名称（冗余字段，便于查询） |
| `sku_code` | varchar(64) | SKU编码（唯一标识） |
| `price` | int64 | 销售价格（单位：分） |
| `market_price` | int64 | 市场价/划线价（单位：分） |
| `cost_price` | int64 | 成本价（单位：分） |
| `stock` | int64 | 库存数量 |
| `locked_stock` | int64 | 锁定库存数量（秒杀/预占占用的库存） |
| `version` | bigint unsigned | 乐观锁版本号（用于并发控制） |
| `spec_data` | text | 规格数据（JSON格式），如 `{"颜色":"红色","尺寸":"XL"}` |
| `images` | text | SKU专属图片列表（JSON数组，可空） |
| `weight` | int64 | SKU重量（单位：克） |
| `status` | int64 | 状态：1-启用，2-禁用 |
| `created_at` | datetime | 创建时间 |
| `updated_at` | datetime | 更新时间 |
| `delete_at` | datetime | 删除时间（可空，软删除） |

索引：`uk_sku_code`（唯一索引）、`idx_spu_id`、`idx_status`、`idx_price`

### 3.4 库存变动流水表 (`stock_log`)

记录库存变更的详细流水。

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | bigint unsigned | 主键 |
| `sku_id` | bigint unsigned | SKU ID |
| `order_sn` | varchar(64) | 关联订单号（可空） |
| `change_type` | int | 变动类型：1-预扣（锁定），2-回滚（释放），3-实际扣减，4-退货加回 |
| `quantity` | int | 变动数量（正数增加，负数减少） |
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
| `type` | int64 | 计费方式：1-按件数，2-按重量 |
| `default_fee` | int64 | 基础运费（单位：分） |
| `default_quantity` | int64 | 基础数量（首件/首重） |
| `extra_fee` | int64 | 续件/续重费用（单位：分） |
| `free_threshold_amount` | int64 | 满额包邮阈值（单位：分，0表示不包邮） |
| `free_threshold_quantity` | int64 | 满件包邮阈值 |
| `is_default` | int64 | 是否默认模板：0-否，1-是 |
| `status` | int64 | 状态：1-启用，2-禁用 |
| `created_at` | datetime | 创建时间 |
| `updated_at` | datetime | 更新时间 |
| `deleted_at` | datetime | 删除时间（可空） |

索引：`idx_default`

## 4. 订单服务 (`shop_order`)

### 4.1 购物车表 (`cart`)

用户购物车商品。

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | bigint unsigned | 主键 |
| `user_id` | bigint unsigned | 用户 ID |
| `sku_id` | bigint unsigned | SKU ID |
| `count` | bigint unsigned | 商品数量 |
| `selected` | bool | 是否勾选：false-未选中，true-已选中 |
| `created_at` | datetime | 创建时间 |
| `updated_at` | datetime | 更新时间 |

索引：`uk_user_sku` (user_id, sku_id)（唯一索引）

### 4.2 订单主表 (`order_info`)

订单核心信息。

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | bigint unsigned | 主键 |
| `order_sn` | varchar(64) | 订单号（唯一标识） |
| `order_type` | int64 | 订单类型：1-普通订单，2-秒杀订单，3-拼团订单 |
| `idempotency_key` | varchar(64) | 幂等键（同一用户下唯一，防止重复提交） |
| `user_id` | bigint unsigned | 用户 ID |
| `status` | int64 | 订单状态：10-待支付，20-已支付，30-已发货，40-已完成，50-已取消，60-售后中 |
| `address_snap` | text | 收货地址快照（JSON格式） |
| `total_amount` | int64 | 商品总金额（单位：分） |
| `freight_amount` | int64 | 运费金额（单位：分） |
| `coupon_id` | bigint unsigned | 优惠券 ID：0-未使用优惠券 |
| `coupon_discount` | int64 | 优惠券优惠金额（单位：分） |
| `pay_amount` | int64 | 实付金额（单位：分） |
| `remark` | varchar(500) | 订单备注（用户留言） |
| `pay_time` | datetime | 支付时间（可空） |
| `pay_type` | int64 | 支付类型：1-微信支付，2-支付宝支付，3-其他支付（可空） |
| `delivery_time` | datetime | 发货时间（可空） |
| `delivery_sn` | varchar(64) | 物流单号（可空） |
| `delivery_corp` | varchar(64) | 物流公司名称（可空） |
| `receive_time` | datetime | 确认收货时间（可空） |
| `cancel_time` | datetime | 取消时间（可空） |
| `cancel_reason_type` | int64 | 取消原因类型：0-未取消，1-用户主动取消，2-超时未支付取消，3-库存不足，4-管理员取消，5-其他 |
| `cancel_reason` | varchar(255) | 取消原因详情（用户填写/系统备注） |
| `seckill_activity_id` | bigint unsigned | 秒杀活动 ID：普通订单为0 |
| `seckill_price` | int64 | 秒杀价格快照（单位：分，普通订单为0） |
| `created_at` | datetime | 创建时间 |
| `expire_time` | datetime | 订单支付过期时间（可空） |
| `updated_at` | datetime | 更新时间 |

索引：
- `uk_order_sn`（唯一索引）
- `uk_user_idempotency`（user_id, idempotency_key）（唯一索引）
- `idx_user_id`
- `idx_status`
- `idx_order_type`
- `idx_seckill_activity_id`
- `idx_user_created`（user_id, created_at）
- `idx_status_created`（status, created_at）
- `idx_expire_status`（expire_time, status）
- `idx_user_status`（user_id, status）

### 4.3 订单商品表 (`order_item`)

订单中的商品快照。

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | bigint unsigned | 主键 |
| `order_id` | bigint unsigned | 订单主表 ID |
| `order_sn` | varchar(64) | 订单号 |
| `sku_id` | bigint unsigned | SKU ID |
| `spu_id` | bigint unsigned | SPU ID |
| `spu_name` | varchar(128) | SPU名称快照 |
| `sku_name` | varchar(128) | SKU名称快照 |
| `sku_pic` | varchar(500) | SKU图片快照 |
| `price` | int64 | 成交单价（单位：分） |
| `count` | int64 | 购买数量 |
| `total_amount` | int64 | 小计金额（单价×数量，单位：分） |
| `is_seckill` | int64 | 是否秒杀商品：0-否，1-是 |
| `created_at` | datetime | 创建时间 |

索引：`idx_order_id`、`idx_order_sn`

### 4.4 售后表 (`after_sale`)

售后单信息。

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | bigint unsigned | 主键 |
| `after_sale_sn` | varchar(64) | 售后单号（唯一标识） |
| `idempotency_key` | varchar(64) | 幂等键（同一用户下唯一） |
| `order_id` | bigint unsigned | 关联订单 ID |
| `order_sn` | varchar(64) | 关联订单号 |
| `user_id` | bigint unsigned | 用户 ID |
| `type` | int64 | 售后类型：1-仅退款，2-退货退款 |
| `status` | int64 | 售后状态：10-待审核，20-待商家收货，30-退款中，40-已完成，50-已拒绝，60-已取消 |
| `refund_status` | int64 | 退款状态：0-未退款，10-退款中，20-退款成功，30-退款失败 |
| `apply_amount` | int64 | 申请退款金额（单位：分） |
| `approved_amount` | int64 | 审核通过金额（单位：分） |
| `real_refund_amount` | int64 | 实际退款金额（单位：分） |
| `reason` | varchar(500) | 售后原因 |
| `audit_remark` | varchar(500) | 审核备注（可空） |
| `audit_time` | datetime | 审核时间（可空） |
| `images` | text | 凭证图片 URL 数组（JSON格式，可空） |
| `return_tracking_sn` | varchar(64) | 退货物流单号（可空） |
| `return_tracking_corp` | varchar(64) | 退货物流公司（可空） |
| `return_received_time` | datetime | 商家收货时间（可空） |
| `refund_sn` | varchar(64) | 商户退款单号（可空） |
| `refund_time` | datetime | 退款时间（可空） |
| `third_refund_sn` | varchar(128) | 第三方退款单号（可空） |
| `cancel_time` | datetime | 取消时间（可空） |
| `created_at` | datetime | 创建时间 |
| `updated_at` | datetime | 更新时间 |

索引：
- `uk_after_sale_sn`（唯一索引）
- `uk_refund_sn`（唯一索引）
- `idx_order_sn`
- `idx_user_id`
- `idx_status`

### 4.5 售后商品表 (`after_sale_item`)

售后商品快照。

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | bigint unsigned | 主键 |
| `after_sale_sn` | varchar(64) | 售后单号 |
| `order_item_id` | bigint unsigned | 原订单商品项 ID |
| `sku_id` | bigint unsigned | SKU ID |
| `spu_id` | bigint unsigned | SPU ID |
| `product_name` | varchar(128) | SPU名称快照 |
| `sku_name` | varchar(128) | SKU名称快照 |
| `sku_pic` | varchar(500) | SKU图片快照 |
| `price` | int64 | 原成交单价（单位：分） |
| `quantity` | int64 | 售后数量 |
| `total_amount` | int64 | 原小计金额（单价×数量，单位：分） |
| `refund_amount` | int64 | 该项退款金额（单位：分） |
| `created_at` | datetime | 创建时间 |

索引：`idx_after_sale_sn`、`idx_order_item_id`

---

## 5. 支付服务 (`shop_payment`)

### 5.1 支付记录表 (`payment`)

支付交易记录。

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | bigint unsigned | 主键 |
| `payment_sn` | varchar(64) | 支付流水号（唯一标识） |
| `order_sn` | varchar(64) | 订单号 |
| `user_id` | bigint unsigned | 用户 ID |
| `amount` | int64 | 支付金额（单位：分） |
| `channel` | varchar(32) | 支付渠道：`wechat`-微信支付，`alipay`-支付宝 |
| `channel_order_sn` | varchar(128) | 第三方支付平台订单号（可空） |
| `transaction_id` | varchar(128) | 第三方支付交易流水号（可空） |
| `status` | int64 | 支付状态：0-待支付，1-支付成功，2-支付失败，3-退款中，4-退款成功 |
| `pay_time` | datetime | 支付成功时间（可空） |
| `callback_data` | text | 支付回调数据快照（JSON格式，可空） |
| `error_msg` | varchar(255) | 错误信息（可空） |
| `created_at` | datetime | 创建时间 |
| `updated_at` | datetime | 更新时间 |

索引：
- `uk_payment_sn`（唯一索引）
- `uk_order_sn`（唯一索引）
- `idx_user_id`
- `idx_status`

### 5.2 支付回调日志表 (`payment_callback_log`)

记录支付渠道回调的详细日志。

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | bigint unsigned | 主键 |
| `payment_sn` | varchar(64) | 支付流水号 |
| `order_sn` | varchar(64) | 订单号 |
| `channel` | varchar(32) | 支付渠道 |
| `request_body` | text | 回调请求体 |
| `response_body` | text | 响应内容 |
| `status` | int64 | 处理状态：0-处理中，1-成功，2-失败 |
| `error_msg` | varchar(500) | 错误信息（可空） |
| `created_at` | datetime | 创建时间 |

索引：`idx_payment_sn`、`idx_order_sn`

### 5.3 退款记录表 (`refund`)

退款申请与记录。

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | bigint unsigned | 主键 |
| `refund_sn` | varchar(64) | 退款单号（唯一标识） |
| `payment_sn` | varchar(64) | 支付流水号 |
| `order_sn` | varchar(64) | 订单号 |
| `user_id` | bigint unsigned | 用户 ID |
| `amount` | int64 | 退款金额（单位：分） |
| `reason` | varchar(255) | 退款原因 |
| `status` | int64 | 退款状态：0-处理中，1-退款成功，2-退款失败 |
| `transaction_id` | varchar(128) | 第三方退款交易流水号（可空） |
| `refund_time` | datetime | 退款成功时间（可空） |
| `error_msg` | varchar(255) | 错误信息（可空） |
| `created_at` | datetime | 创建时间 |
| `updated_at` | datetime | 更新时间 |

索引：
- `uk_refund_sn`（唯一索引）
- `idx_payment_sn`
- `idx_order_sn`

---

## 6. 营销服务 (`shop_marketing`)

### 6.1 优惠券表 (`coupon`)

优惠券定义。

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | bigint unsigned | 主键 |
| `name` | varchar(128) | 优惠券名称 |
| `type` | int64 | 优惠券类型：1-满减券，2-折扣券，3-无门槛券 |
| `threshold_amount` | int64 | 使用门槛金额（单位：分），0表示无门槛 |
| `reduce_amount` | int64 | 满减金额（单位：分），仅满减券有效 |
| `discount_rate` | int64 | 折扣率（万比分比）：8000表示0.8折，仅折扣券有效 |
| `max_discount_amount` | int64 | 折扣封顶金额（单位：分），仅折扣券有效 |
| `total_quantity` | int64 | 发行总量 |
| `used_quantity` | int64 | 已使用数量 |
| `per_user_limit` | int64 | 每人限领数量，0表示不限制 |
| `start_time` | datetime | 生效开始时间 |
| `end_time` | datetime | 失效结束时间 |
| `status` | int64 | 状态：1-启用，2-禁用 |
| `description` | varchar(255) | 优惠券描述/使用说明 |
| `created_at` | datetime | 创建时间 |
| `updated_at` | datetime | 更新时间 |

索引：`idx_time`、`idx_status`

### 6.2 用户优惠券表 (`user_coupon`)

用户领取的优惠券。

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | bigint unsigned | 主键 |
| `coupon_id` | bigint unsigned | 优惠券 ID |
| `user_id` | bigint unsigned | 用户 ID |
| `order_sn` | varchar(64) | 使用该券的订单号（可空） |
| `status` | int64 | 状态：0-未使用，1-已使用，2-已过期 |
| `used_time` | datetime | 使用时间（可空） |
| `source` | varchar(32) | 来源：`ADMIN`-管理员发放，`ACTIVITY`-活动领取 |
| `created_at` | datetime | 领取时间 |
| `expire_time` | datetime | 过期时间 |

索引：
- `uk_user_coupon` (user_id, coupon_id)
- `idx_coupon_id`
- `idx_user_id`
- `idx_status`
- `idx_expire_time`

### 6.3 秒杀活动表 (`seckill_activity`)

秒杀活动配置。

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | bigint unsigned | 主键 |
| `name` | varchar(128) | 活动名称 |
| `sku_id` | bigint unsigned | 秒杀商品 SKU ID |
| `sku_name` | varchar(128) | 商品名称快照 |
| `sku_pic` | varchar(500) | 商品图片快照 |
| `seckill_price` | int64 | 秒杀价格（单位：分） |
| `stock` | int64 | 活动库存 |
| `sold_count` | int64 | 已售数量 |
| `per_user_limit` | int64 | 每人限购数量 |
| `start_time` | datetime | 开始时间 |
| `end_time` | datetime | 结束时间 |
| `status` | int64 | 状态：1-进行中，2-已结束，3-已取消 |
| `version` | bigint unsigned | 乐观锁版本号 |
| `created_at` | datetime | 创建时间 |
| `updated_at` | datetime | 更新时间 |

索引：`idx_sku_id`、`idx_time`

### 6.4 秒杀预占记录表 (`seckill_preorder`)

秒杀成功时的预占记录，支付成功后生成正式订单。

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | bigint unsigned | 主键 |
| `preorder_sn` | varchar(64) | 预占单号（唯一标识） |
| `activity_id` | bigint unsigned | 秒杀活动 ID |
| `sku_id` | bigint unsigned | SKU ID |
| `user_id` | bigint unsigned | 用户 ID |
| `seckill_price` | int64 | 秒杀价格（单位：分） |
| `quantity` | int64 | 预占数量 |
| `order_sn` | varchar(64) | 支付成功后关联的正式订单号（可空） |
| `status` | int64 | 状态：0-预占成功待支付，1-已支付，2-已取消，3-已超时 |
| `expire_time` | datetime | 预占过期时间（通常5-15分钟） |
| `created_at` | datetime | 创建时间 |
| `updated_at` | datetime | 更新时间 |

索引：
- `uk_preorder_sn`（唯一索引）
- `uk_user_activity` (user_id, activity_id)
- `idx_activity_id`
- `idx_status`
- `idx_expire_time`

---

## 7. 后台管理服务 (`shop_admin`)

### 7.1 管理员表 (`admin`)

后台账号。

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | bigint unsigned | 主键 |
| `username` | varchar(64) | 登录用户名（唯一） |
| `password` | varchar(255) | 登录密码（加密） |
| `real_name` | varchar(64) | 真实姓名 |
| `avatar` | varchar(500) | 头像 URL |
| `role_id` | bigint unsigned | 角色 ID：0-无角色 |
| `status` | int64 | 状态：1-启用，2-禁用 |
| `last_login_time` | datetime | 最后登录时间（可空） |
| `last_login_ip` | varchar(64) | 最后登录 IP |
| `created_at` | datetime | 创建时间 |
| `updated_at` | datetime | 更新时间 |

索引：`uk_username`（唯一索引）

### 7.2 角色表 (`role`)

权限角色。

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | bigint unsigned | 主键 |
| `name` | varchar(64) | 角色名称 |
| `code` | varchar(64) | 角色标识 |
| `remark` | varchar(255) | 备注说明 |
| `created_at` | datetime | 创建时间 |
| `updated_at` | datetime | 更新时间 |

### 7.3 权限表 (`permission`)

权限定义。

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | bigint unsigned | 主键 |
| `name` | varchar(64) | 权限名称 |
| `code` | varchar(128) | 权限标识（唯一） |
| `module` | varchar(64) | 所属模块 |
| `created_at` | datetime | 创建时间 |

索引：`uk_code`（唯一索引）

### 7.4 角色权限表 (`role_permission`)

角色与权限的关联关系。

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | bigint unsigned | 主键 |
| `role_id` | bigint unsigned | 角色 ID |
| `permission_id` | bigint unsigned | 权限 ID |

索引：`uk_role_permission` (role_id, permission_id)（唯一索引）

### 7.5 操作日志表 (`admin_log`)

管理员操作记录。

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | bigint unsigned | 主键 |
| `admin_id` | bigint unsigned | 管理员 ID |
| `username` | varchar(64) | 管理员用户名 |
| `module` | varchar(64) | 操作模块 |
| `action` | varchar(64) | 操作类型 |
| `request_method` | varchar(10) | 请求方法：GET/POST/PUT/DELETE |
| `request_url` | varchar(255) | 请求 URL |
| `request_params` | text | 请求参数（JSON格式） |
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
| `link_url` | varchar(500) | 跳转链接（可空） |
| `type` | int64 | 跳转类型：1-商品详情，2-分类页，3-活动页，4-H5页面 |
| `target_id` | bigint unsigned | 关联目标 ID（根据 type 类型） |
| `sort` | int64 | 排序值 |
| `platform` | varchar(20) | 展示平台：`ALL`-全部，`PC`-PC端，`H5`-移动端 |
| `status` | int64 | 状态：1-启用，2-禁用 |
| `start_time` | datetime | 生效开始时间 |
| `end_time` | datetime | 生效结束时间 |
| `created_at` | datetime | 创建时间 |
| `updated_at` | datetime | 更新时间 |

索引：`idx_status`

### 7.7 公告表 (`notice`)

系统公告。

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | bigint unsigned | 主键 |
| `title` | varchar(128) | 标题 |
| `content` | text | 内容 |
| `type` | int64 | 公告类型：1-系统公告，2-活动公告 |
| `priority` | int64 | 优先级：1-普通，2-重要，3-置顶 |
| `status` | int64 | 状态：1-发布，2-草稿，3-已下线 |
| `publish_time` | datetime | 发布时间（可空） |
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
| `sort` | int64 | 排序值 |
| `status` | int64 | 状态：1-启用，2-禁用 |
| `created_at` | datetime | 创建时间 |

### 7.7 帮助中心文章表 (`help_article`)

帮助文档文章。

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | bigint unsigned | 主键 |
| `category_id` | bigint unsigned | 分类 ID |
| `title` | varchar(128) | 标题 |
| `content` | longtext | 内容（HTML） |
| `keywords` | varchar(255) | SEO 关键词 |
| `view_count` | int64 | 浏览次数 |
| `sort` | int64 | 排序值 |
| `status` | int64 | 状态：1-发布，2-草稿 |
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
| `result_count` | int64 | 搜索结果数量 |
| `created_at` | datetime | 创建时间 |

索引：`idx_user_id`、`idx_created_at`

### 8.2 热搜词表 (`hot_keyword`)

统计热门搜索词。

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | bigint unsigned | 主键 |
| `keyword` | varchar(128) | 关键词（唯一） |
| `search_count` | bigint unsigned | 搜索次数 |
| `sort` | int64 | 排序值 |
| `status` | int64 | 状态：1-显示，2-隐藏 |
| `updated_at` | datetime | 更新时间 |

索引：`uk_keyword`（唯一索引）、`idx_search_count`

### 8.3 ES 数据同步记录表 (`es_sync_log`)

记录需要同步到 Elasticsearch 的数据变更。

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | bigint unsigned | 主键 |
| `table_name` | varchar(64) | 表名：`product_spu`、`product_sku` |
| `record_id` | bigint unsigned | 记录 ID |
| `action` | varchar(20) | 操作类型：`INSERT`、`UPDATE`、`DELETE` |
| `status` | int64 | 同步状态：0-待同步，1-已同步，2-同步失败 |
| `retry_count` | int | 重试次数 |
| `error_msg` | varchar(500) | 错误信息（可空） |
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
| 售后 | `after_sale.order_id` | `order_info.id` | 外键逻辑关联 |
| 售后商品 | `after_sale_item.after_sale_id` | `after_sale.id` | 外键逻辑关联 |
| 支付 | `payment.order_sn` | `order_info.order_sn` | 外键逻辑关联 |
| 退款 | `refund.payment_sn` | `payment.payment_sn` | 外键逻辑关联 |
| 秒杀预占 | `seckill_preorder.activity_id` | `seckill_activity.id` | 外键逻辑关联 |
| 秒杀预占 | `seckill_preorder.order_sn` | `order_info.order_sn` | 支付成功后关联 |
| 用户优惠券 | `user_coupon.coupon_id` | `coupon.id` | 外键逻辑关联 |
| 管理员 | `admin.role_id` | `role.id` | 外键逻辑关联 |
| 帮助文章 | `help_article.category_id` | `help_category.id` | 外键逻辑关联 |

> 由于微服务架构，各数据库之间不建立物理外键，通过应用层保证数据一致性。

---

## 10. 订单状态流转说明

### 10.1 订单状态

| 状态码 | 状态名称 | 说明 |
|--------|----------|------|
| 10 | 待支付 | 订单已创建，等待用户支付 |
| 20 | 已支付 | 用户已支付，等待商家发货 |
| 30 | 已发货 | 商家已发货，等待用户确认收货 |
| 40 | 已完成 | 用户已确认收货或系统自动确认 |
| 50 | 已取消 | 订单已取消（超时/用户取消/管理员取消） |
| 60 | 售后中 | 订单正在售后处理中 |

### 10.2 售后状态

| 状态码 | 状态名称 | 说明 |
|--------|----------|------|
| 10 | 待审核 | 等待商家审核 |
| 20 | 待商家收货 | 审核通过，等待用户退货/商家收货 |
| 30 | 退款中 | 商家已收货/仅退款审核通过，退款处理中 |
| 40 | 已完成 | 退款成功，售后完成 |
| 50 | 已拒绝 | 商家拒绝售后申请 |
| 60 | 已取消 | 用户取消售后申请 |

### 10.3 支付状态

| 状态码 | 状态名称 | 说明 |
|--------|----------|------|
| 0 | 待支付 | 等待用户支付 |
| 1 | 支付成功 | 第三方支付成功 |
| 2 | 支付失败 | 第三方支付失败 |
| 3 | 退款中 | 退款处理中 |
| 4 | 退款成功 | 退款完成 |

---

## 11. 版本历史

| 日期 | 版本 | 说明 |
|------|------|------|
| 2026-04-06 | 1.0 | 初始版本，包含用户、商品、订单、支付、营销、后台、搜索七大服务数据库表结构 |
| 2026-05-30 | 1.1 | 更新订单主表字段（金额单位统一为分），新增支付类型、物流信息、取消原因等字段；新增售后表和售后商品表；更新商品SKU表（新增锁定库存、重量等字段）；补充错误码文档 |
