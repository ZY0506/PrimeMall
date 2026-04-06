-- =============================================
-- 服务: 订单服务 (Order Service)
-- 数据库: shop_order
-- 说明: 购物车、订单、订单商品快照
-- 修正: 增加 order_type 字段，统一普通订单和秒杀订单
-- =============================================

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
    `selected`   TINYINT         NOT NULL DEFAULT '1' COMMENT '是否选中：0-未选中，1-已选中',
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
    `id`                    BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '订单ID',
    `order_sn`              VARCHAR(64)     NOT NULL COMMENT '订单号（业务唯一标识）',
    `order_type`            TINYINT         NOT NULL DEFAULT '1' COMMENT '订单类型：1-普通订单，2-秒杀订单，3-拼团订单',
    `idempotency_key`       VARCHAR(64)     NOT NULL COMMENT '幂等键（防止重复提交）',
    `user_id`               BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    `status`                TINYINT         NOT NULL DEFAULT '0' COMMENT '订单状态：10-待支付，20-已支付，30-已发货，40-已完成，50-已取消，60-售后中',
    `address_snap`          JSON            NOT NULL COMMENT '收货地址快照（JSON格式）',
    `total_amount`          DECIMAL(10,2)   NOT NULL DEFAULT '0.00' COMMENT '商品总金额',
    `freight_amount`        DECIMAL(10,2)   NOT NULL DEFAULT '0.00' COMMENT '运费金额',
    `coupon_id`             BIGINT UNSIGNED          DEFAULT '0' COMMENT '优惠券ID：0-未使用优惠券',
    `coupon_discount`       DECIMAL(10,2)            DEFAULT '0.00' COMMENT '优惠券优惠金额',
    `pay_amount`            DECIMAL(10,2)   NOT NULL DEFAULT '0.00' COMMENT '实付金额',
    `remark`                VARCHAR(500)             DEFAULT '' COMMENT '订单备注（用户留言）',
    `pay_time`              DATETIME                 DEFAULT NULL COMMENT '支付时间',
    `delivery_time`         DATETIME                 DEFAULT NULL COMMENT '发货时间',
    `receive_time`          DATETIME                 DEFAULT NULL COMMENT '确认收货时间',
    `cancel_time`           DATETIME                 DEFAULT NULL COMMENT '取消时间',
    `seckill_activity_id`   BIGINT UNSIGNED          DEFAULT '0' COMMENT '秒杀活动ID：普通订单为0',
    `seckill_price`         DECIMAL(10,2)            DEFAULT '0.00' COMMENT '秒杀价格快照',
    `created_at`            DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`            DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_order_sn` (`order_sn`),
    UNIQUE KEY `uk_idempotency_key` (`idempotency_key`),
    KEY `idx_user_id` (`user_id`),
    KEY `idx_status` (`status`),
    KEY `idx_order_type` (`order_type`),
    KEY `idx_seckill_activity_id` (`seckill_activity_id`),
    KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci
  COMMENT='订单主表';

-- =============================================
-- 订单商品表
-- =============================================
CREATE TABLE IF NOT EXISTS `order_item` (
    `id`           BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '订单商品ID',
    `order_id`     BIGINT UNSIGNED NOT NULL COMMENT '订单ID',
    `order_sn`     VARCHAR(64)     NOT NULL COMMENT '订单号',
    `sku_id`       BIGINT UNSIGNED NOT NULL COMMENT 'SKU ID',
    `spu_id`       BIGINT UNSIGNED NOT NULL COMMENT 'SPU ID',
    `sku_name`     VARCHAR(128)    NOT NULL DEFAULT '' COMMENT 'SKU名称快照',
    `sku_pic`      VARCHAR(500)    NOT NULL DEFAULT '' COMMENT 'SKU图片快照',
    `spec_data`    JSON                     DEFAULT NULL COMMENT '规格数据快照（JSON格式）',
    `price`        DECIMAL(10,2)   NOT NULL DEFAULT '0.00' COMMENT '成交单价',
    `count`        INT             NOT NULL DEFAULT '1' COMMENT '购买数量',
    `total_amount` DECIMAL(10,2)   NOT NULL DEFAULT '0.00' COMMENT '小计金额（单价×数量）',
    `is_seckill`   TINYINT         NOT NULL DEFAULT '0' COMMENT '是否秒杀商品：0-否，1-是',
    `created_at`   DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    PRIMARY KEY (`id`),
    KEY `idx_order_id` (`order_id`),
    KEY `idx_order_sn` (`order_sn`)
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci
  COMMENT='订单商品表';
