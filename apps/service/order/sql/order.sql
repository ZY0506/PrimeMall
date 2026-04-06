-- =============================================
-- 服务: 订单服务 (Order Service)
-- 数据库: shop_order
-- 说明: 购物车、订单、订单商品快照
-- 修正: 增加 order_type 字段，统一普通订单和秒杀订单
-- =============================================

CREATE DATABASE IF NOT EXISTS `shop_order` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE `shop_order`;

-- 购物车表
CREATE TABLE IF NOT EXISTS `cart` (
                                      `id` bigint unsigned NOT NULL AUTO_INCREMENT,
                                      `user_id` bigint unsigned NOT NULL,
                                      `sku_id` bigint unsigned NOT NULL,
                                      `count` int unsigned NOT NULL DEFAULT '1',
                                      `selected` tinyint NOT NULL DEFAULT '1',
                                      `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
                                      `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
                                      PRIMARY KEY (`id`),
    UNIQUE KEY `uk_user_sku` (`user_id`, `sku_id`)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='购物车表';

-- 订单主表
CREATE TABLE IF NOT EXISTS `order_info` (
                                            `id` bigint unsigned NOT NULL AUTO_INCREMENT,
                                            `order_sn` varchar(64) NOT NULL COMMENT '订单号',
    `order_type` tinyint NOT NULL DEFAULT '1' COMMENT '订单类型: 1:普通订单 2:秒杀订单 3:拼团订单',
    `idempotency_key` varchar(64) NOT NULL COMMENT '幂等键',
    `user_id` bigint unsigned NOT NULL,
    `status` tinyint NOT NULL DEFAULT '0' COMMENT '10:待支付 20:已支付 30:已发货 40:已完成 50:已取消 60:售后中',
    `address_snap` json NOT NULL COMMENT '地址快照',
    `total_amount` decimal(10,2) NOT NULL DEFAULT '0.00' COMMENT '商品总金额',
    `freight_amount` decimal(10,2) NOT NULL DEFAULT '0.00' COMMENT '运费',
    `coupon_id` bigint unsigned DEFAULT '0',
    `coupon_discount` decimal(10,2) DEFAULT '0.00',
    `pay_amount` decimal(10,2) NOT NULL DEFAULT '0.00' COMMENT '实付金额',
    `remark` varchar(500) DEFAULT '',
    -- 时间字段
    `pay_time` datetime DEFAULT NULL,
    `delivery_time` datetime DEFAULT NULL,
    `receive_time` datetime DEFAULT NULL,
    `cancel_time` datetime DEFAULT NULL,
    -- 秒杀相关字段（仅秒杀订单使用）
    `seckill_activity_id` bigint unsigned DEFAULT '0' COMMENT '秒杀活动ID，普通订单为0',
    `seckill_price` decimal(10,2) DEFAULT '0.00' COMMENT '秒杀价格快照',
    `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_order_sn` (`order_sn`),
    UNIQUE KEY `uk_idempotency_key` (`idempotency_key`),
    KEY `idx_user_id` (`user_id`),
    KEY `idx_status` (`status`),
    KEY `idx_order_type` (`order_type`),
    KEY `idx_seckill_activity_id` (`seckill_activity_id`),
    KEY `idx_created_at` (`created_at`)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='订单主表';

-- 订单商品表
CREATE TABLE IF NOT EXISTS `order_item` (
                                            `id` bigint unsigned NOT NULL AUTO_INCREMENT,
                                            `order_id` bigint unsigned NOT NULL,
                                            `order_sn` varchar(64) NOT NULL,
    `sku_id` bigint unsigned NOT NULL,
    `spu_id` bigint unsigned NOT NULL,
    `sku_name` varchar(128) NOT NULL DEFAULT '',
    `sku_pic` varchar(500) NOT NULL DEFAULT '',
    `spec_data` json DEFAULT NULL,
    `price` decimal(10,2) NOT NULL DEFAULT '0.00' COMMENT '成交单价',
    `count` int NOT NULL DEFAULT '1',
    `total_amount` decimal(10,2) NOT NULL DEFAULT '0.00',
    `is_seckill` tinyint NOT NULL DEFAULT '0' COMMENT '是否秒杀商品: 0否 1是',
    `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_order_id` (`order_id`),
    KEY `idx_order_sn` (`order_sn`)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='订单商品表';