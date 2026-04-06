-- =============================================
-- 服务: 用户服务 (User Service)
-- 数据库: shop_user
-- 说明: 用户注册登录、个人信息、地址管理、风控
-- =============================================

CREATE DATABASE IF NOT EXISTS `shop_user` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE `shop_user`;

-- 用户主表
CREATE TABLE IF NOT EXISTS `user` (
                                      `id` bigint unsigned NOT NULL AUTO_INCREMENT,
                                      `phone` varchar(20) NOT NULL DEFAULT '',
    `password` varchar(255) NOT NULL DEFAULT '',
    `nickname` varchar(64) NOT NULL DEFAULT '',
    `avatar` varchar(500) NOT NULL DEFAULT '',
    `gender` tinyint NOT NULL DEFAULT '0' COMMENT '0:未知 1:男 2:女',
    `birthday` date DEFAULT NULL,
    `status` tinyint NOT NULL DEFAULT '1' COMMENT '1:正常 2:限制下单 3:封禁',
    `last_login_time` datetime DEFAULT NULL,
    `last_login_ip` varchar(64) DEFAULT '',
    `ext_info` json DEFAULT NULL,
    `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at` datetime DEFAULT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_phone` (`phone`),
    KEY `idx_status` (`status`)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户主表';

-- 用户地址表
CREATE TABLE IF NOT EXISTS `user_address` (
                                              `id` bigint unsigned NOT NULL AUTO_INCREMENT,
                                              `user_id` bigint unsigned NOT NULL,
                                              `tag` varchar(20) DEFAULT 'HOME' COMMENT 'HOME, OFFICE, SCHOOL',
    `receiver_name` varchar(64) NOT NULL DEFAULT '',
    `receiver_phone` varchar(20) NOT NULL DEFAULT '',
    `is_default` tinyint(1) NOT NULL DEFAULT '0',
    `info` json NOT NULL COMMENT '{province, city, district, detail, postal_code}',
    `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_user` (`user_id`)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户地址表';

-- 用户风控记录表
CREATE TABLE IF NOT EXISTS `user_punish_log` (
                                                 `id` bigint unsigned NOT NULL AUTO_INCREMENT,
                                                 `user_id` bigint unsigned NOT NULL DEFAULT '0',
                                                 `action_type` varchar(32) NOT NULL DEFAULT '' COMMENT 'BAN_LOGIN, BAN_ORDER',
    `reason` varchar(255) NOT NULL DEFAULT '',
    `operator` varchar(64) NOT NULL DEFAULT 'SYSTEM',
    `start_time` datetime DEFAULT NULL,
    `end_time` datetime DEFAULT NULL,
    `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_user_id` (`user_id`)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户风控记录表';

-- 用户登录日志表
CREATE TABLE IF NOT EXISTS `user_login_log` (
                                                `id` bigint unsigned NOT NULL AUTO_INCREMENT,
                                                `user_id` bigint unsigned NOT NULL DEFAULT '0',
                                                `login_type` varchar(32) NOT NULL DEFAULT '' COMMENT 'password, sms, wechat',
    `login_ip` varchar(64) NOT NULL DEFAULT '',
    `user_agent` varchar(500) NOT NULL DEFAULT '',
    `status` tinyint NOT NULL DEFAULT '1' COMMENT '1:成功 2:失败',
    `fail_reason` varchar(255) DEFAULT '',
    `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_user_id` (`user_id`),
    KEY `idx_created_at` (`created_at`)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户登录日志表';