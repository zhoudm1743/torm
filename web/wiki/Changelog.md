# 更新日志

本文档记录了 TORM 项目的所有重要变更。

## [v1.2.12] - 2024-12-20

🎯 **重大重构版本！架构完全重写**

### ✨ 核心重构

#### 🏗️ 架构重构
- **深度封装**: 对 database/sql 进行深度封装
- **统一连接管理**: 重新设计 DbManager 连接管理器
- **SQL 构建器分离**: 独立的 MySQL、PostgreSQL、SQLite 构建器
- **模型系统重构**: 完整的 BaseModel 实现

#### 🔗 数据库连接
- **多连接支持**: 支持注册和管理多个命名数据库连接
- **连接池优化**: 可配置的最大连接数、空闲连接数、连接生命周期
- **读写分离**: 主从配置、负载均衡支持
- **优雅关闭**: CloseAll 方法关闭所有活动连接

#### 🔍 查询构建器
- **链式调用**: 流畅的 API 设计
- **参数化查询**: 完全支持占位符，防止 SQL 注入
- **数组参数**: 原生支持 IN (?, ?, ?) 数组参数展开
- **跨数据库兼容**: 自动适配 MySQL(?) 和 PostgreSQL($N) 占位符

#### 📊 模型系统
- **BaseModel**: 完整的基础模型实现
- **时间戳管理**: 自动 created_at/updated_at
- **软删除**: SoftDelete 支持
- **访问器/修改器**: 字段获取器和设置器
- **JSON 字段**: 自动序列化/反序列化

#### 🔗 关联关系
- **HasOne**: 一对一关联
- **HasMany**: 一对多关联
- **BelongsTo**: 反向关联
- **BelongsToMany**: 多对多关联（中间表）
- **预加载**: With 方法支持嵌套预加载

#### 💼 事务支持
- **自动事务**: Transaction 回调自动提交/回滚
- **嵌套事务**: Savepoint 支持
- **手动控制**: Begin/Commit/Rollback

#### 💾 缓存系统
- **内存缓存**: 高性能内存缓存实现
- **缓存标签**: 支持按标签清除缓存
- **Redis 适配器**: 可选的 Redis 缓存支持

#### 🔄 数据库迁移
- **AutoMigrate**: 根据模型自动创建/更新表
- **智能差异检测**: 只更新必要的字段
- **跨数据库 DDL**: 自动生成数据库特定的 DDL

#### 🌐 上下文支持
- **WithContext**: 传递 context.Context
- **WithTimeout**: 设置操作超时
- **取消支持**: 查询取消和超时错误处理

### 🔧 技术改进
- **错误处理**: 完整的错误码和上下文信息
- **日志系统**: 可配置的日志级别和 SQL 日志
- **属性测试**: 14 个核心属性的 PBT 测试覆盖

---

## [v1.3.0] - 2025-08-22

🎯 **实现所有核心查询功能**

### ✨ 新增功能

#### 🔍 增强 WHERE 查询方法
- **WhereNull/WhereNotNull**: 支持 NULL 值查询判断
- **WhereBetween/WhereNotBetween**: 支持范围查询，自动参数绑定
- **WhereExists/WhereNotExists**: 支持子查询存在性检查
- **完整模型支持**: 所有新方法均支持模型链式调用

```go
// 使用示例
user.WhereNotNull("email").
    WhereBetween("age", []interface{}{18, 65}).
    WhereExists("SELECT 1 FROM orders WHERE orders.user_id = users.id").
    Get()
```

#### 📊 高级排序功能
- **OrderRand**: 跨数据库随机排序，自动适配 MySQL/PostgreSQL/SQLite
- **OrderField**: 按字段值优先级排序
- **FieldRaw**: 原生字段表达式

```go
// 使用示例
query.OrderField("status", []interface{}{"premium", "active", "trial"}, "asc").
    OrderRand()
```

---

## [v1.1.0] - 2025-08-10

🚀 **重大功能更新！实现高级功能**

### ✨ 新增功能

#### 🔗 关联预加载系统
- **预加载管理器**: 解决 N+1 查询问题
- **模型集合**: 支持批量关联数据预载入
- **条件预载入**: 支持为关联查询添加自定义条件

```go
// 使用示例
collection := model.NewModelCollection(users)
collection.With("profile", "posts").Load(ctx)
```

#### 📄 分页器系统
- **简单分页器**: 标准分页功能
- **游标分页器**: 适用于大数据量的高性能分页
- **JSON 序列化**: 完整的分页信息输出

```go
// 使用示例
result, _ := query.Paginate(1, 20)
```

---

## [v1.0.0] - 2024-01-10

🎉 **首个正式版本发布！**

### ✨ 核心功能

- 完整的数据库连接管理系统
- 支持 MySQL、PostgreSQL、SQLite
- 查询构建器
- 模型系统
- 事务处理
- 缓存系统
- 日志系统
- 数据迁移

### 📚 文档

- 完整的 GitHub Wiki 文档
- 详细的 API 参考文档
- 丰富的使用示例

---

## 🔄 升级指南

### 从 v1.1.x 升级到 v1.2.x

```bash
go get github.com/zhoudm1743/torm@v1.2.12
go mod tidy
```

### 从 v1.0.x 升级到 v1.1.x

```bash
go get github.com/zhoudm1743/torm@v1.1.0
go mod tidy
```

---

## 📋 版本规划

### v1.3.0 (计划中)
- [ ] 分布式事务支持
- [ ] 读写分离增强

### v2.0.0 (长期规划)
- [ ] 代码生成工具
- [ ] 图形化管理界面
- [ ] 云原生支持

---

**📞 如有升级问题，请查看 [故障排除文档](Troubleshooting) 或联系 zhoudm1743@163.com**
