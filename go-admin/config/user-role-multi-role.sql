CREATE TABLE IF NOT EXISTS `sys_user_role` (
  `user_id` bigint NOT NULL COMMENT '用户ID',
  `role_id` bigint NOT NULL COMMENT '角色ID',
  `is_primary` tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否主角色',
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`user_id`,`role_id`),
  KEY `idx_sys_user_role_role_id` (`role_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户角色关联表';

INSERT INTO `sys_user_role` (`user_id`, `role_id`, `is_primary`, `created_at`, `updated_at`)
SELECT `user_id`, `role_id`, 1, NOW(3), NOW(3)
FROM `sys_user`
WHERE `role_id` IS NOT NULL AND `role_id` > 0
ON DUPLICATE KEY UPDATE
  `is_primary` = VALUES(`is_primary`),
  `updated_at` = VALUES(`updated_at`);
