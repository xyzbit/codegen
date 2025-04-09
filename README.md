# Codegen
代码生成工具系列

## 安装
go install github.com/xyzbit/codegen@latest

## errcode


## dbrepo
> 用于生成 repo 层代码

1. 进入 example 目录
```shell
   cd sqlgen/example
```
2. 生成 repo 层代码: xx_adpter.go(接口实现)、xx_repo.go(接口定义)、xx_entity(实体)
   
- 使用配置文件
```shell 
codegen dbrepo gorm -c sqlgen.yaml
```

- 混合使用配置文件和命令行参数
```shell 
codegen dbrepo gorm -c sqlgen.yaml --mock-type sqlite
```

- 纯命令行方式
```shell 
codegen dbrepo gorm -f ./testdata/user.sql --table "user*" \
   --output './data' \
   --repo-output './service' \
   --entity-output './entity' \
   --repo-package 'github.com/xyzbit/codegen/sqlgen/example/service' \
   --entity-package 'github.com/xyzbit/codegen/sqlgen/example/entity'
   ```
   
3. Mock 代码生成
```shell
codegen dbrepo gorm -c sqlgen.yaml --mock-type sqlite --mock-type docker
```

生成文件的文件在如下地址(文件已存则不会重复生成)

4. 如何使用？
   ```go
    import "github.com/xyzbit/gpkg/gormx"
    // ....
    
    // 列表查询
    users, err := s.usersRepo.List(ctx, gormx.NewQuery().Eq(entity.UserNickName, "lee"))

    // 事务
    err := gormx.Transaction(ctx, repo, func(txCtx context.Context) error {
		user, err := userRepo.Create(txCtx, &User{UserNickName: "test"})
		if err != nil {
			return err
		}

		err = logRepo.Update(txCtx, &Log{Content: user.Name})
		if err != nil {
			return err
		}
		return nil
	})

   ```