package {{$.AdapterPackageName}}

import (
    "context"
    "fmt"

    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"

    entity "{{$.EntityPackage}}"
    repo "{{$.RepoPackage}}"
)

// SQLiteMock{{UpperCamel $.Table.Name}}Adapter SQLite 测试适配器
type SQLiteMock{{UpperCamel $.Table.Name}}Adapter struct {
    db *gorm.DB
}

// NewSQLiteMock{{UpperCamel $.Table.Name}}Repo 创建一个新的基于 SQLite 的测试适配器
func NewSQLiteMock{{UpperCamel $.Table.Name}}Repo() (repo.{{UpperCamel $.Table.Name}}Repo, error) {
    // 使用内存数据库
    db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
        Logger: logger.Default.LogMode(logger.Silent),
    })
    if err != nil {
        return nil, fmt.Errorf("failed to connect database: %w", err)
    }

    // 启用外键约束
    db.Exec("PRAGMA foreign_keys = ON")

    // 自动迁移表结构
    if err := db.AutoMigrate(&entity.{{UpperCamel $.Table.Name}}{}); err != nil {
        return nil, fmt.Errorf("failed to migrate table: %w", err)
    }

    return &SQLiteMock{{UpperCamel $.Table.Name}}Adapter{
        db: db,
    }, nil
}

func (m *SQLiteMock{{UpperCamel $.Table.Name}}Adapter) DB(ctx context.Context) *gorm.DB {
    return m.db.WithContext(ctx)
}

func (m *SQLiteMock{{UpperCamel $.Table.Name}}Adapter) GetByID(ctx context.Context, id int64) (*entity.{{UpperCamel $.Table.Name}}, error) {
    var result entity.{{UpperCamel $.Table.Name}}
    err := m.db.WithContext(ctx).First(&result, id).Error
    if err != nil {
        return nil, err
    }
    return &result, nil
}

func (m *SQLiteMock{{UpperCamel $.Table.Name}}Adapter) Create(ctx context.Context, data ...*entity.{{UpperCamel $.Table.Name}}) error {
    return m.db.WithContext(ctx).Create(&data).Error
}

func (m *SQLiteMock{{UpperCamel $.Table.Name}}Adapter) List(ctx context.Context, query *gormx.Query) ([]*entity.{{UpperCamel $.Table.Name}}, error) {
    var results []*entity.{{UpperCamel $.Table.Name}}
    db := m.db.WithContext(ctx)
    if query != nil {
        db = query.WithDB(db)
    }
    if err := db.Find(&results).Error; err != nil {
        return nil, err
    }
    return results, nil
}

func (m *SQLiteMock{{UpperCamel $.Table.Name}}Adapter) Count(ctx context.Context, query *gormx.Query) (int64, error) {
    var count int64
    db := m.db.WithContext(ctx)
    if query != nil {
        db = query.WithDB(db)
    }
    if err := db.Model(&entity.{{UpperCamel $.Table.Name}}{}).Count(&count).Error; err != nil {
        return 0, err
    }
    return count, nil
}

func (m *SQLiteMock{{UpperCamel $.Table.Name}}Adapter) Update(ctx context.Context, e *entity.{{UpperCamel $.Table.Name}}) error {
    return m.db.WithContext(ctx).Save(e).Error
}

func (m *SQLiteMock{{UpperCamel $.Table.Name}}Adapter) Delete(ctx context.Context, id int64) error {
    return m.db.WithContext(ctx).Delete(&entity.{{UpperCamel $.Table.Name}}{}, id).Error
}

func (m *SQLiteMock{{UpperCamel $.Table.Name}}Adapter) IsDuplicatedKeyError(err error) bool {
    if err == nil {
        return false
    }
    return err.Error() == "UNIQUE constraint failed"
}

func (m *SQLiteMock{{UpperCamel $.Table.Name}}Adapter) IsNotFoundError(err error) bool {
    return errors.Is(err, gorm.ErrRecordNotFound)
}

func (m *SQLiteMock{{UpperCamel $.Table.Name}}Adapter) Reset(ctx context.Context) error {
    return m.db.WithContext(ctx).Exec(fmt.Sprintf("DELETE FROM %s", m.db.NamingStrategy.TableName("{{$.Table.Name}}"))).Error
}

func (m *SQLiteMock{{UpperCamel $.Table.Name}}Adapter) Close() error {
    if m.db != nil {
        sqlDB, err := m.db.DB()
        if err != nil {
            return fmt.Errorf("failed to get underlying db: %w", err)
        }
        return sqlDB.Close()
    }
    return nil
} 