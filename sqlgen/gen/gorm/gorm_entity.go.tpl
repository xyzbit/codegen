package entity

// {{UpperCamel $.Table.Name}} entity a {{$.Table.Name}} struct data.
type {{UpperCamel $.Table.Name}} struct {
    {{- range $.Table.Columns}}
    {{UpperCamel .Name}} {{.GoType}} `json:"{{.Name}}"`{{if .HasComment}}// {{TrimNewLine .Comment}}{{end}}
    {{- end}}
}