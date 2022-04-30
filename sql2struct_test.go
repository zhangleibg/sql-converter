package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

const testCreateSQL = "CREATE TABLE IF NOT EXISTS `v_test_table` (`id` BIGINT(20) UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键, 无实际意义',`student_name` VARCHAR(128) NOT NULL COMMENT '学生姓名', `created_at` TIMESTAMP NOT NULL CURRENT_STAMP ON UPDATE CURRENT_STAMP) ENGINE=InnoDB COMMENT='测试表'"

func TestExtractTableStruct(t *testing.T) {
	res, err := extractTableStruct(testCreateSQL)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(res)
}

func TestDefaultConvertFunc(t *testing.T) {
	input := "v_test_table"
	expected := "VTestTable"
	actual := defaultConvertFunc(input)
	fmt.Println(actual)
	assert.Equal(t, expected, actual)
}

func TestFromTableStruct2SS(t *testing.T) {
	input := &TableStruct{
		TableName: "v_test_table",
		Fields: []*FieldInfo{
			{
				FieldName:    "field_1",
				FieldType:    "varchar",
				FieldComment: "测试",
			},
			{
				FieldName:    "field_2",
				FieldType:    "aaa",
				FieldComment: "",
			},
			{
				FieldName:    "field_suf",
				FieldType:    "bigint",
				FieldComment: "测试3",
			},
		},
	}

	expected := &SS{
		StructName: "TestTable",
		Fields: []*SSField{
			{
				FieldName: "Field1",
				FiledType: "string",
				Comment:   "`json:\"field_1\" db:\"field_1\" comment:\"测试\"`",
			},
			{
				FieldName: "Field2",
				FiledType: "interface{}",
				Comment:   "`json:\"field_2\" db:\"field_2\"`",
			},
			{
				FieldName: "Field",
				FiledType: "int64",
				Comment:   "`json:\"field\" db:\"field\" comment:\"测试3\"`",
			},
		},
	}

	parser := &CreateTableSQLParser{
		Tags:            []string{"json", "db"},
		CommentTag:      "comment",
		TableNamePrefix: "v_",
		FieldNameSuffix: "_suf",
	}

	actual := parser.fromTableStruct2SS(input, defaultConvertFunc)

	assert.Equal(t, expected, actual)

}

func TestFormatOne(t *testing.T) {
	input := &SS{
		StructName: "TestTable",
		Fields: []*SSField{
			{
				FieldName: "Field1",
				FiledType: "string",
				Comment:   "测试",
			},
			{
				FieldName: "Field2",
				FiledType: "int64",
				Comment:   "测试int64",
			},
		},
	}
	expected :=
		`package main

type TestTable struct {
	Field1 string 测试
	Field2 int64  测试int64
}`
	parser := &CreateTableSQLParser{}

	actual := parser.formatOne(input)
	assert.Equal(t, expected, actual)
}

func TestReadCreateSQL(t *testing.T) {
	sql := "CREATE TABLE `v_test_1`();SELECT * FROM v_test_1; CREATE TABLE `v_test_2`; ALTER TABLE `v_test_1` ADD INDEX `idx_q`(`q`)"
	expected := []string{
		"CREATE TABLE `v_test_1`()",
		"CREATE TABLE `v_test_2`",
	}
	actual, err := readCreateSQL([]byte(sql))
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, expected, actual)
}

func TestParseArg(t *testing.T) {
	cases := map[string][]string{
		"Case1": {
			"-aa", "-file",
		},
		"Case2": {
			"./test.sql", "-aa", "-file",
		},
		"Case3": {
			"./test.sql",
			"-tags=json,db",
			"-comment_tag",
		},
		"Case4": {
			"./test.sql",
			"-tags=json,db",
			"-comment_tag=comment",
			"-field_prefix=v_",
			"-table_prefix=v_",
			"-table_suffix=v_=v_",
		},
	}
	expected := map[string]map[string]string{
		"Case1": nil,
		"Case2": nil,
		"Case3": {
			"-file":        "./test.sql",
			"-tags":        "json,db",
			"-comment_tag": "",
		},
		"Case4": {
			"-file":         "./test.sql",
			"-tags":         "json,db",
			"-comment_tag":  "comment",
			"-field_prefix": "v_",
			"-table_prefix": "v_",
			"-table_suffix": "v_",
		},
	}

	for name, args := range cases {
		t.Run(name, func(t *testing.T) {
			parser, _ := parseArg(args)
			assert.Equal(t, expected[name], parser)
		})
	}
}
