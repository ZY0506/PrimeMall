-- =============================================
-- 服务: 支付服务 (Payment Service)
-- 数据库: shop_payment
-- 说明: 支付记录、支付回调、退款记录
-- 金额字段统一使用 BIGINT 类型，单位为分（避免浮点数精度问题）
-- =============================================

CREATE DATABASE IF NOT EXISTS `shop_payment`
CHARACTER SET utf8mb4
COLLATE utf8mb4_unicode_ci;

USE `shop_payment`;

-- =============================================
-- 支付记录表
-- =============================================
CREATE TABLE IF NOT EXISTS `payment` (
    `id`               BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '支付记录ID',
    `payment_sn`       VARCHAR(64)     NOT NULL COMMENT '支付流水号（业务唯一标识）',
    `order_sn`         VARCHAR(64)     NOT NULL COMMENT '订单号',
    `user_id`          BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    `amount`           BIGINT          NOT NULL COMMENT '支付金额（单位：分）',
    `channel`          VARCHAR(32)     NOT NULL COMMENT '支付渠道：wechat-微信支付，alipay-支付宝',
    `channel_order_sn` VARCHAR(128)             DEFAULT '' COMMENT '第三方支付平台订单号',
    `transaction_id`   VARCHAR(128)             DEFAULT '' COMMENT '第三方支付交易流水号',
    `status`           TINYINT         NOT NULL DEFAULT '0' COMMENT '支付状态：0-待支付，1-支付成功，2-支付失败，3-退款中，4-退款成功',
    `pay_time`         DATETIME                 DEFAULT NULL COMMENT '支付成功时间',
    `callback_data`    JSON                     DEFAULT NULL COMMENT '支付回调数据快照（JSON格式）',
    `error_msg`        VARCHAR(255)             DEFAULT '' COMMENT '错误信息',
    `created_at`       DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`       DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_payment_sn` (`payment_sn`),
    UNIQUE KEY `uk_order_sn` (`order_sn`),
    KEY `idx_user_id` (`user_id`),
    KEY `idx_status` (`status`)
) ENGINE=InnoDB
DEFAULT CHARSET=utf8mb4
COLLATE=utf8mb4_unicode_ci
COMMENT='支付记录表';

-- =============================================
-- 支付回调日志表
-- =============================================
CREATE TABLE IF NOT EXISTS `payment_callback_log` (
    `id`              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '日志ID',
    `payment_sn`      VARCHAR(64)     NOT NULL DEFAULT '' COMMENT '支付流水号',
    `order_sn`        VARCHAR(64)     NOT NULL DEFAULT '' COMMENT '订单号',
    `channel`         VARCHAR(32)     NOT NULL COMMENT '支付渠道：wechat-微信支付，alipay-支付宝',
    `request_body`    TEXT                     DEFAULT NULL COMMENT '回调请求原始数据',
    `response_body`   TEXT                     DEFAULT NULL COMMENT '回调响应数据',
    `status`          TINYINT         NOT NULL DEFAULT '0' COMMENT '处理状态：0-处理中，1-处理成功，2-处理失败',
    `error_msg`       VARCHAR(500)             DEFAULT '' COMMENT '错误信息',
    `created_at`      DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    PRIMARY KEY (`id`),
    KEY `idx_payment_sn` (`payment_sn`),
    KEY `idx_order_sn` (`order_sn`)
) ENGINE=InnoDB
DEFAULT CHARSET=utf8mb4
COLLATE=utf8mb4_unicode_ci
COMMENT='支付回调日志表';

-- =============================================
-- 退款记录表
-- =============================================
CREATE TABLE IF NOT EXISTS `refund` (
    `id`             BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '退款记录ID',
    `refund_sn`      VARCHAR(64)     NOT NULL COMMENT '退款单号（业务唯一标识）',
    `payment_sn`     VARCHAR(64)     NOT NULL COMMENT '支付流水号',
    `order_sn`       VARCHAR(64)     NOT NULL COMMENT '订单号',
    `user_id`        BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    `amount`         BIGINT          NOT NULL COMMENT '退款金额（单位：分）',
    `reason`         VARCHAR(255)    NOT NULL COMMENT '退款原因',
    `status`         TINYINT         NOT NULL DEFAULT '0' COMMENT '退款状态：0-处理中，1-退款成功，2-退款失败',
    `transaction_id` VARCHAR(128)             DEFAULT '' COMMENT '第三方退款交易流水号',
    `refund_time`    DATETIME                 DEFAULT NULL COMMENT '退款成功时间',
    `error_msg`      VARCHAR(255)             DEFAULT '' COMMENT '错误信息',
    `created_at`     DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`     DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_refund_sn` (`refund_sn`),
    KEY `idx_payment_sn` (`payment_sn`),
    KEY `idx_order_sn` (`order_sn`)
) ENGINE=InnoDB
DEFAULT CHARSET=utf8mb4
COLLATE=utf8mb4_unicode_ci
COMMENT='退款记录表';
