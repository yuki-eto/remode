CREATE TABLE IF NOT EXISTS items (
    `id` BIGINT(20) UNSIGNED NOT NULL,
    `type` ENUM('consumable', 'important') COLLATE utf8mb4_unicode_ci NOT NULL,
    `rarity` ENUM('R', 'SR', 'SSR') COLLATE utf8mb4_unicode_ci NOT NULL,
    `name` VARCHAR(255) COLLATE utf8mb4_unicode_ci NOT NULL,
    `max_count` SMALLINT(5) UNSIGNED NOT NULL,
    PRIMARY KEY (`id`),
    KEY `type` (`type`),
    KEY `rarity` (`rarity`)
);
