-- =============================================
-- 服务: 后台管理服务 (Admin Service)
-- 数据库: shop_admin
-- 说明: 管理员账号、角色权限、操作日志、轮播图、公告等CMS内容
-- =============================================

-- goctl model mysql ddl --src=./apps/service/admin/sql/06-admin.sql --dir=./apps/service/admin/rpc/internal/model

CREATE DATABASE IF NOT EXISTS `shop_admin`
    CHARACTER SET utf8mb4
    COLLATE utf8mb4_unicode_ci;

USE `shop_admin`;

-- =============================================
-- 管理员表
-- =============================================
CREATE TABLE IF NOT EXISTS `admin` (
    `id`              BIGINT UNSIGNED NOT NULL COMMENT '管理员ID',
    `username`        VARCHAR(64)     NOT NULL COMMENT '登录用户名',
    `password`        VARCHAR(255)    NOT NULL COMMENT '登录密码（加密）',
    `real_name`       VARCHAR(64)     NOT NULL DEFAULT '' COMMENT '真实姓名',
    `avatar`          VARCHAR(500)             DEFAULT '' COMMENT '头像URL',
    `role_id`         BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '角色ID：0-无角色',
    `status`          TINYINT         NOT NULL DEFAULT 1 COMMENT '状态：1-启用，2-禁用',
    `last_login_time` DATETIME                 DEFAULT NULL COMMENT '最后登录时间',
    `last_login_ip`   VARCHAR(64)              DEFAULT '' COMMENT '最后登录IP',
    `created_at`      DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`      DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_username` (`username`)
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci
  COMMENT='管理员表';

-- =============================================
-- 角色表
-- =============================================
CREATE TABLE IF NOT EXISTS `role` (
    `id`          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '角色ID',
    `name`        VARCHAR(64)     NOT NULL COMMENT '角色名称',
    `code`        VARCHAR(64)     NOT NULL COMMENT '角色标识',
    `remark`      VARCHAR(255)             DEFAULT '' COMMENT '备注说明',
    `created_at`  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`)
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci
  COMMENT='角色表';

-- =============================================
-- 权限表
-- =============================================
CREATE TABLE IF NOT EXISTS `permission` (
    `id`          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `name`        VARCHAR(64)     NOT NULL COMMENT '权限名称',
    `code`        VARCHAR(128)    NOT NULL COMMENT '权限标识',
    `module`      VARCHAR(64)              DEFAULT '' COMMENT '所属模块',
    `created_at`  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_code` (`code`)
)ENGINE=InnoDB
DEFAULT CHARSET=utf8mb4
COLLATE=utf8mb4_unicode_ci
COMMENT='权限表';

-- =============================================
-- 角色权限表
-- =============================================
CREATE TABLE IF NOT EXISTS `role_permission` (
    `id`             BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `role_id`        BIGINT UNSIGNED NOT NULL,
    `permission_id`  BIGINT UNSIGNED NOT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_role_permission` (`role_id`, `permission_id`)
)ENGINE=InnoDB
DEFAULT CHARSET=utf8mb4
COLLATE=utf8mb4_unicode_ci
COMMENT='角色权限表';

-- =============================================
-- 管理员操作日志表
-- =============================================
CREATE TABLE IF NOT EXISTS `admin_log` (
    `id`              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '日志ID',
    `admin_id`        BIGINT UNSIGNED NOT NULL COMMENT '管理员ID',
    `username`        VARCHAR(64)     NOT NULL COMMENT '管理员用户名',
    `module`          VARCHAR(64)     NOT NULL COMMENT '操作模块：用户管理、订单管理、商品管理等',
    `action`          VARCHAR(64)     NOT NULL COMMENT '操作类型：新增、编辑、删除、审核等',
    `request_method`  VARCHAR(10)              DEFAULT '' COMMENT '请求方法：GET、POST、PUT、DELETE',
    `request_url`     VARCHAR(255)             DEFAULT '' COMMENT '请求URL',
    `request_params`  TEXT                     DEFAULT NULL COMMENT '请求参数（JSON格式）',
    `response_result` TEXT                     DEFAULT NULL COMMENT '响应结果（JSON格式）',
    `ip`              VARCHAR(64)              DEFAULT '' COMMENT '操作IP地址',
    `duration_ms`     INT                      DEFAULT '0' COMMENT '请求耗时（毫秒）',
    `created_at`      DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '操作时间',
    PRIMARY KEY (`id`),
    KEY `idx_admin_id` (`admin_id`),
    KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci
  COMMENT='管理员操作日志表';

-- =============================================
-- 轮播图表 (CMS)
-- =============================================
CREATE TABLE IF NOT EXISTS `banner` (
    `id`          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '轮播图ID',
    `title`       VARCHAR(128)    NOT NULL DEFAULT '' COMMENT '轮播图标题',
    `image_url`   VARCHAR(500)    NOT NULL COMMENT '图片URL',
    `link_url`    VARCHAR(500)             DEFAULT '' COMMENT '跳转链接URL',
    `type`        TINYINT         NOT NULL DEFAULT '1' COMMENT '跳转类型：1-商品详情，2-分类页面，3-活动页面',
    `target_id`   BIGINT UNSIGNED          DEFAULT '0' COMMENT '目标ID：对应商品ID/分类ID/活动ID',
    `sort`        INT             NOT NULL DEFAULT '0' COMMENT '排序值：数值越小越靠前',
    `platform`    VARCHAR(20)              DEFAULT 'ALL' COMMENT '展示平台：ALL-全部，PC-电脑端，H5-移动端',
    `status`      TINYINT         NOT NULL DEFAULT '1' COMMENT '状态：1-启用，2-禁用',
    `start_time`  DATETIME                 DEFAULT NULL COMMENT '生效开始时间',
    `end_time`    DATETIME                 DEFAULT NULL COMMENT '生效结束时间',
    `created_at`  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    KEY `idx_status` (`status`)
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci
  COMMENT='轮播图表';

-- =============================================
-- 公告表 (CMS)
-- =============================================
CREATE TABLE IF NOT EXISTS `notice` (
    `id`            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '公告ID',
    `title`         VARCHAR(128)    NOT NULL COMMENT '公告标题',
    `content`       TEXT            NOT NULL COMMENT '公告内容（支持HTML）',
    `type`          TINYINT         NOT NULL DEFAULT '1' COMMENT '公告类型：1-系统公告，2-活动公告',
    `priority`      TINYINT         NOT NULL DEFAULT '1' COMMENT '优先级：1-普通，2-重要，3-置顶',
    `status`        TINYINT         NOT NULL DEFAULT '1' COMMENT '状态：1-已发布，2-草稿，3-已下线',
    `publish_time`  DATETIME                 DEFAULT NULL COMMENT '发布时间',
    `created_by`    VARCHAR(64)              DEFAULT '' COMMENT '创建人（管理员用户名）',
    `created_at`    DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`    DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    KEY `idx_status` (`status`)
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci
  COMMENT='公告表';

-- =============================================
-- 帮助中心分类表 (CMS)
-- =============================================
CREATE TABLE IF NOT EXISTS `help_category` (
    `id`         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '分类ID',
    `name`       VARCHAR(64)     NOT NULL COMMENT '分类名称',
    `icon`       VARCHAR(255)             DEFAULT '' COMMENT '分类图标URL',
    `sort`       INT             NOT NULL DEFAULT '0' COMMENT '排序值：数值越小越靠前',
    `status`     TINYINT         NOT NULL DEFAULT '1' COMMENT '状态：1-启用，2-禁用',
    `created_at` DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    PRIMARY KEY (`id`)
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci
  COMMENT='帮助中心分类表';

-- =============================================
-- 帮助中心文章表 (CMS)
-- =============================================
CREATE TABLE IF NOT EXISTS `help_article` (
    `id`          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '文章ID',
    `category_id` BIGINT UNSIGNED NOT NULL COMMENT '所属分类ID',
    `title`       VARCHAR(128)    NOT NULL COMMENT '文章标题',
    `content`     LONGTEXT        NOT NULL COMMENT '文章内容（支持HTML）',
    `keywords`    VARCHAR(255)             DEFAULT '' COMMENT '关键词（用于搜索优化）',
    `view_count`  INT             NOT NULL DEFAULT '0' COMMENT '浏览次数',
    `sort`        INT             NOT NULL DEFAULT '0' COMMENT '排序值：数值越小越靠前',
    `status`      TINYINT         NOT NULL DEFAULT '1' COMMENT '状态：1-已发布，2-草稿',
    `created_at`  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    KEY `idx_category_id` (`category_id`)
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci
  COMMENT='帮助中心文章表';

-- =============================================
-- 索引优化（高并发场景）
-- =============================================
ALTER TABLE `admin` ADD INDEX `idx_role_id` (`role_id`);
ALTER TABLE `admin` ADD INDEX `idx_status` (`status`);
ALTER TABLE `role` ADD UNIQUE INDEX `uk_code` (`code`);
ALTER TABLE `role_permission` ADD INDEX `idx_permission_id` (`permission_id`);
ALTER TABLE `admin_log` ADD INDEX `idx_admin_created` (`admin_id`, `created_at`);
ALTER TABLE `admin_log` ADD INDEX `idx_module` (`module`);
ALTER TABLE `banner` ADD INDEX `idx_status_platform` (`status`, `type`, `platform`);
ALTER TABLE `banner` ADD INDEX `idx_status_time` (`status`, `start_time`, `end_time`);
ALTER TABLE `notice` ADD INDEX `idx_type_status` (`type`, `status`);
ALTER TABLE `notice` ADD INDEX `idx_status_publish` (`status`, `publish_time`);
ALTER TABLE `help_category` ADD INDEX `idx_status_sort` (`status`, `sort`);
ALTER TABLE `help_article` ADD INDEX `idx_category_status` (`category_id`, `status`);
