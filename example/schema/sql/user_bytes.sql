CREATE TABLE IF NOT EXISTS `user_bytes` (
    `id` BIGINT(20) UNSIGNED NOT NULL,
    `user_id` BIGINT(20) UNSIGNED NOT NULL,
    `bytes` binary,
    `tags` set('one', 'two', 'three') NOT NULL,
    `created_at` DATETIME,
    `updated_at` DATETIME,
    PRIMARY KEY (`id`),
    UNIQUE KEY `user_id` (`user_id`)
);