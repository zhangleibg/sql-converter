In routine programming using golang , it often takes a long time to construct persistent object `po struct` line by line when given the create sql. The repository provides a out-of-the-box script to automically generate `po`  when taken one sql file containing create statement as input.

The usage is detailed as follows:

### 1. install the script
```
go get github.com/zhangleibg/sql-converter
```
the command will download the repository as a go module, and output a executable file into the `bin` folder of `GOENV`. Note: the `$GOENV$/bin` should be included in the os environment variable `PATH` before using the script

### 2. use it
The usage of script is as follows:
```
Usage: sql.converter <path> [-tags=<tags>] [-comment_tag=<comment_tag>] [-table_prefix=<table_prefix>] [-table_suffix=<table_suffix>] [-field_prefix=<field_prefix>] [-field_suffix=<field_suffix>] [-h] [-target=<target>] [-mode=<mode>]
Param:
	path: 			the sql file
	-tags: 			field tag, default: "json,db",
	-comment_tag: 	comment tag, default: "comment",
	-table_prefix: 	the prefix of table name,
	-table_suffix: 	the suffix of table name,
	-field_prefix: 	the prefix of field name,
	-field_prefix: 	the suffix of field name,
	-h: 			the hint for usage,
	-target: 		the directory of generated go file
```

### 3. example
test.sql
```
CREATE TABLE IF NOT EXISTS `v_test_table` (
    `id` BIGINT(20) UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'primary key',
    `student_name` VARCHAR(128) NOT NULL COMMENT 'student name',
    `created_at` TIMESTAMP NOT NULL CURRENT_STAMP ON UPDATE CURRENT_STAMP,
    PRIMARY KEY `id`,
    KEY `idx_name` (`student_name`)
) ENGINE=InnoDB COMMENT='test';
```
when input the command
```
sql-converter ./test.sql -table_prefix=v_ -tags=db,json
```
the `generator.go` will be output:
```
package main


type TestTable struct {
	ID          int64  `db:"id" json:"id" alias:"primary key"`
	StudentName string `db:"student_name" json:"student_name" alias:"student name"`
	CreatedAt   int64  `db:"created_at" json:"created_at"`
}
```


