-- =============================================
-- 服务: 订单服务 (Order Service)
-- 数据库: shop_order
-- 说明: 购物车、订单、订单商品快照
-- 修正: 增加 order_type 字段，统一普通订单和秒杀订单
-- 金额字段统一使用 BIGINT 类型，单位为分（避免浮点数精度问题）
-- =============================================
-- goctl model mysql ddl --src=./apps/service/order/sql/03-order.sql --dir=./apps/service/order/rpc/internal/model
CREATE DATABASE IF NOT EXISTS `shop_order`
    CHARACTER SET utf8mb4
    COLLATE utf8mb4_unicode_ci;

USE `shop_order`;

-- =============================================
-- 购物车表
-- =============================================
CREATE TABLE IF NOT EXISTS `cart` (
    `id`         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '购物车ID',
    `user_id`    BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    `sku_id`     BIGINT UNSIGNED NOT NULL COMMENT 'SKU ID',
    `count`      INT UNSIGNED    NOT NULL DEFAULT '1' COMMENT '商品数量',
    `selected`   BOOL            NOT NULL DEFAULT '1' COMMENT '是否选中：false-未选中，true-已选中',
    `created_at` DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_user_sku` (`user_id`, `sku_id`)
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci
  COMMENT='购物车表';

-- =============================================
-- 订单主表
-- =============================================
CREATE TABLE IF NOT EXISTS `order_info` (
    `id`                    BIGINT UNSIGNED NOT NULL COMMENT '订单ID',
    `order_sn`              VARCHAR(64)     NOT NULL COMMENT '订单号（业务唯一标识）',
    `order_type`            TINYINT         NOT NULL DEFAULT '1' COMMENT '订单类型：1-普通订单，2-秒杀订单，3-拼团订单',
    `idempotency_key`       VARCHAR(64)     NOT NULL COMMENT '幂等键（防止重复提交）',
    `user_id`               BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    `status`                TINYINT         NOT NULL DEFAULT '10' COMMENT '订单状态：10-待支付，20-已支付，30-已发货，40-已完成，50-已取消，60-售后中',
    `address_snap`          JSON            NOT NULL COMMENT '收货地址快照（JSON格式）',
    `total_amount`          BIGINT          NOT NULL DEFAULT '0' COMMENT '商品总金额（单位：分）',
    `freight_amount`        BIGINT          NOT NULL DEFAULT '0' COMMENT '运费金额（单位：分）',
    `coupon_id`             BIGINT UNSIGNED NOT NULL DEFAULT '0'  COMMENT '优惠券ID：0-未使用优惠券',
    `coupon_discount`       BIGINT                   DEFAULT '0' COMMENT '优惠券优惠金额（单位：分）',
    `pay_amount`            BIGINT          NOT NULL DEFAULT '0' COMMENT '实付金额（单位：分）',
    `remark`                VARCHAR(500)             DEFAULT '' COMMENT '订单备注（用户留言）',
    `pay_time`              DATETIME                 DEFAULT NULL COMMENT '支付时间',
    `pay_type`              TINYINT                  DEFAULT '1' COMMENT '支付类型：1-微信支付，2-支付宝支付，3-其他支付',
    `delivery_time`         DATETIME                 DEFAULT NULL COMMENT '发货时间',
    `delivery_sn`           VARCHAR(64)              DEFAULT '' COMMENT '物流单号',
    `delivery_corp`         VARCHAR(64)              DEFAULT '' COMMENT '物流公司名称',
    `receive_time`          DATETIME                 DEFAULT NULL COMMENT '确认收货时间',
    `cancel_time`           DATETIME                 DEFAULT NULL COMMENT '取消时间',
    `cancel_reason_type`    TINYINT                  DEFAULT 0 COMMENT '取消原因类型：0-未取消 1-用户主动取消 2-超时未支付取消 3-库存不足 4-管理员取消 5-其他',
    `cancel_reason`         VARCHAR(255)             DEFAULT '' COMMENT '取消原因详情（用户填写/系统备注）',
    `seckill_activity_id`   BIGINT UNSIGNED NOT NULL DEFAULT '0' COMMENT '秒杀活动ID：普通订单为0',
    `seckill_price`         BIGINT                   DEFAULT '0' COMMENT '秒杀价格快照（单位：分）',
    `created_at`            DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `expire_time`           DATETIME                 DEFAULT NULL COMMENT '订单支付过期时间',
    `updated_at`            DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_order_sn` (`order_sn`),
    UNIQUE KEY `uk_user_idempotency` (`user_id`, `idempotency_key`),
    KEY `idx_user_id` (`user_id`),
    KEY `idx_status` (`status`),
    KEY `idx_order_type` (`order_type`),
    KEY `idx_seckill_activity_id` (`seckill_activity_id`),
    KEY `idx_user_created` (user_id, created_at)
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci
  COMMENT='订单主表';

-- =============================================
-- 订单商品表
-- =============================================
CREATE TABLE IF NOT EXISTS `order_item` (
    `id`           BIGINT UNSIGNED NOT NULL COMMENT '订单商品ID',
    `order_id`     BIGINT UNSIGNED NOT NULL COMMENT '订单ID',
    `order_sn`     VARCHAR(64)     NOT NULL COMMENT '订单号',
    `sku_id`       BIGINT UNSIGNED NOT NULL COMMENT 'SKU ID',
    `spu_id`       BIGINT UNSIGNED NOT NULL COMMENT 'SPU ID',
    `spu_name`     VARCHAR(128)    NOT NULL DEFAULT 'SPU 名称快照',
    `sku_name`     VARCHAR(128)    NOT NULL DEFAULT '' COMMENT 'SKU名称快照',
    `sku_pic`      VARCHAR(500)    NOT NULL DEFAULT '' COMMENT 'SKU图片快照',
    `price`        BIGINT          NOT NULL DEFAULT '0' COMMENT '成交单价（单位：分）',
    `count`        INT             NOT NULL DEFAULT '1' COMMENT '购买数量',
    `total_amount` BIGINT          NOT NULL DEFAULT '0' COMMENT '小计金额（单价×数量，单位：分）',
    `is_seckill`   TINYINT         NOT NULL DEFAULT '0' COMMENT '是否秒杀商品：0-否，1-是',
    `created_at`   DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    PRIMARY KEY (`id`),
    KEY `idx_order_id` (`order_id`),
    KEY `idx_order_sn` (`order_sn`),
    KEY `idx_sku_id` (`sku_id`)
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci
  COMMENT='订单商品表';

-- =============================================
-- 售后主表
-- =============================================
CREATE TABLE IF NOT EXISTS `after_sale` (
    `id`                       BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '售后单ID',
    `after_sale_sn`            VARCHAR(64)     NOT NULL COMMENT '售后单号',
    `idempotency_key`          VARCHAR(64)     NOT NULL COMMENT '幂等键（同一用户下唯一）',
    `order_id`                 BIGINT UNSIGNED NOT NULL COMMENT '订单ID',
    `order_sn`                 VARCHAR(64)     NOT NULL COMMENT '订单号',
    `user_id`                  BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    `type`                     TINYINT         NOT NULL DEFAULT 1 COMMENT '售后类型：1仅退款 2退货退款',
    `status`                   TINYINT         NOT NULL DEFAULT 10 COMMENT '状态：10待审核 20待商家收货 30退款中 40已完成 50已拒绝 60已取消',
    `refund_status`            TINYINT         NOT NULL DEFAULT 0 COMMENT '退款状态：0未退款 10退款中 20退款成功 30退款失败',
    `apply_amount`             BIGINT          NOT NULL DEFAULT 0 COMMENT '申请退款金额（分）',
    `approved_amount`          BIGINT                   DEFAULT 0 COMMENT '审核通过金额（分）',
    `real_refund_amount`       BIGINT                   DEFAULT 0 COMMENT '实际退款金额（分）',
    `reason`                   VARCHAR(200)    NOT NULL DEFAULT '' COMMENT '售后原因',
    `audit_remark`             VARCHAR(500)             DEFAULT '' COMMENT '审核备注',
    `audit_time`               DATETIME                 DEFAULT NULL COMMENT '审核时间',
    `images`                   JSON                     DEFAULT NULL COMMENT '凭证图片URL数组',
    `return_tracking_sn`       VARCHAR(100)             DEFAULT '' COMMENT '退货物流单号',
    `return_tracking_corp`     VARCHAR(50)              DEFAULT '' COMMENT '退货物流公司',
    `return_received_time`     DATETIME                 DEFAULT NULL COMMENT '商家收货时间',
    `refund_sn`                VARCHAR(64)              DEFAULT '' COMMENT '商户退款单号（唯一）',
    `refund_time`              DATETIME                 DEFAULT NULL COMMENT '退款时间',
    `third_refund_sn`          VARCHAR(64)              DEFAULT '' COMMENT '第三方退款单号',
    `cancel_time`              DATETIME                 DEFAULT NULL COMMENT '取消时间',
    `created_at`               DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`               DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_after_sale_sn` (`after_sale_sn`),
    UNIQUE KEY `uk_user_idempotency` (`user_id`, `idempotency_key`),   -- 幂等键带用户作用域
    UNIQUE KEY `uk_refund_sn` (`refund_sn`),                           -- 退款单号唯一，防重复退款
    KEY `idx_order_sn` (`order_sn`),
    KEY `idx_user_id` (`user_id`),
    KEY `idx_status` (`status`)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='售后主表';

-- =============================================
-- 售后商品表（快照）
-- =============================================
CREATE TABLE IF NOT EXISTS `after_sale_item` (
    `id`               BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `after_sale_sn`    VARCHAR(64)     NOT NULL COMMENT '售后单号',
    `order_item_id`    BIGINT UNSIGNED NOT NULL COMMENT '原订单商品项ID',
    `sku_id`           BIGINT UNSIGNED NOT NULL,
    `spu_id`           BIGINT UNSIGNED NOT NULL,
    `product_name`     VARCHAR(128)    NOT NULL DEFAULT '',
    `sku_name`         VARCHAR(128)    NOT NULL DEFAULT '',
    `sku_pic`          VARCHAR(500)    NOT NULL DEFAULT '',
    `price`            BIGINT          NOT NULL DEFAULT 0 COMMENT '原成交单价（分）',
    `quantity`         INT             NOT NULL DEFAULT 1 COMMENT '售后数量',
    `total_amount`     BIGINT          NOT NULL DEFAULT 0 COMMENT '原小计金额（单价×数量，单位：分）',
    `refund_amount`    BIGINT          NOT NULL DEFAULT 0 COMMENT '该项退款金额（分）',
    `created_at`       DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_after_sale_sn` (`after_sale_sn`),
    KEY `idx_order_item_id` (`order_item_id`)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='售后商品快照表';

-- =============================================
-- 索引优化（高并发场景）
-- =============================================
ALTER TABLE `order_info` ADD INDEX `idx_status_created` (`status`, `created_at`);
ALTER TABLE `order_info` ADD INDEX `idx_expire_status` (`expire_time`, `status`);
ALTER TABLE `order_info` ADD INDEX `idx_user_status` (`user_id`, `status`);
ALTER TABLE `order_item` ADD INDEX `idx_spu_id` (`spu_id`);
ALTER TABLE `after_sale` ADD INDEX `idx_user_status` (`user_id`, `status`);
ALTER TABLE `after_sale` ADD INDEX `idx_status_created` (`status`, `created_at`);
ALTER TABLE `after_sale` ADD INDEX `idx_type` (`type`);
ALTER TABLE `after_sale_item` ADD INDEX `idx_sku_id` (`sku_id`);