-- =============================================
-- 服务: 营销服务 (Marketing Service)
-- 数据库: shop_marketing
-- 说明: 优惠券、秒杀、满减活动
-- 金额字段统一使用 BIGINT 类型，单位为分（避免浮点数精度问题）
-- =============================================

CREATE DATABASE IF NOT EXISTS `shop_marketing`
    CHARACTER SET utf8mb4
    COLLATE utf8mb4_unicode_ci;

USE `shop_marketing`;

-- =============================================
-- 优惠券表
-- =============================================
CREATE TABLE IF NOT EXISTS `coupon` (
    `id`                    BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '优惠券ID',
    `name`                  VARCHAR(128)    NOT NULL COMMENT '优惠券名称',
    `type`                  TINYINT         NOT NULL COMMENT '优惠券类型：1-满减券，2-折扣券，3-无门槛券',
    `threshold_amount`      BIGINT                   DEFAULT '0' COMMENT '使用门槛金额（单位：分）：0表示无门槛',
    `reduce_amount`         BIGINT                   DEFAULT '0' COMMENT '减免金额（单位：分，满减券使用）',
    `discount_rate`         INT                      DEFAULT '0' COMMENT '折扣率（万分比）：8000表示0.8=8折，仅折扣券使用',
    `max_discount_amount`   BIGINT                   DEFAULT '0' COMMENT '最大优惠金额（单位：分，折扣券封顶）',
    `total_quantity`        INT             NOT NULL COMMENT '发放总量',
    `used_quantity`         INT             NOT NULL DEFAULT '0' COMMENT '已使用数量',
    `per_user_limit`        INT             NOT NULL DEFAULT '1' COMMENT '每人限领数量',
    `start_time`            DATETIME        NOT NULL COMMENT '生效开始时间',
    `end_time`              DATETIME        NOT NULL COMMENT '生效结束时间',
    `status`                TINYINT         NOT NULL DEFAULT '1' COMMENT '状态：1-启用，2-禁用',
    `description`           VARCHAR(255)             DEFAULT '' COMMENT '优惠券描述',
    `created_at`            DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`            DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    KEY `idx_time` (`start_time`, `end_time`)
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci
  COMMENT='优惠券表';

-- =============================================
-- 用户优惠券表
-- =============================================
CREATE TABLE IF NOT EXISTS `user_coupon` (
    `id`          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '用户优惠券ID',
    `coupon_id`   BIGINT UNSIGNED NOT NULL COMMENT '优惠券ID',
    `user_id`     BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    `order_sn`    VARCHAR(64)              DEFAULT '' COMMENT '使用的订单号：空表示未使用',
    `status`      TINYINT         NOT NULL DEFAULT '0' COMMENT '状态：0-未使用，1-已使用，2-已过期',
    `used_time`   DATETIME                 DEFAULT NULL COMMENT '使用时间',
    `source`      VARCHAR(32)              DEFAULT 'ADMIN' COMMENT '领取来源：ADMIN-后台发放，REGISTER-注册赠送，ACTIVITY-活动领取',
    `created_at`  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '领取时间',
    `expire_time` DATETIME        NOT NULL COMMENT '过期时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_user_coupon` (`user_id`, `coupon_id`),
    KEY `idx_coupon_id` (`coupon_id`),
    KEY `idx_user_id` (`user_id`)
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci
  COMMENT='用户优惠券表';

-- =============================================
-- 秒杀活动表
-- =============================================
CREATE TABLE IF NOT EXISTS `seckill_activity` (
    `id`              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '秒杀活动ID',
    `name`            VARCHAR(128)    NOT NULL COMMENT '秒杀活动名称',
    `sku_id`          BIGINT UNSIGNED NOT NULL COMMENT '参与秒杀的SKU ID',
    `sku_name`        VARCHAR(128)    NOT NULL DEFAULT '' COMMENT '商品名称快照',
    `sku_pic`         VARCHAR(500)    NOT NULL DEFAULT '' COMMENT '商品图片快照',
    `seckill_price`   BIGINT          NOT NULL COMMENT '秒杀价格（单位：分）',
    `stock`           INT             NOT NULL COMMENT '秒杀库存总量',
    `sold_count`      INT             NOT NULL DEFAULT '0' COMMENT '已售数量',
    `per_user_limit`  INT             NOT NULL DEFAULT '1' COMMENT '每人限购数量',
    `start_time`      DATETIME        NOT NULL COMMENT '活动开始时间',
    `end_time`        DATETIME        NOT NULL COMMENT '活动结束时间',
    `status`          TINYINT         NOT NULL DEFAULT '1' COMMENT '活动状态：1-进行中，2-已结束，3-已取消',
    `version`         INT UNSIGNED    NOT NULL DEFAULT '0' COMMENT '乐观锁版本号（用于并发控制）',
    `created_at`      DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`      DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    KEY `idx_sku_id` (`sku_id`),
    KEY `idx_time` (`start_time`, `end_time`)
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci
  COMMENT='秒杀活动表';

-- =============================================
-- 秒杀预占记录表
-- 说明: 秒杀成功时先在此表创建记录，支付成功后关联到 order_info
-- =============================================
CREATE TABLE IF NOT EXISTS `seckill_preorder` (
    `id`              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '预占记录ID',
    `preorder_sn`     VARCHAR(64)     NOT NULL COMMENT '预占单号（业务唯一标识）',
    `activity_id`     BIGINT UNSIGNED NOT NULL COMMENT '秒杀活动ID',
    `sku_id`          BIGINT UNSIGNED NOT NULL COMMENT 'SKU ID',
    `user_id`         BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    `seckill_price`   BIGINT          NOT NULL COMMENT '秒杀价格快照（单位：分）',
    `quantity`        INT             NOT NULL DEFAULT '1' COMMENT '购买数量',
    `order_sn`        VARCHAR(64)              DEFAULT '' COMMENT '正式订单号：支付成功后生成，空表示未支付',
    `status`          TINYINT         NOT NULL DEFAULT '0' COMMENT '预占状态：0-预占成功，1-已支付，2-已取消，3-已超时',
    `expire_time`     DATETIME        NOT NULL COMMENT '预占过期时间（通常5-15分钟）',
    `created_at`      DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`      DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_preorder_sn` (`preorder_sn`),
    UNIQUE KEY `uk_user_activity` (`user_id`, `activity_id`),
    KEY `idx_activity_id` (`activity_id`),
    KEY `idx_status` (`status`),
    KEY `idx_expire_time` (`expire_time`)
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci
  COMMENT='秒杀预占记录表（支付成功后才生成正式订单）';
