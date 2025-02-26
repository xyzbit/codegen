package templatex

import (
	"strings"
	"text/template"

	"github.com/iancoleman/strcase"
	"github.com/xyzbit/codegen/pkg/stringx"
)

func lineComment(s string) string {
	fields := strings.FieldsFunc(s, func(r rune) bool {
		return r == '\n'
	})
	return strings.Join(fields, "\n// ")
}

var funcMap = template.FuncMap{
	"UpperCamel":  strcase.ToCamel,
	"LowerCamel":  strcase.ToLowerCamel,
	"Join":        strings.Join,
	"TrimNewLine": stringx.TrimNewLine,
	"LineComment": lineComment,
}
