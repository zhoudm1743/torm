# 更新日志

本文档记录了TORM项目的所有重要变更。

## [v1.1.6] - 2025-08-22

🎯 **完美对标ThinkORM！实现所有核心查询功能**

### ✨ 新增功能

#### 🔍 增强WHERE查询方法
- **WhereNull/WhereNotNull**: 支持NULL值查询判断
- **WhereBetween/WhereNotBetween**: 支持范围查询，自动参数绑定
- **WhereExists/WhereNotExists**: 支持子查询存在性检查，完整SQL注入防护
- **完整模型支持**: 所有新方法均支持模型链式调用

```go
// 使用示例
user.WhereNotNull("email").
    WhereBetween("age", []interface{}{18, 65}).
    WhereExists("SELECT 1 FROM orders WHERE orders.user_id = users.id").
    Get()
```

#### 📊 高级排序功能
- **OrderRand**: 跨数据库随机排序，自动适配MySQL/PostgreSQL/SQLite
- **OrderField**: 按字段值优先级排序，使用CASE WHEN优化
- **FieldRaw**: 原生字段表达式，支持复杂聚合函数
- **数据库兼容**: 自动检测数据库类型，使用最优语法

```go
// 使用示例
query.OrderField("status", []interface{}{"premium", "active", "trial"}, "asc").
    OrderRand().
    FieldRaw("COUNT(*) as total_count")
```

#### ⚡ 性能优化
- **SQL构建优化**: 增强SQL构建器，支持复杂查询条件组合
- **接口统一**: 查询构建器和模型接口完全统一

### 🔧 技术改进
- **完整ThinkORM兼容**: 实现了ThinkORM的所有主要查询方法
- **代码质量**: 新增完整测试覆盖，包含边界情况处理
- **文档更新**: 全面更新API文档和示例代码

### 🐛 修复
- **接口一致性**: 修复WhereBetween方法签名不一致问题
- **SQL注入防护**: 强化所有新增方法的参数绑定和验证

## [v1.1.0] - 2025-08-10

🚀 **重大功能更新！实现高级功能**

### ✨ 新增功能

#### 🔗 关联预加载系统
- **预加载管理器 (EagerLoadManager)**: 解决 N+1 查询问题
- **模型集合 (ModelCollection)**: 支持批量关联数据预载入
- **条件预载入**: 支持为关联查询添加自定义条件
- **字段限制**: 可选择性预载入关联模型的特定字段
- **缓存支持**: 关联数据缓存，提升重复查询性能

```go
// 使用示例
collection := model.NewModelCollection(users)
collection.With("profile", "posts").
    WithClosure("posts", func(q db.QueryInterface) db.QueryInterface {
        return q.Where("status", "=", "published").Limit(5)
    })
err := collection.Load(ctx)
```

#### 📄 分页器系统
- **简单分页器 (SimplePaginator)**: 标准分页功能
- **游标分页器 (CursorPaginator)**: 适用于大数据量的高性能分页
- **查询分页器 (QueryPaginator)**: 无缝集成到查询构造器
- **URL生成**: 自动生成上一页/下一页链接
- **JSON序列化**: 完整的分页信息输出

```go
// 使用示例
paginator := paginator.NewQueryPaginator(query, ctx)
result, err := paginator.SetPerPage(10).SetPage(1).Paginate()
```

#### 🔍 JSON查询支持
- **跨数据库JSON查询**: 支持 MySQL、PostgreSQL、SQLite 的 JSON 字段查询
- **JSON路径查询**: 使用 JSONPath 语法查询嵌套数据
- **JSON包含查询**: 检查 JSON 字段是否包含特定值
- **JSON数组操作**: 支持数组长度、元素查询等
- **智能降级**: 不支持JSON的数据库自动使用兼容语法

```go
// 使用示例
advQuery := query.NewAdvancedQueryBuilder(baseQuery)
result := advQuery.
    WhereJSON("metadata", "$.age", ">", 18).
    WhereJSONContains("skills", "$.language", "Go").
    WhereJSONLength("certifications", "$", ">=", 1)
```

#### 🏗️ 高级查询功能
- **子查询支持**: EXISTS、NOT EXISTS、IN、NOT IN 子查询
- **窗口函数**: ROW_NUMBER、RANK、DENSE_RANK、LAG/LEAD
- **窗口聚合**: 分区计数、求和、平均值等
- **复杂条件组合**: 支持复杂的查询条件嵌套

```go
// 子查询示例
result := advQuery.WhereExists(func(q db.QueryInterface) db.QueryInterface {
    return q.Where("posts.user_id", "=", "users.id").
        Where("posts.status", "=", "published")
})

// 窗口函数示例
result := advQuery.
    WithRowNumber("row_num", "department", "salary DESC").
    WithRank("salary_rank", "department", "salary DESC")
```

### 🔧 技术改进

#### 接口完整性
- 为所有查询构建器实现了 `Paginate` 方法
- 修复了接口兼容性问题
- 统一了分页API设计

#### 性能优化
- 预加载减少数据库查询次数（解决N+1问题）
- 游标分页适用于大数据量场景
- 窗口函数提升统计查询性能

#### 代码质量
- 完整的单元测试覆盖
- 类型安全的接口设计
- 详细的错误处理

### 📚 文档更新

- 新增高级功能使用示例
- 更新API参考文档
- 添加性能优化指南
- 完善故障排除文档

### 🧪 测试覆盖

- 新增预加载功能测试
- 分页器功能完整测试
- JSON查询跨数据库测试
- 高级查询功能测试
- 保持95%+代码覆盖率

### 💡 使用示例

完整的高级功能演示请参考：
- `examples/advanced_features_demo.go` - 完整功能演示
- [Examples](Examples) - 更新的示例文档
- [Quick-Start](Quick-Start) - 快速开始指南

### ⚡ 性能提升

- **N+1 查询问题**: 通过预加载完全解决
- **大数据分页**: 游标分页性能提升90%+
- **复杂查询**: 窗口函数减少多次查询的需求
- **JSON查询**: 原生数据库JSON功能，性能优异

### 🔄 向后兼容

此版本完全向后兼容 v1.0.0，现有代码无需修改即可升级。

---

## [v1.0.0] - 2024-01-10

🎉 **首个正式版本发布！**

### ✨ 新增功能

#### 核心ORM功能
- 实现了完整的数据库连接管理系统
- 支持连接池配置和管理
- 实现了事务处理机制
- 添加了查询日志和性能监控

#### 多数据库支持
- ✅ **MySQL**: 完整支持 MySQL 5.7+ / 8.0+
- ✅ **PostgreSQL**: 完整支持 PostgreSQL 11+
- ✅ **SQLite**: 完整支持 SQLite 3.8+
- ✅ **MongoDB**: 完整支持 MongoDB 4.4+
- 🚧 **SQL Server**: 基础支持

#### 数据迁移系统
- 实现了数据迁移工具
- 支持版本化数据库结构管理
- 提供了强大的结构构建器（SchemaBuilder）
- 实现了批次管理和回滚功能
- 支持跨数据库的DDL生成

#### 查询构建器
- 实现了类型安全的SQL查询构建
- 支持复杂的WHERE条件组合
- 提供了JOIN、子查询、聚合等高级功能
- 实现了MongoDB查询构建器

#### 模型系统
- 实现了基础模型（BaseModel）
- 支持属性管理和脏检查
- 提供了时间戳自动管理
- 实现了软删除功能

#### 关联关系
- ✅ **HasOne**: 一对一关系
- ✅ **HasMany**: 一对多关系  
- ✅ **BelongsTo**: 属于关系
- ✅ **ManyToMany**: 多对多关系
- 使用反射实现动态方法调用

#### 缓存系统
- 实现了高性能内存缓存
- 支持TTL和自动清理
- 提供了缓存键管理

#### 日志系统
- 集成了结构化日志记录
- 支持查询日志和性能分析
- 提供了自定义日志器接口

### 🔧 技术特性

- **高并发**: 优化的连接池管理
- **类型安全**: 编译时类型检查
- **跨数据库**: 统一的API接口
- **高性能**: 平均比其他ORM快30%
- ****: 完整的迁移和事务支持

### 📚 文档和示例

- 完整的GitHub Wiki文档
- 详细的API参考文档
- 丰富的使用示例
- 快速开始指南
- 故障排除文档

### 🧪 测试覆盖

- 95%+ 代码覆盖率
- 完整的单元测试
- 集成测试
- 多数据库兼容性测试
- 性能基准测试

### 📦 依赖管理

```go
require (
    github.com/go-sql-driver/mysql v1.7.1
    github.com/lib/pq v1.10.9
    github.com/glebarez/sqlite v1.11.0
    go.mongodb.org/mongo-driver v1.17.4
    github.com/patrickmn/go-cache v2.1.0+incompatible
    github.com/sirupsen/logrus v1.9.3
    github.com/stretchr/testify v1.8.4
)
```

## [v0.9.0] - 2024-01-05

### ✨ 新增功能
- 实现MongoDB查询构建器
- 添加MongoDB事务支持
- 完善MongoDB连接管理

### 🐛 修复问题
- 修复MongoDB查询条件累积问题
- 修复事务回调函数类型错误
- 解决MongoDB连接适配器类型转换问题

### 🔧 改进
- 优化MongoDB查询性能
- 改进错误处理机制
- 完善测试覆盖率

## [v0.8.0] - 2024-01-03

### ✨ 新增功能
- 实现SQLite数据库支持
- 添加多数据库同时支持
- 创建跨数据库数据同步功能

### 🔧 改进
- 优化连接池配置
- 改进配置验证逻辑
- 完善错误信息

## [v0.7.0] - 2024-01-01

### ✨ 新增功能
- 实现PostgreSQL完整支持
- 添加PostgreSQL查询构建器
- 创建PostgreSQL演示程序

### 🔧 改进
- 统一查询构建器接口
- 优化PostgreSQL特定功能
- 完善文档和示例

## [v0.6.0] - 2023-12-28

### ✨ 新增功能
- 实现数据迁移系统核心功能
- 添加结构构建器（SchemaBuilder）
- 支持版本化迁移管理

### 🔧 改进
- 优化迁移执行性能
- 改进批次管理逻辑
- 完善回滚机制

## [v0.5.0] - 2023-12-25

### ✨ 新增功能
- 实现关联关系系统
- 添加HasOne、HasMany、BelongsTo关系
- 支持ManyToMany复杂关系

### 🔧 改进
- 优化关系查询性能
- 改进关联数据加载
- 完善关系定义语法

## [v0.4.0] - 2023-12-22

### ✨ 新增功能
- 实现模型系统
- 添加BaseModel基础模型
- 支持属性管理和脏检查

### 🔧 改进
- 优化模型操作性能
- 改进数据绑定机制
- 完善时间戳管理

## [v0.3.0] - 2023-12-20

### ✨ 新增功能
- 实现查询构建器
- 添加SQL查询生成
- 支持复杂查询条件

### 🔧 改进
- 优化查询生成性能
- 改进参数绑定安全性
- 完善查询缓存机制

## [v0.2.0] - 2023-12-18

### ✨ 新增功能
- 实现缓存系统
- 添加日志系统
- 支持查询性能监控

### 🔧 改进
- 优化缓存命中率
- 改进日志格式
- 完善性能分析工具

## [v0.1.0] - 2023-12-15

### ✨ 新增功能
- 初始版本发布
- 实现基础数据库连接
- 添加MySQL支持
- 实现事务处理

### 🔧 改进
- 建立项目基础架构
- 实现核心接口设计
- 添加基础测试框架

---

## 📋 版本规划

### v1.1.0 ✅ (已发布)
- [x] 查询构建器高级功能 (JSON查询、子查询、窗口函数)
- [x] 模型关系预加载优化 (EagerLoadManager、ModelCollection)
- [x] 分页器系统 (简单分页、游标分页)
- [x] 高级查询构建器 (AdvancedQueryBuilder)

### v1.2.0 (计划中)
- [ ] 分布式事务支持
- [ ] 读写分离

### v1.2.0 (计划中)  
- [ ] 数据库分片支持
- [ ] 查询缓存优化
- [ ] 性能监控仪表板
- [ ] 断点重连机制

### v2.0.0 (长期规划)
- [ ] 代码生成工具
- [ ] 图形化管理界面  
- [ ] 云原生支持
- [ ] 机器学习集成

---

## 🔄 升级指南

### 从 v0.x 升级到 v1.0

1. **更新依赖**:
```bash
go get github.com/zhoudm1743/torm@v1.0.0
go mod tidy
```

2. **配置迁移**:
```go
// 旧版本
config := db.Config{
    Driver: "mysql",
    DSN:    "user:pass@tcp(localhost:3306)/db",
}

// 新版本
config := &db.Config{
    Driver:   "mysql",
    Host:     "localhost", 
    Port:     3306,
    Database: "db",
    Username: "user",
    Password: "pass",
}
```

3. **API变更**:
- `db.Connect()` → `db.AddConnection()`
- `db.GetDB()` → `db.DB()`
- 迁移工具完全重写，请参考新的文档

---

**📞 如有升级问题，请查看 [故障排除文档](Troubleshooting) 或联系 zhoudm1743@163.com** 