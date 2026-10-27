# protoc-gen-sgorm

把 proto 消息生成 GORM 实体（**sgorm** = simple gorm）的 protoc 插件。

每个带 `option (sgorm.opts).ormable = true;` 的消息生成三样东西：

1. `<Msg>ORM` 结构体（含 `gorm:` tag，注释从 proto 抄过来：字段/消息的**前导注释**优先，
   没有就用**行尾注释**——两种写法在 proto 里都很常见）；
2. `func (<Msg>ORM) TableName() string`；
3. `func From<Msg>(ctx, *native.<Msg>) (<Msg>ORM, error)` 与 `func (m *<Msg>ORM) ToPB(ctx) (*native.<Msg>, error)`。

**不生成** CRUD、hook、关联装配、默认查询方法——那些由业务代码自己写。这是与
[protoc-gen-gorm](https://github.com/infobloxopen/protoc-gen-gorm) 的唯一区别：借它的写法，
砍掉用不上的部分。

> **仓库**：`git@github.com:jenken827/protoc-gen-sgorm.git`
>
> 模块路径就是 `github.com/jenken827/protoc-gen-sgorm`：`sgorm/options.proto` 的 `go_package`、
> 示例的 import、下面的安装命令都以它为准。换托管地址时这三处必须一起改。

## 为什么不直接用 pb 类型当 GORM 模型

| 问题 | 说明 |
|---|---|
| 时间字段 | `*timestamppb.Timestamp` 没有 `driver.Valuer` / `sql.Scanner`，且不能给外部包类型加方法 |
| 没有 tag 位 | 主键、列名、索引、唯一约束、软删、默认值全都没地方写；GORM 的隐式主键只认字段名 `ID`，而 pb 生成的是 `Id` |
| 不能加方法 | `TableName()`、hook 都加不了（Go 不允许给非本包类型加方法） |
| repeated | 切片会被 GORM 当成 has-many 关系，而库里往往只是关联表、没有这一列 |

所以实体结构体必须手写或生成——**这个插件负责生成**，你只需要在 proto 里写 tag。

## 安装

```bash
go install github.com/jenken827/protoc-gen-sgorm/cmd/protoc-gen-sgorm@latest
```

## 使用

### 1. 引入选项定义

把本仓库的 [`sgorm/options.proto`](sgorm/options.proto) 复制到项目的 proto 根目录
（与 `buf.yaml` 的模块根同级），保持 import 路径 `sgorm/options.proto` 不变。

### 2. 在 proto 里标注

```proto
import "sgorm/options.proto";

// Transaction 交易记录表
message Transaction {
  option (sgorm.opts).ormable = true;

  optional uint32 id = 1 [(sgorm.field).tag = {primary_key: true, auto_increment: true}];
  optional google.protobuf.Timestamp created_at = 2 [(sgorm.field).tag = {auto_create_time: true}];
  optional google.protobuf.Timestamp deleted_at = 3 [(sgorm.field).tag = {type: "deleted_at"}];
  // 金额，单位为分
  optional int64 amount = 4 [ (sgorm.field).tag = {not_null: true, default: "0"}];
  // 库里列名不一样
  optional string user_id = 5 [(sgorm.field).tag = {not_null: true, column: "owner_id"}];
  // 只给接口用，不落库
  optional string note = 6 [(sgorm.field).ignore = true];
  // 走关联表，没有对应列
  repeated uint32 tag_ids = 7;
}
```

### 3. 在 buf.gen.yaml 里加一个插件条目

```yaml
version: v2
plugins:
  - local: protoc-gen-go                 # 原有的 pb 生成
    out: .
    opt: module=example.com/project
  - local: protoc-gen-sgorm              # 新增：只生成实体
    out: ./internal/entity
    opt: package=entity                  # 生成文件的 Go 包名
```

生成的文件名是 `<proto 基名>.sgorm.go`，直接落在 `out` 目录里（不会再套一层目录）。
没有 `ormable` 消息的 proto 不产生文件。

### 4. 用起来

```go
orm, err := entity.FromTransaction(ctx, req)      // 请求 → 实体
pb, err := orm.ToPB(ctx)                          // 实体 → 响应
```

## 选项

消息级 `(sgorm.opts)`：

| 字段 | 含义 |
|---|---|
| `ormable` | `true` 才生成实体 |
| `table` | 表名覆盖；默认消息名转蛇形 + 复数（`Category` → `categories`） |

字段级 `(sgorm.field)`：

| 字段 | 含义 |
|---|---|
| `ignore` | 不生成到实体 |
| `tag.primary_key` / `auto_increment` / `not_null` / `unique` | 对应 gorm tag |
| `tag.column` | 覆盖列名 |
| `tag.type` | 列类型（`uuid` / `jsonb` / `bytea`…）；特殊值 `deleted_at` 生成 `*gorm.DeletedAt` |
| `tag.size` / `precision` | 尺寸与精度 |
| `tag.default` | 默认值，原样写入 |
| `tag.index` / `unique_index` | 索引名，可带参数（`idx_f_npoi,length:10`） |
| `tag.auto_create_time` / `auto_update_time` | 只对 Timestamp 字段生效 |
| `tag.serializer` | 如 `json` |

类型映射：

| proto | 实体 |
|---|---|
| `optional string/bool/整数/浮点` | `*T`（无 presence 时是 `T`） |
| `optional 枚举` | `*int32` |
| `google.protobuf.Timestamp` | `*time.Time` |
| Timestamp + `type: "deleted_at"` | `*gorm.DeletedAt` |
| `bytes` | `[]byte` |
| `repeated` / `map` / 其它嵌套消息 | **跳过**（需要时用 `ignore` 显式声明） |

## 已知限制

- 文件名取 proto 基名：不同目录下的同名 proto（如两个 `file.proto`）会撞名，请用 `--sgorm_opt` 的
  `suffix` 参数或拆目录。
- 不支持嵌套消息、oneof、关联自动装配——这些本来就该手写。
- 表名复数化只处理常见规则；不规则表名用 `table` 覆盖。

## 开发

```bash
bash dev.sh gen     # 编译插件 + 用 example/ 里的 proto 生成示例代码
bash dev.sh test    # gen + go vet + go build + go test
bash dev.sh fmt     # gofmt
```

目录：

| 路径 | 说明 |
|---|---|
| `cmd/protoc-gen-sgorm/` | 插件入口（protogen + 参数） |
| `internal/gen/` | 生成逻辑（模型、tag、类型映射、转换代码、命名） |
| `sgorm/options.proto` | 选项定义（使用者需要复制这一份） |
| `example/proto/` | 示例 proto，覆盖各种字段形态 |
| `example/gen/entity/` | 示例生成物（`dev.sh gen` 产物） |
| `example/entity_test.go` | 用 GORM 的 schema 解析器校验 tag + From/ToPB 往返 |
