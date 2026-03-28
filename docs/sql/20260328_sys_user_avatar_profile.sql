ALTER TABLE `sys_user`
  ADD COLUMN `avatar_type` varchar(16) NULL COMMENT '头像类型',
  ADD COLUMN `avatar_color` varchar(16) NULL COMMENT '头像背景色';

UPDATE `sys_user`
SET `avatar_type` = 'image'
WHERE `avatar` <> ''
  AND (`avatar_type` IS NULL OR `avatar_type` = '');
