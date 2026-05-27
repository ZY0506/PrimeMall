-- =============================================
-- 服务: 用户服务 (User Service)
-- 数据库: shop_user
-- 说明: 用户注册登录、个人信息、地址管理、风控
-- =============================================

-- goctl model mysql ddl --src=./apps/service/user/sql/01-user.sql --dir=./apps/service/user/rpc/internal/model

CREATE DATABASE IF NOT EXISTS `shop_user`
    CHARACTER SET utf8mb4
    COLLATE utf8mb4_unicode_ci;

USE `shop_user`;

-- =============================================
-- 用户主表
-- =============================================
CREATE TABLE IF NOT EXISTS `user` (
    `id`              BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    `phone`           VARCHAR(20)     NOT NULL DEFAULT '' COMMENT '手机号码',
    `password`        VARCHAR(255)    NOT NULL DEFAULT '' COMMENT '登录密码（加密）',
    `nickname`        VARCHAR(64)     NOT NULL DEFAULT '' COMMENT '用户昵称',
    `avatar`          VARCHAR(500)    NOT NULL DEFAULT '' COMMENT '头像URL',
    `gender`          TINYINT         NOT NULL DEFAULT '0' COMMENT '性别：0-未知，1-男，2-女',
    `birthday`        DATE                     DEFAULT NULL COMMENT '生日',
    `status`          TINYINT         NOT NULL DEFAULT '1' COMMENT '账号状态：1-正常，2-限制下单，3-封禁',
    `last_login_time` DATETIME                 DEFAULT NULL COMMENT '最后登录时间',
    `last_login_ip`   VARCHAR(64)              DEFAULT '' COMMENT '最后登录IP',
    `ext_info`        JSON                     DEFAULT NULL COMMENT '扩展信息（JSON格式）',
    `created_at`      DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`      DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `deleted_at`      DATETIME                 DEFAULT NULL COMMENT '删除时间（软删除）',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_phone` (`phone`),
    KEY `idx_status` (`status`)
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci
  COMMENT='用户主表';

-- =============================================
-- 用户地址表
-- =============================================
CREATE TABLE IF NOT EXISTS `user_address` (
    `id`              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '地址ID',
    `user_id`         BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    `tag`             VARCHAR(20)              DEFAULT 'HOME' COMMENT '地址标签：HOME-家，OFFICE-公司，SCHOOL-学校',
    `receiver_name`   VARCHAR(64)     NOT NULL DEFAULT '' COMMENT '收货人姓名',
    `receiver_phone`  VARCHAR(20)     NOT NULL DEFAULT '' COMMENT '收货人电话',
    `is_default`      TINYINT(1)      NOT NULL DEFAULT '0' COMMENT '是否默认地址：0-否，1-是',
    `info`            JSON            NOT NULL COMMENT '地址详细信息：{province-省份, city-城市, district-区县, detail-详细地址, postal_code-邮编}',
    `created_at`      DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`      DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    KEY `idx_user` (`user_id`)
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci
  COMMENT='用户地址表';

-- =============================================
-- 用户风控记录表
-- =============================================
CREATE TABLE IF NOT EXISTS `user_punish_log` (
    `id`          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '记录ID',
    `user_id`     BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '用户ID',
    `phone`       VARCHAR(20)     NOT NULL DEFAULT '' COMMENT '手机号码',
    `action_type` TINYINT         NOT NULL DEFAULT 0 COMMENT '处罚类型：1-禁止下单,2-禁止登录',
    `reason`      VARCHAR(255)    NOT NULL DEFAULT '' COMMENT '处罚原因',
    `banned_by`   VARCHAR(64)     NOT NULL DEFAULT 'SYSTEM' COMMENT '操作人：SYSTEM-系统自动，ADMIN-管理员(记录管理员账号)',
    `unbanned_by` VARCHAR(64)              DEFAULT NULL COMMENT '操作人：SYSTEM-系统自动，ADMIN-管理员',
    `start_time`  DATETIME                 DEFAULT NULL COMMENT '处罚开始时间',
    `end_time`    DATETIME                 DEFAULT NULL COMMENT '处罚结束时间',
    `created_at`  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    PRIMARY KEY (`id`),
    KEY `uk_phone` (`phone`),
    KEY `idx_user_id` (`user_id`)
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci
  COMMENT='用户风控记录表';

-- =============================================
-- 用户登录日志表
-- =============================================
CREATE TABLE IF NOT EXISTS `user_login_log` (
    `id`           BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '日志ID',
    `user_id`      BIGINT UNSIGNED NOT NULL DEFAULT '0' COMMENT '用户ID',
    `login_type`   VARCHAR(32)     NOT NULL DEFAULT '' COMMENT '登录方式：password-密码登录，captcha-短信登录',
    `login_ip`     VARCHAR(64)     NOT NULL DEFAULT '' COMMENT '登录IP地址',
    `user_agent`   VARCHAR(500)    NOT NULL DEFAULT '' COMMENT '用户代理（浏览器信息）',
    `status`       TINYINT         NOT NULL DEFAULT '1' COMMENT '登录状态：1-成功，2-失败',
    `fail_reason`  VARCHAR(255)             DEFAULT '' COMMENT '失败原因',
    `created_at`   DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '登录时间',
    PRIMARY KEY (`id`),
    KEY `idx_user_id` (`user_id`),
    KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci
  COMMENT='用户登录日志表';

-- =============================================
-- 索引优化（高并发场景）
-- =============================================
ALTER TABLE `user` ADD INDEX `idx_nickname` (`nickname`);
ALTER TABLE `user` ADD INDEX `idx_created_at` (`created_at`);
ALTER TABLE `user_address` ADD INDEX `idx_user_default` (`user_id`, `is_default`);
ALTER TABLE `user_punish_log` ADD INDEX `idx_user_action` (`user_id`, `action_type`);
ALTER TABLE `user_login_log` ADD INDEX `idx_user_created` (`user_id`, `created_at`);
