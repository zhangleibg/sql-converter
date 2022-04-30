package main

import (
	"strings"
)

type MappedGoFieldType string

const (
	FeildTypeDefault MappedGoFieldType = "interface{}"
	FeildTypeString  MappedGoFieldType = "string"
	FeildTypeInt64   MappedGoFieldType = "int64"
	FeildTypeInt32   MappedGoFieldType = "int32"
	FeildTypeInt     MappedGoFieldType = "int"
	FeildTypeFloat64 MappedGoFieldType = "float64"
	FeildTypeFloat32 MappedGoFieldType = "float32"
	FieldTypeTime    MappedGoFieldType = "time.Time"
)

func (t MappedGoFieldType) getString() string {
	return string(t)
}

type FieldTypeMapper map[MappedGoFieldType][]string

func (m FieldTypeMapper) getGoStructType(sqlField string) MappedGoFieldType {
	upperFieldName := strings.ToUpper(sqlField)
	for t, fieldNames := range m {
		for _, fieldName := range fieldNames {
			if upperFieldName == fieldName {
				return t
			}
		}
	}
	return FeildTypeDefault
}

var (
	mapping FieldTypeMapper
	abbr    map[string]struct{}
)

func init() {
	mapping = make(FieldTypeMapper)
	mapping[FeildTypeString] = []string{
		"CHAR", "VARCHAR", "BINARY", "VARBINARY", "TINYBLOB", "TINYTEXT",
		"TEXT", "BLOB", "MEDIUMTEXT", "LONGTEXT", "LONGBLOB", "ENUM", "SET",
	}
	mapping[FeildTypeInt64] = []string{
		"BIGINT", "TIMESTAMP",
	}
	mapping[FeildTypeInt32] = []string{
		"INT", "INTEGER",
	}
	mapping[FeildTypeInt] = []string{
		"BIT", "BOOL", "BOOLEAN", "SMALLINT", "MEDIUMINT",
	}
	mapping[FeildTypeFloat64] = []string{
		"DOUBLE", "DECIMAL", "DEC",
	}
	mapping[FeildTypeFloat32] = []string{
		"FLOAT",
	}
	mapping[FieldTypeTime] = []string{
		"DATE", "DATETIME", "TIME", "YEAR",
	}

	abbr = make(map[string]struct{})
	abbr["id"] = struct{}{}
}

func MaxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
