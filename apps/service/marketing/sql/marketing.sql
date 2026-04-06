-- =============================================
-- 服务: 营销服务 (Marketing Service)
-- 数据库: shop_marketing
-- 说明: 优惠券、秒杀、满减活动
-- 修正: seckill_order 表只做预占记录，不存储完整订单
-- =============================================

CREATE DATABASE IF NOT EXISTS `shop_marketing` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE `shop_marketing`;

-- 优惠券表
CREATE TABLE IF NOT EXISTS `coupon` (
                                        `id` bigint unsigned NOT NULL AUTO_INCREMENT,
                                        `name` varchar(128) NOT NULL,
    `type` tinyint NOT NULL COMMENT '1:满减券 2:折扣券 3:无门槛券',
    `threshold_amount` decimal(10,2) DEFAULT '0.00',
    `reduce_amount` decimal(10,2) DEFAULT '0.00',
    `discount_rate` decimal(3,2) DEFAULT '0.00',
    `max_discount_amount` decimal(10,2) DEFAULT '0.00',
    `total_quantity` int NOT NULL,
    `used_quantity` int NOT NULL DEFAULT '0',
    `per_user_limit` int NOT NULL DEFAULT '1',
    `start_time` datetime NOT NULL,
    `end_time` datetime NOT NULL,
    `status` tinyint NOT NULL DEFAULT '1' COMMENT '1:启用 2:禁用',
    `description` varchar(255) DEFAULT '',
    `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_time` (`start_time`, `end_time`)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='优惠券表';

-- 用户优惠券表
CREATE TABLE IF NOT EXISTS `user_coupon` (
                                             `id` bigint unsigned NOT NULL AUTO_INCREMENT,
                                             `coupon_id` bigint unsigned NOT NULL,
                                             `user_id` bigint unsigned NOT NULL,
                                             `order_sn` varchar(64) DEFAULT '',
    `status` tinyint NOT NULL DEFAULT '0' COMMENT '0:未使用 1:已使用 2:已过期',
    `used_time` datetime DEFAULT NULL,
    `source` varchar(32) DEFAULT 'ADMIN',
    `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `expire_time` datetime NOT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_user_coupon` (`user_id`, `coupon_id`),
    KEY `idx_coupon_id` (`coupon_id`),
    KEY `idx_user_id` (`user_id`)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户优惠券表';

-- 秒杀活动表
CREATE TABLE IF NOT EXISTS `seckill_activity` (
                                                  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
                                                  `name` varchar(128) NOT NULL,
    `sku_id` bigint unsigned NOT NULL,
    `sku_name` varchar(128) NOT NULL DEFAULT '' COMMENT '商品名称快照',
    `sku_pic` varchar(500) NOT NULL DEFAULT '' COMMENT '商品图片快照',
    `seckill_price` decimal(10,2) NOT NULL,
    `stock` int NOT NULL,
    `sold_count` int NOT NULL DEFAULT '0',
    `per_user_limit` int NOT NULL DEFAULT '1',
    `start_time` datetime NOT NULL,
    `end_time` datetime NOT NULL,
    `status` tinyint NOT NULL DEFAULT '1' COMMENT '1:进行中 2:已结束 3:已取消',
    `version` int unsigned NOT NULL DEFAULT '0' COMMENT '乐观锁',
    `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_sku_id` (`sku_id`),
    KEY `idx_time` (`start_time`, `end_time`)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='秒杀活动表';

-- 秒杀预占记录表
-- 秒杀成功时先在此表创建记录，支付成功后关联到 order_info
CREATE TABLE IF NOT EXISTS `seckill_preorder` (
                                                  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
                                                  `preorder_sn` varchar(64) NOT NULL COMMENT '预占单号',
    `activity_id` bigint unsigned NOT NULL,
    `sku_id` bigint unsigned NOT NULL,
    `user_id` bigint unsigned NOT NULL,
    `seckill_price` decimal(10,2) NOT NULL,
    `quantity` int NOT NULL DEFAULT '1',
    `order_sn` varchar(64) DEFAULT '' COMMENT '支付成功后关联的正式订单号',
    `status` tinyint NOT NULL DEFAULT '0' COMMENT '0:预占成功 1:已支付 2:已取消 3:已超时',
    `expire_time` datetime NOT NULL COMMENT '预占过期时间（通常5-15分钟）',
    `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_preorder_sn` (`preorder_sn`),
    UNIQUE KEY `uk_user_activity` (`user_id`, `activity_id`),
    KEY `idx_activity_id` (`activity_id`),
    KEY `idx_status` (`status`),
    KEY `idx_expire_time` (`expire_time`)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='秒杀预占记录表（支付成功后才生成正式订单）';