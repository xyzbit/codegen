package service

import (
	"context"

	"github.com/xyzbit/gpkg/gormx"
	"gorm.io/gorm"

	entity "github.com/xyzbit/codegen/sqlgen/example/entity"
)

type UserRepo interface {
	DB(ctx context.Context) *gorm.DB

	GetByID(ctx context.Context, id int64) (*entity.User, error)
	Create(ctx context.Context, data ...*entity.User) error
	List(ctx context.Context, query *gormx.Query) ([]*entity.User, error)
	Count(ctx context.Context, query *gormx.Query) (int64, error)
	Update(ctx context.Context, e *entity.User) error
	Delete(ctx context.Context, id int64) error

	IsDuplicatedKeyError(err error) bool
	IsNotFoundError(err error) bool
}
