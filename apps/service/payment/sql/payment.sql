-- =============================================
-- 服务: 支付服务 (Payment Service)
-- 数据库: shop_payment
-- 说明: 支付记录、支付回调、退款记录
-- =============================================

CREATE DATABASE IF NOT EXISTS `shop_payment` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE `shop_payment`;

-- 支付记录表
CREATE TABLE IF NOT EXISTS `payment` (
                                         `id` bigint unsigned NOT NULL AUTO_INCREMENT,
                                         `payment_sn` varchar(64) NOT NULL,
    `order_sn` varchar(64) NOT NULL,
    `user_id` bigint unsigned NOT NULL,
    `amount` decimal(10,2) NOT NULL,
    `channel` varchar(32) NOT NULL COMMENT 'wechat, alipay',
    `channel_order_sn` varchar(128) DEFAULT '',
    `transaction_id` varchar(128) DEFAULT '',
    `status` tinyint NOT NULL DEFAULT '0' COMMENT '0:待支付 1:成功 2:失败 3:退款中 4:退款成功',
    `pay_time` datetime DEFAULT NULL,
    `callback_data` json DEFAULT NULL,
    `error_msg` varchar(255) DEFAULT '',
    `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_payment_sn` (`payment_sn`),
    UNIQUE KEY `uk_order_sn` (`order_sn`),
    KEY `idx_user_id` (`user_id`),
    KEY `idx_status` (`status`)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='支付记录表';

-- 支付回调日志表
CREATE TABLE IF NOT EXISTS `payment_callback_log` (
                                                      `id` bigint unsigned NOT NULL AUTO_INCREMENT,
                                                      `payment_sn` varchar(64) NOT NULL DEFAULT '',
    `order_sn` varchar(64) NOT NULL DEFAULT '',
    `channel` varchar(32) NOT NULL,
    `request_body` text,
    `response_body` text,
    `status` tinyint NOT NULL DEFAULT '0' COMMENT '0:处理中 1:成功 2:失败',
    `error_msg` varchar(500) DEFAULT '',
    `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_payment_sn` (`payment_sn`),
    KEY `idx_order_sn` (`order_sn`)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='支付回调日志表';

-- 退款记录表
CREATE TABLE IF NOT EXISTS `refund` (
                                        `id` bigint unsigned NOT NULL AUTO_INCREMENT,
                                        `refund_sn` varchar(64) NOT NULL,
    `payment_sn` varchar(64) NOT NULL,
    `order_sn` varchar(64) NOT NULL,
    `user_id` bigint unsigned NOT NULL,
    `amount` decimal(10,2) NOT NULL,
    `reason` varchar(255) NOT NULL,
    `status` tinyint NOT NULL DEFAULT '0' COMMENT '0:处理中 1:成功 2:失败',
    `transaction_id` varchar(128) DEFAULT '',
    `refund_time` datetime DEFAULT NULL,
    `error_msg` varchar(255) DEFAULT '',
    `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_refund_sn` (`refund_sn`),
    KEY `idx_payment_sn` (`payment_sn`),
    KEY `idx_order_sn` (`order_sn`)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='退款记录表';