package {{$.AdapterPackageName}}

import (
    "context"
    "fmt"
    "time"

    "github.com/testcontainers/testcontainers-go"
    "github.com/testcontainers/testcontainers-go/wait"
    "gorm.io/driver/mysql"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"

    entity "{{$.EntityPackage}}"
    repo "{{$.RepoPackage}}"
)

// DockerMock{{UpperCamel $.Table.Name}}Adapter Docker MySQL 测试适配器
type DockerMock{{UpperCamel $.Table.Name}}Adapter struct {
    db        *gorm.DB
    container testcontainers.Container
}

// NewDockerMock{{UpperCamel $.Table.Name}}Repo 创建一个新的基于 Docker MySQL 的测试适配器
func NewDockerMock{{UpperCamel $.Table.Name}}Repo() (repo.{{UpperCamel $.Table.Name}}Repo, error) {
    ctx := context.Background()
    req := testcontainers.ContainerRequest{
        Image:        "mysql:8.0",
        ExposedPorts: []string{"3306/tcp"},
        Env: map[string]string{
            "MYSQL_ROOT_PASSWORD": "test",
            "MYSQL_DATABASE":      "test",
        },
        WaitingFor: wait.ForAll(
            wait.ForLog("ready for connections"),
            wait.ForListeningPort("3306/tcp"),
        ),
    }

    container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
        ContainerRequest: req,
        Started:          true,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to start container: %w", err)
    }

    // 获取映射端口
    mappedPort, err := container.MappedPort(ctx, "3306")
    if err != nil {
        container.Terminate(ctx)
        return nil, fmt.Errorf("failed to get container port: %w", err)
    }

    // 使用 localhost 而不是容器 IP
    dsn := fmt.Sprintf("root:test@tcp(localhost:%s)/test?charset=utf8mb4&parseTime=True&loc=Local",
        mappedPort.Port(),
    )

    // 添加重试逻辑
    var db *gorm.DB
    maxRetries := 5
    var lastErr error
    for i := 0; i < maxRetries; i++ {
        db, lastErr = gorm.Open(mysql.Open(dsn), &gorm.Config{
            Logger: logger.Default.LogMode(logger.Silent),
        })
        if lastErr == nil {
            break
        }
        fmt.Printf("retry %d: %v\n", i+1, lastErr)
        time.Sleep(time.Second * 2)
    }
    if lastErr != nil {
        _ = container.Terminate(ctx)
        return nil, fmt.Errorf("failed to connect to database after %d retries: %w", maxRetries, lastErr)
    }

    if err := db.AutoMigrate(&entity.{{UpperCamel $.Table.Name}}{}); err != nil {
        _ = container.Terminate(ctx)
        return nil, fmt.Errorf("failed to migrate table: %w", err)
    }

    return &DockerMock{{UpperCamel $.Table.Name}}Adapter{
        db:        db,
        container: container,
    }, nil
}

func (m *DockerMock{{UpperCamel $.Table.Name}}Adapter) DB(ctx context.Context) *gorm.DB {
    return m.db.WithContext(ctx)
}

func (m *DockerMock{{UpperCamel $.Table.Name}}Adapter) GetByID(ctx context.Context, id int64) (*entity.{{UpperCamel $.Table.Name}}, error) {
    var result entity.{{UpperCamel $.Table.Name}}
    err := m.db.WithContext(ctx).First(&result, id).Error
    if err != nil {
        return nil, err
    }
    return &result, nil
}

func (m *DockerMock{{UpperCamel $.Table.Name}}Adapter) Create(ctx context.Context, data ...*entity.{{UpperCamel $.Table.Name}}) error {
    return m.db.WithContext(ctx).Create(&data).Error
}

func (m *DockerMock{{UpperCamel $.Table.Name}}Adapter) List(ctx context.Context, query *gormx.Query) ([]*entity.{{UpperCamel $.Table.Name}}, error) {
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

func (m *DockerMock{{UpperCamel $.Table.Name}}Adapter) Count(ctx context.Context, query *gormx.Query) (int64, error) {
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

func (m *DockerMock{{UpperCamel $.Table.Name}}Adapter) Update(ctx context.Context, e *entity.{{UpperCamel $.Table.Name}}) error {
    return m.db.WithContext(ctx).Save(e).Error
}

func (m *DockerMock{{UpperCamel $.Table.Name}}Adapter) Delete(ctx context.Context, id int64) error {
    return m.db.WithContext(ctx).Delete(&entity.{{UpperCamel $.Table.Name}}{}, id).Error
}

func (m *DockerMock{{UpperCamel $.Table.Name}}Adapter) IsDuplicatedKeyError(err error) bool {
    return errors.Is(err, gorm.ErrDuplicatedKey)
}

func (m *DockerMock{{UpperCamel $.Table.Name}}Adapter) IsNotFoundError(err error) bool {
    return errors.Is(err, gorm.ErrRecordNotFound)
}

func (m *DockerMock{{UpperCamel $.Table.Name}}Adapter) Close() error {
    if m.container != nil {
        if err := m.container.Terminate(context.Background()); err != nil {
            return fmt.Errorf("failed to terminate container: %w", err)
        }
    }

    if m.db != nil {
        sqlDB, err := m.db.DB()
        if err != nil {
            return fmt.Errorf("failed to get underlying db: %w", err)
        }
        return sqlDB.Close()
    }
    return nil
} 