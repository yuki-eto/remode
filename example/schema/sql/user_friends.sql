CREATE TABLE IF NOT EXISTS `user_friends` (
  `id` BIGINT(20) UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id` BIGINT(20) UNSIGNED NOT NULL,
  `other_user_id` BIGINT(20) UNSIGNED NOT NULL,
  `created_at` DATETIME,
  `updated_at` DATETIME,
  PRIMARY KEY (`id`),
  UNIQUE KEY `user_relation` (`user_id`, `other_user_id`),
  KEY `other_user_relation` (`other_user_id`)
);
