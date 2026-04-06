-- =============================================
-- 服务: 商品服务 (Product Service)
-- 数据库: shop_product
-- 说明: 商品分类、SPU、SKU、库存管理
-- =============================================

CREATE DATABASE IF NOT EXISTS `shop_product` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE `shop_product`;

-- 商品分类表
CREATE TABLE IF NOT EXISTS `category` (
                                          `id` bigint unsigned NOT NULL AUTO_INCREMENT,
                                          `parent_id` bigint unsigned NOT NULL DEFAULT '0',
                                          `name` varchar(64) NOT NULL DEFAULT '',
    `icon` varchar(500) NOT NULL DEFAULT '',
    `sort` int NOT NULL DEFAULT '0',
    `level` tinyint NOT NULL DEFAULT '1',
    `status` tinyint NOT NULL DEFAULT '1' COMMENT '1:启用 2:禁用',
    `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_parent_id` (`parent_id`)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='商品分类表';

-- 商品SPU表
CREATE TABLE IF NOT EXISTS `product_spu` (
                                             `id` bigint unsigned NOT NULL AUTO_INCREMENT,
                                             `category_id` bigint unsigned NOT NULL DEFAULT '0',
                                             `name` varchar(128) NOT NULL DEFAULT '',
    `brand` varchar(64) NOT NULL DEFAULT '',
    `desc` varchar(255) NOT NULL DEFAULT '',
    `content` longtext,
    `main_pic` varchar(500) NOT NULL DEFAULT '',
    `sub_pics` json DEFAULT NULL,
    `video_url` varchar(500) NOT NULL DEFAULT '',
    `status` tinyint NOT NULL DEFAULT '1' COMMENT '1:上架 2:下架',
    `sales_count` int NOT NULL DEFAULT '0',
    `virtual_sales` int NOT NULL DEFAULT '0',
    `weight` decimal(10,2) DEFAULT '0.00',
    `freight_template_id` bigint unsigned DEFAULT '0',
    `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_category_id` (`category_id`),
    KEY `idx_status` (`status`)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='商品SPU表';

-- 商品SKU表
CREATE TABLE IF NOT EXISTS `product_sku` (
                                             `id` bigint unsigned NOT NULL AUTO_INCREMENT,
                                             `spu_id` bigint unsigned NOT NULL,
                                             `sku_code` varchar(64) NOT NULL DEFAULT '',
    `price` decimal(10,2) NOT NULL,
    `market_price` decimal(10,2) NOT NULL,
    `cost_price` decimal(10,2) NOT NULL DEFAULT '0.00',
    `stock` int NOT NULL DEFAULT '0',
    `version` int unsigned NOT NULL DEFAULT '0' COMMENT '乐观锁',
    `spec_data` json NOT NULL,
    `images` json DEFAULT NULL,
    `weight` decimal(10,2) DEFAULT '0.00',
    `status` tinyint NOT NULL DEFAULT '1' COMMENT '1:启用 2:禁用',
    `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_sku_code` (`sku_code`),
    KEY `idx_spu_id` (`spu_id`),
    KEY `idx_price` (`price`)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='商品SKU表';

-- 库存变动流水表
CREATE TABLE IF NOT EXISTS `stock_log` (
                                           `id` bigint unsigned NOT NULL AUTO_INCREMENT,
                                           `sku_id` bigint unsigned NOT NULL,
                                           `order_sn` varchar(64) NOT NULL,
    `change_type` tinyint NOT NULL COMMENT '1:预扣 2:回滚 3:实际扣减 4:退货加回',
    `quantity` int NOT NULL,
    `before_stock` int NOT NULL DEFAULT '0',
    `after_stock` int NOT NULL DEFAULT '0',
    `remark` varchar(255) DEFAULT '',
    `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_order_sku_type` (`order_sn`, `sku_id`, `change_type`),
    KEY `idx_sku_id` (`sku_id`)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='库存变动流水表';

-- 运费模板表
CREATE TABLE IF NOT EXISTS `freight_template` (
                                                  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
                                                  `name` varchar(64) NOT NULL DEFAULT '',
    `type` tinyint NOT NULL DEFAULT '1' COMMENT '1:按件数 2:按重量',
    `default_fee` decimal(10,2) NOT NULL DEFAULT '0.00',
    `default_quantity` int NOT NULL DEFAULT '1',
    `extra_fee` decimal(10,2) NOT NULL DEFAULT '0.00',
    `free_threshold_amount` decimal(10,2) DEFAULT '0.00',
    `free_threshold_quantity` int DEFAULT '0',
    `is_default` tinyint NOT NULL DEFAULT '0',
    `status` tinyint NOT NULL DEFAULT '1',
    `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='运费模板表';