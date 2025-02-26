package {{$.RepoPackageName}}

import (
    "context"

    "gorm.io/gorm"
    "github.com/xyzbit/gpkg/gormx"

    entity "{{$.EntityPackage}}"
)

type {{UpperCamel $.Table.Name}}Repo interface {
    DB(ctx context.Context) *gorm.DB

    GetByID(ctx context.Context, id int64) (*entity.{{UpperCamel $.Table.Name}}, error)
    Create(ctx context.Context, data ...*entity.{{UpperCamel $.Table.Name}}) error
    List(ctx context.Context, query *gormx.Query) ([]*entity.{{UpperCamel $.Table.Name}}, error)
    Count(ctx context.Context, query *gormx.Query) (int64, error)
    Update(ctx context.Context, e *entity.{{UpperCamel $.Table.Name}}) error
    Delete(ctx context.Context, id int64) error

    IsDuplicatedKeyError(err error) bool
    IsNotFoundError(err error) bool
}
