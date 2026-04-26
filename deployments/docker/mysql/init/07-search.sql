-- =============================================
-- 服务: 搜索服务 (Search Service)
-- 数据库: shop_search
-- 说明: 搜索历史、热搜词、搜索同步记录
-- 注意: 实际商品搜索建议使用 Elasticsearch
-- =============================================

CREATE DATABASE IF NOT EXISTS `shop_search`
CHARACTER SET utf8mb4
COLLATE utf8mb4_unicode_ci;

USE `shop_search`;

-- =============================================
-- 用户搜索历史表
-- =============================================
CREATE TABLE IF NOT EXISTS `search_history` (
    `id`           BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '历史记录ID',
    `user_id`      BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    `keyword`      VARCHAR(128)    NOT NULL COMMENT '搜索关键词',
    `result_count` INT                      DEFAULT '0' COMMENT '搜索结果数量',
    `created_at`   DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '搜索时间',
    PRIMARY KEY (`id`),
    KEY `idx_user_id` (`user_id`),
    KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci
  COMMENT='用户搜索历史表';

-- =============================================
-- 热搜词表
-- =============================================
CREATE TABLE IF NOT EXISTS `hot_keyword` (
    `id`           BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '热搜词ID',
    `keyword`      VARCHAR(128)    NOT NULL COMMENT '热搜关键词',
    `search_count` BIGINT          NOT NULL DEFAULT '1' COMMENT '搜索次数（累计）',
    `sort`         INT             NOT NULL DEFAULT '0' COMMENT '排序值：数值越小越靠前',
    `status`       TINYINT         NOT NULL DEFAULT '1' COMMENT '状态：1-显示，2-隐藏',
    `created_at`   DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`   DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_keyword` (`keyword`),
    KEY `idx_search_count` (`search_count`)
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci
  COMMENT='热搜词表';

-- =============================================
-- ES数据同步记录表
-- 说明: 用于记录哪些数据需要同步到Elasticsearch
-- =============================================
CREATE TABLE IF NOT EXISTS `es_sync_log` (
    `id`          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '同步记录ID',
    `table_name`  VARCHAR(64)     NOT NULL COMMENT '源数据表名：product_spu, product_sku等',
    `record_id`   BIGINT UNSIGNED NOT NULL COMMENT '源数据记录ID',
    `action`      VARCHAR(20)     NOT NULL COMMENT '操作类型：INSERT-新增，UPDATE-更新，DELETE-删除',
    `status`      TINYINT         NOT NULL DEFAULT '0' COMMENT '同步状态：0-待同步，1-已同步，2-同步失败',
    `retry_count` INT             NOT NULL DEFAULT '0' COMMENT '重试次数',
    `error_msg`   VARCHAR(500)             DEFAULT '' COMMENT '错误信息（同步失败时记录）',
    `created_at`  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    KEY `idx_status` (`status`),
    KEY `idx_table_record` (`table_name`, `record_id`)
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci
  COMMENT='ES数据同步记录表';
