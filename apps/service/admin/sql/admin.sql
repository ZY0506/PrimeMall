-- =============================================
-- 服务: 后台管理服务 (Admin Service)
-- 数据库: shop_admin
-- 说明: 管理员账号、角色权限、操作日志、轮播图、公告等CMS内容
-- =============================================

CREATE DATABASE IF NOT EXISTS `shop_admin` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE `shop_admin`;

-- 管理员表
CREATE TABLE IF NOT EXISTS `admin` (
                                       `id` bigint unsigned NOT NULL AUTO_INCREMENT,
                                       `username` varchar(64) NOT NULL,
    `password` varchar(255) NOT NULL,
    `real_name` varchar(64) NOT NULL DEFAULT '',
    `avatar` varchar(500) DEFAULT '',
    `role_id` bigint unsigned NOT NULL DEFAULT '0',
    `status` tinyint NOT NULL DEFAULT '1' COMMENT '1:启用 2:禁用',
    `last_login_time` datetime DEFAULT NULL,
    `last_login_ip` varchar(64) DEFAULT '',
    `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_username` (`username`)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='管理员表';

-- 角色表
CREATE TABLE IF NOT EXISTS `role` (
                                      `id` bigint unsigned NOT NULL AUTO_INCREMENT,
                                      `name` varchar(64) NOT NULL,
    `permissions` json DEFAULT NULL COMMENT '权限列表',
    `remark` varchar(255) DEFAULT '',
    `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='角色表';

-- 操作日志表
CREATE TABLE IF NOT EXISTS `admin_log` (
                                           `id` bigint unsigned NOT NULL AUTO_INCREMENT,
                                           `admin_id` bigint unsigned NOT NULL,
                                           `username` varchar(64) NOT NULL,
    `module` varchar(64) NOT NULL COMMENT '操作模块',
    `action` varchar(64) NOT NULL COMMENT '操作类型',
    `request_method` varchar(10) DEFAULT '',
    `request_url` varchar(255) DEFAULT '',
    `request_params` text,
    `response_result` text,
    `ip` varchar(64) DEFAULT '',
    `duration_ms` int DEFAULT '0',
    `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_admin_id` (`admin_id`),
    KEY `idx_created_at` (`created_at`)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='管理员操作日志表';

-- 轮播图表 (CMS)
CREATE TABLE IF NOT EXISTS `banner` (
                                        `id` bigint unsigned NOT NULL AUTO_INCREMENT,
                                        `title` varchar(128) NOT NULL DEFAULT '',
    `image_url` varchar(500) NOT NULL,
    `link_url` varchar(500) DEFAULT '',
    `type` tinyint NOT NULL DEFAULT '1' COMMENT '1:商品 2:分类 3:活动页',
    `target_id` bigint unsigned DEFAULT '0',
    `sort` int NOT NULL DEFAULT '0',
    `platform` varchar(20) DEFAULT 'ALL' COMMENT 'ALL, PC, H5',
    `status` tinyint NOT NULL DEFAULT '1' COMMENT '1:启用 2:禁用',
    `start_time` datetime DEFAULT NULL,
    `end_time` datetime DEFAULT NULL,
    `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_status` (`status`)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='轮播图表';

-- 公告表 (CMS)
CREATE TABLE IF NOT EXISTS `notice` (
                                        `id` bigint unsigned NOT NULL AUTO_INCREMENT,
                                        `title` varchar(128) NOT NULL,
    `content` text NOT NULL,
    `type` tinyint NOT NULL DEFAULT '1' COMMENT '1:系统公告 2:活动公告',
    `priority` tinyint NOT NULL DEFAULT '1' COMMENT '1:普通 2:重要 3:置顶',
    `status` tinyint NOT NULL DEFAULT '1' COMMENT '1:发布 2:草稿 3:已下线',
    `publish_time` datetime DEFAULT NULL,
    `created_by` varchar(64) DEFAULT '',
    `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_status` (`status`)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='公告表';

-- 帮助中心分类表 (CMS)
CREATE TABLE IF NOT EXISTS `help_category` (
                                               `id` bigint unsigned NOT NULL AUTO_INCREMENT,
                                               `name` varchar(64) NOT NULL,
    `icon` varchar(255) DEFAULT '',
    `sort` int NOT NULL DEFAULT '0',
    `status` tinyint NOT NULL DEFAULT '1',
    `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='帮助中心分类表';

-- 帮助中心文章表 (CMS)
CREATE TABLE IF NOT EXISTS `help_article` (
                                              `id` bigint unsigned NOT NULL AUTO_INCREMENT,
                                              `category_id` bigint unsigned NOT NULL,
                                              `title` varchar(128) NOT NULL,
    `content` longtext NOT NULL,
    `keywords` varchar(255) DEFAULT '',
    `view_count` int NOT NULL DEFAULT '0',
    `sort` int NOT NULL DEFAULT '0',
    `status` tinyint NOT NULL DEFAULT '1' COMMENT '1:发布 2:草稿',
    `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_category_id` (`category_id`)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='帮助中心文章表';