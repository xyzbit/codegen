package data

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/samber/lo"
	"github.com/xyzbit/gpkg/ctxwrap"
	"github.com/xyzbit/gpkg/gormx"
	"gorm.io/gorm"

	entity "github.com/xyzbit/codegen/sqlgen/example/entity"
	repo "github.com/xyzbit/codegen/sqlgen/example/service"
)

// UserAdapter represents a user adapter.
type UserAdapter struct {
	db *gorm.DB
}

// NewUserRepo returns a new user adapter implemented userRepo.
func NewUserRepo(
	db *gorm.DB,
) repo.UserRepo {
	return &UserAdapter{db: db}
}

func (m *UserAdapter) DB(ctx context.Context) *gorm.DB {
	tx := ctxwrap.FromGormDBContext(ctx)
	if tx != nil {
		return tx
	}
	return m.db.WithContext(ctx)
}

// Create creates  user data.
func (m *UserAdapter) Create(ctx context.Context, es ...*entity.User) error {
	if len(es) == 0 {
		return fmt.Errorf("data is empty")
	}

	pos := lo.Map(es, func(v *entity.User, _ int) *User {
		return toUserPO(ctx, v)
	})

	return m.DB(ctx).Create(&pos).Error
}

// GetByID get user by id.
func (r *UserAdapter) GetByID(ctx context.Context, id int64) (*entity.User, error) {
	var result entity.User

	err := r.DB(ctx).Where("id = ?", id).First(&result).Error

	return &result, err
}

// List list user.
func (m *UserAdapter) List(ctx context.Context, query *gormx.Query) ([]*entity.User, error) {
	var pos []*User

	err := query.
		WithDB(m.DB(ctx)).
		Find(&pos).Error
	if err != nil {
		return nil, err
	}

	entitys := lo.Map(pos, func(v *User, _ int) *entity.User {
		return toUserEntity(ctx, v)
	})

	return entitys, nil
}

// Count count user.
func (m *UserAdapter) Count(ctx context.Context, query *gormx.Query) (int64, error) {
	var count int64

	err := query.
		WithDB(m.DB(ctx)).
		Model(&User{}).
		Count(&count).Error

	return count, err
}

// Update update user.
func (m *UserAdapter) Update(ctx context.Context, e *entity.User) error {
	return m.DB(ctx).Updates(toUserPO(ctx, e)).Error
}

// Delete delete user.
func (m *UserAdapter) Delete(ctx context.Context, id int64) error {
	return m.DB(ctx).
		Where("id = ?", id).
		Delete(&User{}).Error
}

// IsDuplicatedKeyError use to check error is unique key conflict error.
func (m *UserAdapter) IsDuplicatedKeyError(err error) bool {
	return errors.Is(err, gorm.ErrDuplicatedKey)
}

// IsNotFoundError use to check error is record not found error.
func (m *UserAdapter) IsNotFoundError(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}

// User represents a user struct data.
type User struct {
	Id                uint32    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Uid               int64     `gorm:"column:uid" json:"uid"`
	NickName          string    `gorm:"column:nick_name" json:"nick_name"`
	AvatarUri         string    `gorm:"column:avatar_uri" json:"avatar_uri"`
	ReadingPreference int8      `gorm:"column:reading_preference" json:"reading_preference"`
	CreateTime        time.Time `gorm:"column:create_time" json:"create_time"`
	UpdateTime        time.Time `gorm:"column:update_time" json:"update_time"`
	AutoBuy           int8      `gorm:"column:auto_buy" json:"auto_buy"`
	IsAutoBuy         int8      `gorm:"column:is_auto_buy" json:"is_auto_buy"`
}

// TableName returns the table name. it implemented by gorm.Tabler.
func (m *User) TableName() string {
	return "user"
}

func toUserPO(ctx context.Context, e *entity.User) *User {
	_ = ctx
	return &User{
		Id:                e.Id,
		Uid:               e.Uid,
		NickName:          e.NickName,
		AvatarUri:         e.AvatarUri,
		ReadingPreference: e.ReadingPreference,
		CreateTime:        e.CreateTime,
		UpdateTime:        e.UpdateTime,
		AutoBuy:           e.AutoBuy,
		IsAutoBuy:         e.IsAutoBuy,
	}
}

func toUserEntity(ctx context.Context, po *User) *entity.User {
	_ = ctx
	return &entity.User{
		Id:                po.Id,
		Uid:               po.Uid,
		NickName:          po.NickName,
		AvatarUri:         po.AvatarUri,
		ReadingPreference: po.ReadingPreference,
		CreateTime:        po.CreateTime,
		UpdateTime:        po.UpdateTime,
		AutoBuy:           po.AutoBuy,
		IsAutoBuy:         po.IsAutoBuy,
	}
}
