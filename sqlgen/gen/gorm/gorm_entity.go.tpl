package entity

// {{UpperCamel $.Table.Name}} column names.
const(
    {{range $.Table.Columns}}
    {{UpperCamel $.Table.Name}}{{UpperCamel .Name}} = "{{.Name}}" {{if .HasComment}}// {{TrimNewLine .Comment}}{{end}}{{end}}
)

// {{UpperCamel $.Table.Name}} entity a {{$.Table.Name}} struct data.
type {{UpperCamel $.Table.Name}} struct {
    {{- range $.Table.Columns}}
    {{UpperCamel .Name}} {{.GoType}} `json:"{{.Name}}"`{{if .HasComment}}// {{TrimNewLine .Comment}}{{end}}
    {{- end}}
}