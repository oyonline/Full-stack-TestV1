-- ========================================================  
-- 公告管理模块 - MySQL 数据库表结构
-- 表名: sys_notice
-- 对应 Model: go-admin/app/admin/models/sys_notice.go
-- 执行: mysql -u root -p enterprise_admin_starter < sys_notice_mysql.sql
-- ========================================================

CREATE TABLE IF NOT EXISTS `sys_notice` (
    -- 主键
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键编码',
    
    -- 业务字段
    `title` VARCHAR(256) NOT NULL DEFAULT '' COMMENT '公告标题',
    `content` TEXT COMMENT '公告内容（富文本）',
    `type` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '类型：1-通知 2-公告',
    `status` VARCHAR(4) NOT NULL DEFAULT '1' COMMENT '状态：0-禁用 1-启用',
    `sort` INT NOT NULL DEFAULT 0 COMMENT '排序权重，数值越小越靠前',
    `remark` VARCHAR(256) DEFAULT '' COMMENT '备注',
    
    -- 审计字段
    `create_by` INT UNSIGNED DEFAULT 0 COMMENT '创建者',
    `update_by` INT UNSIGNED DEFAULT 0 COMMENT '更新者',
    
    -- 时间字段
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后更新时间',
    `deleted_at` DATETIME DEFAULT NULL COMMENT '删除时间（软删除）',
    
    -- 主键约束
    PRIMARY KEY (`id`),
    
    -- 索引
    KEY `idx_status` (`status`),
    KEY `idx_type` (`type`),
    KEY `idx_sort` (`sort`),
    KEY `idx_create_by` (`create_by`),
    KEY `idx_deleted_at` (`deleted_at`)
    
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='公告管理表-存储系统公告和通知信息';

-- 插入测试数据
INSERT INTO `sys_notice` (`title`, `content`, `type`, `status`, `sort`, `remark`, `create_by`, `update_by`) 
VALUES 
('系统维护通知', '系统将于今晚 12:00 进行维护升级，请提前保存工作。', '1', '1', 1, '维护公告', 1, 1),
('新功能上线公告', '公告管理功能已正式上线，欢迎使用！', '2', '1', 2, '功能上线', 1, 1),
('清明节放假通知', '2026年清明节放假安排通知。', '1', '1', 3, '节假日通知', 1, 1);

-- 验证数据
SELECT * FROM `sys_notice` ORDER BY `sort`;
