CREATE TABLE `user` (
    `id` int unsigned NOT NULL AUTO_INCREMENT,
    `uid` bigint DEFAULT NULL,
    `nick_name` varchar(50) DEFAULT NULL,
    `avatar_uri` varchar(255) DEFAULT NULL,
    `reading_preference` tinyint NOT NULL DEFAULT '0',
    `create_time` datetime DEFAULT NULL,
    `update_time` datetime DEFAULT NULL,
    `auto_buy` tinyint(1) NOT NULL DEFAULT '1',
    `is_auto_buy` tinyint(1) NOT NULL DEFAULT '1',
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_user_uid` (`uid`),
    KEY `user_nick_name_IDX` (`nick_name`,`reading_preference`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;