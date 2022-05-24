CREATE TABLE IF NOT EXISTS `v_test_table` (
    `id` BIGINT(20) UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'primary key',
    `student_name` VARCHAR(128) NOT NULL COMMENT 'student name',
    `created_at` TIMESTAMP NOT NULL CURRENT_STAMP ON UPDATE CURRENT_STAMP,
    PRIMARY KEY `id`,
    KEY `idx_name` (`student_name`)
) ENGINE=InnoDB COMMENT='test';