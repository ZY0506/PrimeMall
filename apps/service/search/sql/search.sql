-- =============================================
-- 服务: 搜索服务 (Search Service)
-- 数据库: shop_search
-- 说明: 搜索历史、热搜词、搜索同步记录
-- 注意: 实际商品搜索建议使用 Elasticsearch
-- =============================================

CREATE DATABASE IF NOT EXISTS `shop_search` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE `shop_search`;

-- 用户搜索历史表
CREATE TABLE IF NOT EXISTS `search_history` (
                                                `id` bigint unsigned NOT NULL AUTO_INCREMENT,
                                                `user_id` bigint unsigned NOT NULL,
                                                `keyword` varchar(128) NOT NULL,
    `result_count` int DEFAULT '0',
    `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_user_id` (`user_id`),
    KEY `idx_created_at` (`created_at`)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户搜索历史表';

-- 热搜词表
CREATE TABLE IF NOT EXISTS `hot_keyword` (
                                             `id` bigint unsigned NOT NULL AUTO_INCREMENT,
                                             `keyword` varchar(128) NOT NULL,
    `search_count` bigint NOT NULL DEFAULT '1',
    `sort` int NOT NULL DEFAULT '0',
    `status` tinyint NOT NULL DEFAULT '1' COMMENT '1:显示 2:隐藏',
    `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_keyword` (`keyword`),
    KEY `idx_search_count` (`search_count`)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='热搜词表';

-- ES数据同步记录表 (用于记录哪些数据需要同步到ES)
CREATE TABLE IF NOT EXISTS `es_sync_log` (
                                             `id` bigint unsigned NOT NULL AUTO_INCREMENT,
                                             `table_name` varchar(64) NOT NULL COMMENT '表名: product_spu, product_sku',
    `record_id` bigint unsigned NOT NULL COMMENT '记录ID',
    `action` varchar(20) NOT NULL COMMENT 'INSERT, UPDATE, DELETE',
    `status` tinyint NOT NULL DEFAULT '0' COMMENT '0:待同步 1:已同步 2:同步失败',
    `retry_count` int NOT NULL DEFAULT '0',
    `error_msg` varchar(500) DEFAULT '',
    `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_status` (`status`),
    KEY `idx_table_record` (`table_name`, `record_id`)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='ES数据同步记录表';