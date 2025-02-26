package {{$.AdapterPackageName}}

import (
    "context"
    "fmt"
    "time"

    "gorm.io/gorm"
    "github.com/samber/lo"
    "github.com/xyzbit/gpkg/gormx"
    "github.com/xyzbit/gpkg/ctxwrap"

    repo "{{$.RepoPackage}}"
    entity "{{$.EntityPackage}}"
)

// {{UpperCamel $.Table.Name}}Adapter represents a {{$.Table.Name}} adapter.
type {{UpperCamel $.Table.Name}}Adapter struct {
    db *gorm.DB
}

// New{{UpperCamel $.Table.Name}}Repo returns a new {{$.Table.Name}} adapter implemented {{$.Table.Name}}Repo.
func New{{UpperCamel $.Table.Name}}Repo (
    db *gorm.DB,
) repo.{{UpperCamel $.Table.Name}}Repo {
    return &{{UpperCamel $.Table.Name}}Adapter{db: db}
}

func (m *{{UpperCamel $.Table.Name}}Adapter) DB(ctx context.Context) *gorm.DB {
	tx := ctxwrap.FromGormDBContext(ctx)
	if tx != nil {
		return tx
	}
	return m.db.WithContext(ctx)
}

// Create creates  {{$.Table.Name}} data.
func (m *{{UpperCamel $.Table.Name}}Adapter) Create(ctx context.Context, es ...*entity.{{UpperCamel $.Table.Name}}) error {
    if len(es)==0{
        return fmt.Errorf("data is empty")
    }
    {{if $.AutoAudit }}operator := ctxwrap.FromOperatorContext(ctx){{end}}

    pos := lo.Map(es, func(v *entity.{{UpperCamel $.Table.Name}}, _ int) *{{UpperCamel $.Table.Name}} {
		{{- if $.AutoAudit }}
        p := to{{UpperCamel $.Table.Name}}PO(ctx, v)
		p.Creator = operator.Username
		p.Operator = operator.Username
		return p
        {{- else}}
        return to{{UpperCamel $.Table.Name}}PO(ctx, v)
        {{- end}}
	})
    
    return m.DB(ctx).Create(&pos).Error
}

// GetByID get {{$.Table.Name}} by id.
func (r *{{UpperCamel $.Table.Name}}Adapter) GetByID(ctx context.Context, id int64) (*entity.{{UpperCamel $.Table.Name}}, error) {
    var result entity.{{UpperCamel $.Table.Name}}
    
    err := r.DB(ctx).Where("id = ?", id).First(&result).Error
    
    return &result, err
}

// List list {{$.Table.Name}}.
func (m *{{UpperCamel $.Table.Name}}Adapter) List(ctx context.Context, query *gormx.Query) ([]*entity.{{UpperCamel $.Table.Name}}, error) {
    var pos []*{{UpperCamel $.Table.Name}}

	err := query.
        WithDB(m.DB(ctx)).
		Find(&pos).Error
	if err != nil {
		return nil, err
	}

    entitys := lo.Map(pos, func(v *{{UpperCamel $.Table.Name}}, _ int) *entity.{{UpperCamel $.Table.Name}} {
		return to{{UpperCamel $.Table.Name}}Entity(ctx, v)
	})

    return entitys, nil
}


// Count count {{$.Table.Name}}.
func (m *{{UpperCamel $.Table.Name}}Adapter) Count(ctx context.Context, query *gormx.Query) (int64, error) {
    var count int64

	err := query.
        WithDB(m.DB(ctx)).
		Model(&{{UpperCamel $.Table.Name}}{}).
		Count(&count).Error

	return count, err
}

// Update update {{$.Table.Name}}.
func (m *{{UpperCamel $.Table.Name}}Adapter) Update(ctx context.Context, e *entity.{{UpperCamel $.Table.Name}}) error {
	{{- if $.AutoAudit }}
    operator := ctxwrap.FromOperatorContext(ctx)
	p := to{{UpperCamel $.Table.Name}}PO(ctx, e)
	p.Operator = operator.Username

    return m.DB(ctx).Updates(p).Error
    {{- else}}
    return m.DB(ctx).Updates(to{{UpperCamel $.Table.Name}}PO(ctx, e)).Error
    {{- end}}
}

// Delete delete {{$.Table.Name}}.
func (m *{{UpperCamel $.Table.Name}}Adapter) Delete(ctx context.Context, id int64) error {
	return m.DB(ctx).
		Where("id = ?", id).
		Delete(&{{UpperCamel $.Table.Name}}{}).Error
}

// IsDuplicatedKeyError use to check error is unique key conflict error.
func (m *{{UpperCamel $.Table.Name}}Adapter) IsDuplicatedKeyError(err error) bool {
	return errors.Is(err, gorm.ErrDuplicatedKey)
}

// IsNotFoundError use to check error is record not found error.
func (m *{{UpperCamel $.Table.Name}}Adapter) IsNotFoundError(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}

// {{UpperCamel $.Table.Name}} represents a {{$.Table.Name}} struct data.
type {{UpperCamel $.Table.Name}} struct { {{range $.Table.Columns}}
    {{UpperCamel .Name}} {{.GoType}} `gorm:"column:{{.Name}}{{if IsPrimary .Name}};primaryKey{{end}}{{if .AutoIncrement}};autoIncrement{{end}}{{if eq .Name "created_time"}};autoCreateTime{{end}}{{if eq .Name "updated_time"}};autoUpdateTime{{end}}" json:"{{.Name}}"`{{if .HasComment}}// {{TrimNewLine .Comment}}{{end}}{{end}}
}

// TableName returns the table name. it implemented by gorm.Tabler.
func (m *{{UpperCamel $.Table.Name}}) TableName() string {
    return "{{$.Table.Name}}"
}

func to{{UpperCamel $.Table.Name}}PO(ctx context.Context, e *entity.{{UpperCamel $.Table.Name}}) *{{UpperCamel $.Table.Name}} {
	_ = ctx
	return &{{UpperCamel $.Table.Name}}{
        {{- range $.Table.Columns}}
        {{UpperCamel .Name}}: e.{{UpperCamel .Name}},
        {{- end}}
    }
}

func to{{UpperCamel $.Table.Name}}Entity(ctx context.Context, po *{{UpperCamel $.Table.Name}}) *entity.{{UpperCamel $.Table.Name}} {
	_ = ctx
	return &entity.{{UpperCamel $.Table.Name}}{
        {{- range $.Table.Columns}}
        {{UpperCamel .Name}}: po.{{UpperCamel .Name}},
        {{- end}}
    }
}