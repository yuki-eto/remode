CREATE TABLE IF NOT EXISTS `users` (
    `id` BIGINT(20) UNSIGNED NOT NULL AUTO_INCREMENT,
    `uuid` VARCHAR(127) COLLATE utf8mb4_unicode_ci NOT NULL,
    `access_token` VARCHAR(127) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
    `outside_user_id` VARCHAR(127) COLLATE utf8mb4_unicode_ci NOT NULL,
    `name` VARCHAR(127) COLLATE utf8mb4_unicode_ci NOT NULL,
    `created_at` DATETIME,
    `updated_at` DATETIME,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uuid` (`uuid`),
    UNIQUE KEY `outside_user_id` (`outside_user_id`)
);
