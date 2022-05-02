package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

type Element int32

const (
	ElementTableDefault Element = iota
	ElementTable
	ElementField
	ElementFieldType
	ElementComment
)

func (t Element) String() string {
	switch t {
	case ElementTableDefault:
		return "unknown"
	case ElementTable:
		return "table"
	case ElementField:
		return "field"
	case ElementFieldType:
		return "field type"
	case ElementComment:
		return "comment"
	}
	return ""
}

type GroupFlag int32

const (
	GroupFlagUnknown GroupFlag = iota
	GroupFlagDefault
	GroupFlagSingleQuote
	GroupFlagBackQuote
)

func (t GroupFlag) String() string {
	switch t {
	case GroupFlagUnknown:
		return "未知"
	case GroupFlagDefault:
		return "空格或括号"
	case GroupFlagSingleQuote:
		return "单引号"
	case GroupFlagBackQuote:
		return "反引号"
	}
	return ""
}

type WriteMode string

const (
	NONE      WriteMode = ""
	APPEND    WriteMode = "APPEND"
	OVERWRITE WriteMode = "OVERWRITE"
)

func (mode WriteMode) IsAllowed() bool {
	for _, m := range AllowedMode {
		if m == mode {
			return true
		}
	}
	return false
}

type ConvertFunc func(string) string

var AllowedMode = []WriteMode{APPEND, OVERWRITE}

var defaultConvertFunc = func(source string) string {
	strs := strings.Split(source, "_")

	var res []string
	for _, str := range strs {
		if _, exist := abbr[str]; exist {
			res = append(res, strings.ToUpper(str))
			continue
		}
		tmp := []rune(str)
		if len(tmp) == 0 {
			continue
		}
		r := []rune(strings.ToUpper(string(tmp[0])))
		tmp[0] = r[0]
		res = append(res, string(tmp))
	}
	return strings.Join(res, "")
}

type TableStruct struct {
	TableName string
	Fields    []*FieldInfo
}

func (t *TableStruct) String() string {
	b, _ := json.Marshal(t)
	return string(b)
}

type SSField struct {
	FieldName string
	FiledType string
	Comment   string
}

type SS struct {
	StructName string
	Fields     []*SSField
}

func (s *SS) String() string {
	b, _ := json.Marshal(s)
	return string(b)
}

type FieldInfo struct {
	FieldName    string // the name of field
	FieldType    string // the type of field
	FieldComment string
}

type CreateTableSQLParser struct {
	Strict          bool
	Tags            []string
	CommentTag      string
	Sqls            []string // the imput create sql
	Converter       ConvertFunc
	TableNamePrefix string
	TableNameSuffix string
	FieldNamePrefix string
	FieldNameSuffix string
	SqlFile         string
	TargetDir       string
	Mode            WriteMode

	structs []*SS
}

func (parser *CreateTableSQLParser) SetDefault() *CreateTableSQLParser {
	if len(parser.Tags) == 0 {
		parser.Tags = []string{"json", "db"}
	}
	if parser.CommentTag == "" {
		parser.CommentTag = "alias"
	}
	if parser.TargetDir == "" {
		d, _ := filepath.Split(parser.SqlFile)
		parser.TargetDir = d
	}
	parser.TargetDir = strings.TrimSuffix(parser.TargetDir, "/")
	if parser.Converter == nil {
		parser.Converter = defaultConvertFunc
	}
	if parser.Mode == NONE {
		parser.Mode = APPEND
	}
	return parser
}

func (parser *CreateTableSQLParser) Parse() error {
	parser.SetDefault()
	if err := parser.parseSQL(); err != nil {
		return err
	}
	return parser.output(string(parser.format()))
}

func (parser *CreateTableSQLParser) parseSQL() error {
	for _, ele := range parser.Sqls {
		table, err := extractTableStruct(strings.ToLower(ele))
		if err != nil {
			return err
		}
		if table == nil {
			continue
		}
		parser.structs = append(parser.structs, parser.fromTableStruct2SS(table, parser.getConvertFunc()))
	}
	return nil
}

func (parser *CreateTableSQLParser) format() []byte {
	var res []string
	res = append(res, "package main\n")
	for _, ss := range parser.structs {
		res = append(res, parser.formatOne(ss))
	}
	return []byte(strings.Join(res, "\n\n"))
}

func (parser *CreateTableSQLParser) formatOne(ss *SS) string {
	var tmp [][]string
	for _, field := range ss.Fields {
		tmp = append(tmp, []string{
			field.FieldName, field.FiledType, field.Comment,
		})
	}

	var reverseTmpWithPadding [][]string
	for j := 0; j < 2; j++ {
		var cols []string
		for i := 0; i < len(tmp); i++ {
			cols = append(cols, tmp[i][j])
		}

		reverseTmpWithPadding = append(reverseTmpWithPadding, colPadding(cols))
	}

	var res []string
	res = append(res, fmt.Sprintf("type %s struct {", ss.StructName))
	for j := 0; j < len(reverseTmpWithPadding[0]); j++ {
		res = append(res, fmt.Sprintf("\t%s %s %s", reverseTmpWithPadding[0][j], reverseTmpWithPadding[1][j], tmp[j][2]))
	}
	res = append(res, "}")

	return strings.Join(res, "\n")
}

func colPadding(cols []string) []string {
	blankPaddingFunc := func(n int) string {
		var tmp []rune
		for i := 0; i < n; i++ {
			tmp = append(tmp, ' ')
		}
		return string(tmp)
	}
	maxLen := 0
	for i := 0; i < len(cols); i++ {
		maxLen = MaxInt(len(cols[i]), maxLen)
	}

	var res []string
	for i := 0; i < len(cols); i++ {
		res = append(res, cols[i]+blankPaddingFunc(maxLen-len(cols[i])))
	}
	return res
}

func (parser *CreateTableSQLParser) getFileHandler(target string) (*os.File, error) {
	switch parser.Mode {
	case APPEND:
		return os.OpenFile(target, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	case OVERWRITE:
		return os.OpenFile(target, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	}
	return nil, fmt.Errorf("write mode should be one of %v", AllowedMode)
}

func (parser *CreateTableSQLParser) output(str string) error {
	targetFile := parser.TargetDir + "/generator.go"
	file, err := parser.getFileHandler(targetFile)
	if err != nil {
		return err
	}
	defer file.Close()
	fmt.Printf("Output: %s\n", targetFile)

	file.WriteString(str)
	return nil
}

func (parser *CreateTableSQLParser) getConvertFunc() ConvertFunc {
	return parser.Converter
}

func (parser *CreateTableSQLParser) fromTableStruct2SS(s *TableStruct, converter ConvertFunc) *SS {
	if s == nil {
		return nil
	}
	res := &SS{
		StructName: converter(parser.cleanTableName(s.TableName)),
	}

	for _, field := range s.Fields {
		clearnFieldName := parser.cleanFieldName(field.FieldName)
		sField := &SSField{
			FieldName: converter(clearnFieldName),
			FiledType: mapping.getGoStructType(field.FieldType).getString(),
		}

		var tmp []string
		for _, tag := range parser.Tags {
			tmp = append(tmp, fmt.Sprintf("%s:\"%s\"", tag, clearnFieldName))
		}
		if field.FieldComment != "" {
			tmp = append(tmp, fmt.Sprintf("%s:\"%s\"", parser.CommentTag, field.FieldComment))
		}
		sField.Comment = wrapperBackQuote(strings.Join(tmp, " "))
		res.Fields = append(res.Fields, sField)
	}

	return res

}

func wrapperBackQuote(str string) string {
	return "`" + str + "`"
}

func (parser *CreateTableSQLParser) cleanTableName(tableName string) string {
	return strings.TrimSuffix(strings.TrimPrefix(tableName, parser.TableNamePrefix), parser.TableNameSuffix)
}

func (parser *CreateTableSQLParser) cleanFieldName(tableName string) string {
	return strings.TrimSuffix(strings.TrimPrefix(tableName, parser.FieldNamePrefix), parser.FieldNameSuffix)
}

func extractTableStruct(sql string) (*TableStruct, error) {
	rs := []rune(sql)

	var (
		i            = 0
		preChar      = ' '
		currChar     = ' '
		preGroupFlag = GroupFlagDefault
	)

	findNextElementFunc := func() (GroupFlag, string) {

		var (
			cs    []rune
			start bool
		)

		for ; i < len(rs); i++ {
			preChar = currChar
			currChar = rs[i]

			isPreSep := isSep(preChar)
			isCurrSep := isSep(currChar)
			groupFlag := getGroupFlag(currChar)

			if isPreSep && isCurrSep && !start {
				continue
			}
			if (!isPreSep && isCurrSep && groupFlag == preGroupFlag) ||
				((groupFlag == GroupFlagSingleQuote || groupFlag == GroupFlagBackQuote) && groupFlag == preGroupFlag) {
				preGroupFlag = GroupFlagUnknown
				i += 1
				return groupFlag, string(cs)

			}
			if isPreSep && !isCurrSep && !start {
				preGroupFlag = getGroupFlag(preChar)
				start = true
			}

			// fmt.Println(string(currChar))
			cs = append(cs, currChar)
		}
		return GroupFlagUnknown, ""
	}

	var (
		elementTypes    []Element
		elementValues   []string
		preElementValue string
		preFlag         GroupFlag
	)
	for {
		flag, elementStr := findNextElementFunc()
		if elementStr == "" {
			break
		}
		if flag == GroupFlagBackQuote {
			if (preFlag == GroupFlagBackQuote && len(elementTypes) > 1) || strings.ToLower(preElementValue) == "key" {
				break
			}
			elementValues = append(elementValues, elementStr)
			if len(elementTypes) == 0 {
				elementTypes = append(elementTypes, ElementTable)
			} else {
				elementTypes = append(elementTypes, ElementField)
			}
		} else {
			if preFlag == GroupFlagBackQuote {
				elementTypes = append(elementTypes, ElementFieldType)
				elementValues = append(elementValues, elementStr)
			}
			if strings.ToLower(preElementValue) == "comment" {
				elementTypes = append(elementTypes, ElementComment)
				elementValues = append(elementValues, elementStr)
			}
		}

		preElementValue = elementStr
		preFlag = flag
	}

	var field *FieldInfo
	table := &TableStruct{}
	for k := 0; k < len(elementTypes); k++ {
		value := elementValues[k]
		switch elementTypes[k] {
		case ElementTable:
			table.TableName = value
		case ElementField:
			if field != nil && field.FieldType != "" && field.FieldName != "" {
				table.Fields = append(table.Fields, field)
			}
			field = &FieldInfo{
				FieldName: value,
			}
		case ElementFieldType:
			field.FieldType = value
		case ElementComment:
			field.FieldComment = value
		}
	}
	if field != nil && field.FieldType != "" && field.FieldName != "" {
		table.Fields = append(table.Fields, field)
	}
	return table, nil
}

func IsFieldOrTableName(str string) bool {
	return strings.HasPrefix(str, "`") || strings.HasSuffix(str, "`")
}

func getGroupFlag(s rune) GroupFlag {
	if unicode.IsSpace(s) || s == '(' || s == ')' {
		return GroupFlagDefault
	}
	if s == '\'' {
		return GroupFlagSingleQuote
	}
	if s == '`' {
		return GroupFlagBackQuote
	}
	return GroupFlagUnknown
}

func is_group_flag(s rune) bool {
	return s == '`' || s == '\''
}

func isSep(s rune) bool {
	return unicode.IsSpace(s) || s == '(' || s == ')' || s == ',' || is_group_flag(s)
}
