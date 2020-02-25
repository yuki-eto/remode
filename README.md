# remodel
generate domain codes from create table ddl

# requirement
- Go 1.13+
- MySQL
- Memcached

# Usage
## install remodel command
```
go get -u github.com/yuki-eto/remodel/cmd/remodel
```

## write create table ddl
make table schema to `(root_dir)/schema/sql`

```
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
```

## to Yaml
run cli and write to `(root_dir)/schema/yaml`

```
remodel -root ./ yaml
```

## to Golang codes
```
remodel -root ./ -module module_sample entity
remodel -root ./ -module module_sample dao
remodel -root ./ -module module_sample model
```
