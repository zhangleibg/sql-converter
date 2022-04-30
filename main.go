package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

const (
	usage = "Usage: sql.converter <path> [-tags=<tags>] [-comment_tag=<comment_tag>] [-table_prefix=<table_prefix>] " +
		"[-table_suffix=<table_suffix>] [-field_prefix=<field_prefix>] [-field_suffix=<field_suffix>] " +
		"[-h] [-target=<target>] [-mode=<mode>]"
	params = `
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
`
)

type handlerFunc func(*CreateTableSQLParser, string) error

var flags = map[string]handlerFunc{
	"-tags": func(cts *CreateTableSQLParser, arg string) error {
		tags := strings.Split(arg, ",")
		if len(tags) == 0 {
			return fmt.Errorf("empty tags parsed, %s", arg)
		}
		var params []string
		for _, tag := range tags {
			params = append(params, strings.TrimSpace(tag))
		}
		cts.Tags = params
		return nil
	},
	"-comment_tag": func(cts *CreateTableSQLParser, arg string) error {
		if tag := strings.TrimSpace(arg); tag != "" {
			cts.CommentTag = tag
			return nil
		}
		return fmt.Errorf("empty comment_tag parsed, %s", arg)
	},
	"-table_prefix": func(cts *CreateTableSQLParser, arg string) error {
		if prefix := strings.TrimSpace(arg); prefix != "" {
			cts.TableNamePrefix = prefix
			return nil
		}
		return fmt.Errorf("empty table_prefix parsed, %s", arg)
	},
	"-table_suffix": func(cts *CreateTableSQLParser, arg string) error {
		if suffix := strings.TrimSpace(arg); suffix != "" {
			cts.TableNameSuffix = suffix
			return nil
		}
		return fmt.Errorf("empty table_suffix parsed, %s", arg)
	},
	"-field_prefix": func(cts *CreateTableSQLParser, arg string) error {
		if prefix := strings.TrimSpace(arg); prefix != "" {
			cts.FieldNamePrefix = prefix
			return nil
		}
		return fmt.Errorf("empty field_prefix parsed, %s", arg)
	},
	"-field_suffix": func(cts *CreateTableSQLParser, arg string) error {
		if suffix := strings.TrimSpace(arg); suffix != "" {
			cts.FieldNameSuffix = suffix
			return nil
		}
		return fmt.Errorf("empty field_suffix parsed, %s", arg)
	},
	"-h": func(cts *CreateTableSQLParser, arg string) error {
		fmt.Printf("%s%s", usage, params)
		return nil
	},
	"-file": func(cts *CreateTableSQLParser, s string) error {
		path := s
		cts.SqlFile = path
		content, err := ioutil.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read sql failed, err: %v", err)
		}
		res, err := readCreateSQL(content)
		if err != nil {
			return fmt.Errorf("read sql failed, err: %v", err)
		}
		cts.Sqls = res
		return nil
	},
	"-target": func(cts *CreateTableSQLParser, s string) error {
		cts.TargetDir = s
		return nil
	},
	"-mode": func(cts *CreateTableSQLParser, s string) error {
		s = strings.ToUpper(strings.TrimSpace(s))
		mode := WriteMode(s)
		if !mode.IsAllowed() {
			return fmt.Errorf("write mode should be one of %v", AllowedMode)
		}
		cts.Mode = mode
		return nil
	},
}

func readCreateSQL(content []byte) ([]string, error) {

	eles := strings.Split(string(content), ";")

	var res []string
	for _, ele := range eles {
		trimed := strings.TrimSpace(ele)
		if strings.HasPrefix(strings.ToLower(trimed), "create") {
			res = append(res, trimed)
		}
	}
	return res, nil
}

func parseArg(args []string) (map[string]string, error) {
	if len(args) == 0 {
		return nil, errors.New("the sql file is necessray")
	}

	flag2Param := make(map[string]string)
	// first arg should be the sql file
	firstArg := args[0]
	if _, exist := flags[firstArg]; exist {
		return nil, fmt.Errorf("the 1st arg passed should be the sql file, passed: %s", firstArg)
	}
	file, err := os.Stat(firstArg)
	if os.IsNotExist(err) || file.IsDir() {
		return nil, fmt.Errorf("%s doesn't exist or is not a directory", firstArg)
	}
	flag2Param["-file"] = firstArg

	for i := 1; i < len(args); i++ {
		if !strings.HasPrefix(args[i], "-") {
			return nil, fmt.Errorf("%s is not a correct flag", args[i])
		}
		eles := strings.Split(args[i], "=")

		var (
			flag  string
			param string
		)
		flag = eles[0]
		if len(eles) > 1 {
			param = eles[1]
		}
		if _, exist := flags[flag]; !exist {
			return nil, fmt.Errorf("%s is not within the given flags", flag)
		}
		flag2Param[flag] = param
	}
	return flag2Param, nil
}

func getParser(flag2param map[string]string) (*CreateTableSQLParser, error) {
	cts := &CreateTableSQLParser{}
	for flag, param := range flag2param {
		err := flags[flag](cts, param)
		if err != nil {
			return nil, fmt.Errorf("param parsing failed, flag: %s, param: %s, error: %v", flag, param, err)
		}
	}
	return cts, nil
}

func main() {
	flag2param, err := parseArg(os.Args[1:])
	if err != nil {
		fmt.Println(err)
		return
	}
	parser, err := getParser(flag2param)
	if err != nil {
		fmt.Println(err)
		return
	}
	if err = parser.Parse(); err != nil {
		fmt.Println(err)
		return
	}
}
