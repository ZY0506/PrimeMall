-- =============================================
-- 服务: 商品服务 (Product Service)
-- 数据库: shop_product
-- 说明: 商品分类、SPU、SKU、库存管理
-- 金额字段统一使用 BIGINT 类型，单位为分（避免浮点数精度问题）
-- =============================================
-- goctl model mysql ddl --src=./apps/service/product/sql/02-product.sql --dir=./apps/service/product/rpc/internal/model
CREATE DATABASE IF NOT EXISTS `shop_product`
CHARACTER SET utf8mb4
COLLATE utf8mb4_unicode_ci;

USE `shop_product`;

-- =============================================
-- 商品分类表
-- =============================================
CREATE TABLE IF NOT EXISTS `category` (
    `id`          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '分类ID',
    `parent_id`   BIGINT UNSIGNED NOT NULL DEFAULT '0' COMMENT '父分类ID：0-顶级分类',
    `name`        VARCHAR(64)     NOT NULL DEFAULT '' COMMENT '分类名称',
    `icon`        VARCHAR(500)    NOT NULL DEFAULT '' COMMENT '分类图标URL',
    `sort`        INT             NOT NULL DEFAULT '0' COMMENT '排序值：数值越小越靠前',
    `level`       TINYINT         NOT NULL DEFAULT '1' COMMENT '分类层级：1-一级分类，2-二级分类，3-三级分类',
    `status`      TINYINT         NOT NULL DEFAULT '1' COMMENT '状态：1-启用，2-禁用',
    `created_at`  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `delete_at`   DATETIME                 DEFAULT NULL COMMENT '删除时间',
    PRIMARY KEY (`id`),
    KEY `idx_parent_id` (`parent_id`)
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci
  COMMENT='商品分类表';

-- =============================================
-- 商品SPU表（标准产品单位）
-- =============================================
CREATE TABLE IF NOT EXISTS `product_spu` (
    `id`                  BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'SPU ID',
    `category_id`         BIGINT UNSIGNED NOT NULL DEFAULT '0' COMMENT '所属分类ID',
    `name`                VARCHAR(128)    NOT NULL DEFAULT '' COMMENT '商品名称',
    `brand`               VARCHAR(64)     NOT NULL DEFAULT '' COMMENT '品牌名称',
    `desc`                VARCHAR(255)    NOT NULL DEFAULT '' COMMENT '商品简短描述',
    `content`             LONGTEXT                 DEFAULT NULL COMMENT '商品详情内容（HTML）',
    `main_pic`            VARCHAR(500)    NOT NULL DEFAULT '' COMMENT '主图URL',
    `sub_pics`            JSON                     DEFAULT NULL COMMENT '副图列表（JSON数组）',
    `video_url`           VARCHAR(500)    NOT NULL DEFAULT '' COMMENT '商品视频URL',
    `status`              TINYINT         NOT NULL DEFAULT '1' COMMENT '上架状态：1-上架，2-下架',
    `sales_count`         INT             NOT NULL DEFAULT '0' COMMENT '实际销量',
    `virtual_sales`       INT             NOT NULL DEFAULT '0' COMMENT '虚拟销量（用于展示）',
    `freight_template_id` BIGINT UNSIGNED          DEFAULT '0' COMMENT '运费模板ID：0-使用默认模板',
    `created_at`          DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`          DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `delete_at`           DATETIME                 DEFAULT NULL COMMENT '删除时间',
    PRIMARY KEY (`id`),
    KEY `idx_category_id` (`category_id`),
    KEY `idx_status` (`status`)
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci
  COMMENT='商品SPU表';

-- =============================================
-- 商品SKU表（库存量单位）
-- =============================================
CREATE TABLE IF NOT EXISTS `product_sku` (
    `id`           BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'SKU ID',
    `spu_id`       BIGINT UNSIGNED NOT NULL COMMENT '所属SPU ID',
    `spu_sname`    VARCHAR(128)    NOT NULL DEFAULT '' COMMENT 'SPU名称',
    `sku_code`     VARCHAR(64)     NOT NULL DEFAULT '' COMMENT 'SKU编码（唯一标识）',
    `price`        BIGINT          NOT NULL COMMENT '销售价格',
    `market_price` BIGINT          NOT NULL COMMENT '市场价（划线价）',
    `cost_price`   BIGINT          NOT NULL DEFAULT '0' COMMENT '成本价',
    `stock`        INT             NOT NULL DEFAULT '0' COMMENT '库存数量',
    `locked_stock` INT             NOT NULL DEFAULT '0' COMMENT '锁定库存数量',
    `version`      INT UNSIGNED    NOT NULL DEFAULT '0' COMMENT '乐观锁版本号（用于并发控制）',
    `spec_data`    JSON            NOT NULL COMMENT '规格数据（JSON格式）：{颜色:红色},{尺寸:XL}',
    `images`       JSON                     DEFAULT NULL COMMENT 'SKU专属图片列表（JSON数组）',
    `weight`       BIGINT                   DEFAULT '0' COMMENT 'SKU重量（kg）',
    `status`       TINYINT         NOT NULL DEFAULT '1' COMMENT '状态：1-启用，2-禁用',
    `created_at`   DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`   DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `deleted_at`   DATETIME                 DEFAULT NULL COMMENT '删除时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_sku_code` (`sku_code`),
    KEY `idx_spu_id` (`spu_id`),
    KEY `idx_price` (`price`)
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci
  COMMENT='商品SKU表';

-- =============================================
-- 库存变动流水表
-- =============================================
CREATE TABLE IF NOT EXISTS `stock_log` (
    `id`           BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '流水ID',
    `sku_id`       BIGINT UNSIGNED NOT NULL COMMENT 'SKU ID',
    `order_sn`     VARCHAR(64)     NOT NULL COMMENT '订单号',
    `change_type`  TINYINT         NOT NULL COMMENT '变动类型：1-锁定库存，2-解锁库存，3-实际扣减，4-回滚库存(退货加回), 5-恢复库存(下单事务失败时回滚)',
    `quantity`     INT             NOT NULL COMMENT '变动数量（正数增加，负数减少）',
    `before_stock` INT             NOT NULL DEFAULT '0' COMMENT '变动前库存',
    `after_stock`  INT             NOT NULL DEFAULT '0' COMMENT '变动后库存',
    `remark`       VARCHAR(255)             DEFAULT '' COMMENT '备注说明',
    `created_at`   DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_order_sku_type` (`order_sn`, `sku_id`, `change_type`),
    KEY `idx_sku_id` (`sku_id`)
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci
  COMMENT='库存变动流水表';

-- =============================================
-- 运费模板表
-- =============================================
CREATE TABLE IF NOT EXISTS `freight_template` (
    `id`                        BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '模板ID',
    `name`                      VARCHAR(64)     NOT NULL DEFAULT '' COMMENT '模板名称',
    `type`                      TINYINT         NOT NULL DEFAULT '1' COMMENT '计费方式：1-按件数，2-按重量',
    `default_fee`               BIGINT          NOT NULL DEFAULT '0' COMMENT '基础运费',
    `default_quantity`          INT             NOT NULL DEFAULT '1' COMMENT '基础数量（首件/首重）',
    `extra_fee`                 BIGINT          NOT NULL DEFAULT '0' COMMENT '续件/续重费用',
    `free_threshold_amount`     BIGINT                   DEFAULT '0' COMMENT '满额包邮阈值：0-不包邮',
    `free_threshold_quantity`   INT                      DEFAULT '0' COMMENT '满件包邮阈值：0-不包邮',
    `is_default`                TINYINT         NOT NULL DEFAULT '0' COMMENT '是否默认模板：0-否，1-是',
    `status`                    TINYINT         NOT NULL DEFAULT '1' COMMENT '状态：1-启用，2-禁用',
    `created_at`                DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`                DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `delete_at`                 DATETIME                 DEFAULT NULL COMMENT '删除时间',
    PRIMARY KEY (`id`)
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci
  COMMENT='运费模板表';
