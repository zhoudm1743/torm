# 查询构建器

TORM 提供了强大而现代化的查询构建器，支持参数化查询、数组参数自动展开、跨数据库占位符适配等革命性特性。

## 📋 目录

- [快速开始](#快速开始)
- [参数化查询](#参数化查询)
- [数组参数支持](#数组参数支持)
- [条件查询](#条件查询)
- [聚合查询](#聚合查询)
- [排序和分页](#排序和分页)
- [JOIN查询](#JOIN查询)
- [事务处理](#事务处理)
- [跨数据库兼容](#跨数据库兼容)

## 🚀 快速开始

### 创建查询构建器

```go
// 基础表查询
query, err := torm.Table("users")
if err != nil {
    log.Fatal(err)
}

// 指定连接
query, err := torm.Table("users", "mysql_connection")
if err != nil {
    log.Fatal(err)
}
```

### 基础查询

```go
// 查询所有记录 (Result系统，支持访问器)
users, err := torm.Table("users").Model(&User{}).Get()          // 返回 *ResultCollection

// 查询指定字段
users, err := torm.Table("users").
    Select("id", "name", "email").
    Model(&User{}).
    Get()                                                       // 返回 *ResultCollection

// 查询单条记录
user, err := torm.Table("users").
    Where("id", "=", 1).
    Model(&User{}).
    First()                                                     // 返回 *Result

// 原始数据查询 (向下兼容，高性能)
rawUsers, err := torm.Table("users").GetRaw()                  // 返回 []map[string]interface{}
rawUser, err := torm.Table("users").
    Where("id", "=", 1).
    FirstRaw()                                                  // 返回 map[string]interface{}

// 检查记录是否存在
exists, err := torm.Table("users").
    Where("email", "=", "user@example.com").
    Exists()
```

## 🔒 参数化查询

```go
// 单参数查询
user, err := torm.Table("users").
    Where("username = ?", "zhangsan").
    First()

// 多参数查询
users, err := torm.Table("users").
    Where("age >= ? AND status = ?", 18, "active").
    Get()

// 复杂条件组合
users, err := torm.Table("users").
    Where("(status = ? OR vip_level > ?) AND created_at > ?", 
          "premium", 3, "2024-01-01").
    Get()

// LIKE 查询
users, err := torm.Table("users").
    Where("name LIKE ?", "%张%").
    Get()

// BETWEEN 查询
users, err := torm.Table("users").
    Where("age BETWEEN ? AND ?", 18, 65).
    Get()
```

### 与传统查询对比

```go
// ✅ 推荐：参数化查询（安全）
users, err := torm.Table("users").
    Where("status = ? AND age >= ?", "active", 18).
    Get()

// ✅ 仍然支持：传统三参数查询
users, err := torm.Table("users").
    Where("status", "=", "active").
    Where("age", ">=", 18).
    Get()

// ❌ 危险：字符串拼接（易受SQL注入攻击）
// 永远不要这样做！
// sql := "SELECT * FROM users WHERE name = '" + userInput + "'"
```

## 🎯 数组参数支持

### 自动数组展开

```go
// 字符串数组
activeUsers, err := torm.Table("users").
    Where("status IN (?)", []string{"active", "premium", "vip"}).
    Get()

// 整数数组
usersByIds, err := torm.Table("users").
    Where("id IN (?)", []int{1, 2, 3, 4, 5}).
    Get()

// 混合类型数组
results, err := torm.Table("orders").
    Where("status IN (?)", []string{"completed", "shipped"}).
    Where("user_id IN (?)", []int64{100, 200, 300}).
    Get()

// 复杂数组查询
complexResults, err := torm.Table("products").
    Where("category IN (?)", []string{"electronics", "books"}).
    Where("price BETWEEN ? AND ?", 10.00, 100.00).
    Where("brand_id IN (?)", []int{1, 3, 5, 7}).
    Get()
```

### 支持的数组类型

```go
// 所有基础类型数组都支持
strings := []string{"a", "b", "c"}
ints := []int{1, 2, 3}
int64s := []int64{100, 200, 300}
floats := []float64{1.1, 2.2, 3.3}
bools := []bool{true, false}

// 使用示例
torm.Table("table").Where("string_field IN (?)", strings).Get()
torm.Table("table").Where("int_field IN (?)", ints).Get()
torm.Table("table").Where("bigint_field IN (?)", int64s).Get()
torm.Table("table").Where("float_field IN (?)", floats).Get()
```

## 🔍 条件查询

### 基础条件

```go
// 等于
torm.Table("users").Where("status", "=", "active")
torm.Table("users").Where("status = ?", "active")

// 比较操作
torm.Table("users").Where("age", ">=", 18)
torm.Table("users").Where("age >= ?", 18)

// LIKE 模糊查询
torm.Table("users").Where("name", "LIKE", "%张%")
torm.Table("users").Where("name LIKE ?", "%张%")

// NULL 查询
torm.Table("users").Where("deleted_at", "IS", nil)
torm.Table("users").Where("deleted_at IS NULL")
```

### OR 条件

```go
// 基础 OR 查询
users, err := torm.Table("users").
    Where("status", "=", "active").
    OrWhere("vip_level", ">", 3).
    Get()

// 参数化 OR 查询
users, err := torm.Table("users").
    Where("status = ?", "active").
    OrWhere("vip_level > ?", 3).
    Get()

// 复杂 OR 条件组合
users, err := torm.Table("users").
    Where("(status = ? OR vip_level > ?) AND age >= ?", 
          "premium", 3, 18).
    Get()
```

### 高级条件

```go
// BETWEEN 范围查询
users, err := torm.Table("users").
    Where("age BETWEEN ? AND ?", 18, 65).
    Where("created_at BETWEEN ? AND ?", "2024-01-01", "2024-12-31").
    Get()

// IN 查询（传统方式）
users, err := torm.Table("users").
    WhereIn("status", []interface{}{"active", "premium"}).
    Get()

// IN 查询（参数化方式，推荐）
users, err := torm.Table("users").
    Where("status IN (?)", []string{"active", "premium"}).
    Get()

// NOT IN 查询
users, err := torm.Table("users").
    Where("status NOT IN (?)", []string{"deleted", "banned"}).
    Get()

// EXISTS 子查询
users, err := torm.Table("users").
    Where("EXISTS (SELECT 1 FROM orders WHERE orders.user_id = users.id)").
    Get()
```

## 📊 聚合查询

### 基础聚合

```go
// 计数
count, err := torm.Table("users").
    Where("status", "=", "active").
    Count()

// 分组计数
results, err := torm.Table("users").
    Select("status", "COUNT(*) as count").
    GroupBy("status").
    Get()

// 带 HAVING 的分组
results, err := torm.Table("users").
    Select("city", "COUNT(*) as user_count").
    GroupBy("city").
    Having("user_count", ">", 10).
    Get()
```

### 分页查询

```go
// 基础分页
result, err := torm.Table("users").
    Where("status", "=", "active").
    Paginate(1, 20) // 第1页，每页20条

// 分页结果包含
// result.Data - 当前页数据
// result.Total - 总记录数  
// result.Page - 当前页码
// result.PerPage - 每页数量
// result.LastPage - 最后一页
```

## 📋 排序和分页

### 排序

```go
// 单字段排序
users, err := torm.Table("users").
    OrderBy("created_at", "desc").
    Get()

// 多字段排序
users, err := torm.Table("users").
    OrderBy("status", "asc").
    OrderBy("created_at", "desc").
    Get()

// 原生排序表达式
users, err := torm.Table("users").
    OrderByRaw("FIELD(status, ?, ?, ?)", "premium", "active", "trial").
    Get()
```

### 限制和偏移

```go
// 基础限制
users, err := torm.Table("users").
    Limit(10).
    Get()

// 限制和偏移
users, err := torm.Table("users").
    Limit(10).
    Offset(20).
    Get()

// 结合排序
users, err := torm.Table("users").
    Where("status", "=", "active").
    OrderBy("created_at", "desc").
    Limit(10).
    Get()
```

## 🔗 JOIN查询

### 基础 JOIN

```go
// INNER JOIN
users, err := torm.Table("users").
    Join("profiles", "profiles.user_id", "=", "users.id").
    Select("users.name", "profiles.avatar").
    Get()

// LEFT JOIN
users, err := torm.Table("users").
    LeftJoin("profiles", "profiles.user_id", "=", "users.id").
    Select("users.*", "profiles.avatar").
    Get()

// RIGHT JOIN
users, err := torm.Table("users").
    RightJoin("orders", "orders.user_id", "=", "users.id").
    Select("users.name", "orders.total").
    Get()
```

### 复杂 JOIN

```go
// 多表 JOIN
results, err := torm.Table("users").
    LeftJoin("profiles", "profiles.user_id", "=", "users.id").
    LeftJoin("orders", "orders.user_id", "=", "users.id").
    Select("users.name", "profiles.avatar", "COUNT(orders.id) as order_count").
    GroupBy("users.id").
    Get()

// 带条件的 JOIN
results, err := torm.Table("users").
    LeftJoin("orders", "orders.user_id", "=", "users.id").
    Where("users.status", "=", "active").
    Where("orders.status", "=", "completed").
    Select("users.name", "SUM(orders.total) as total_spent").
    GroupBy("users.id").
    Get()
```

## 💼 事务处理

### 自动事务管理

```go
// TORM 事务API
err := torm.Transaction(func(tx torm.TransactionInterface) error {
    // 在事务中执行多个操作
    _, err := tx.Exec("INSERT INTO users (name, email) VALUES (?, ?)", 
                     "张三", "zhangsan@example.com")
    if err != nil {
        return err // 自动回滚
    }

    _, err = tx.Exec("INSERT INTO profiles (user_id, avatar) VALUES (?, ?)", 
                    1, "avatar.jpg")
    if err != nil {
        return err // 自动回滚
    }

    return nil // 自动提交
})

if err != nil {
    log.Printf("事务失败: %v", err)
}
```

### 复杂事务示例

```go
// 银行转账事务
func transferMoney(fromUserID, toUserID int64, amount float64) error {
    return torm.Transaction(func(tx torm.TransactionInterface) error {
        // 检查发送方余额
        var fromBalance float64
        row := tx.QueryRow("SELECT balance FROM accounts WHERE user_id = ?", fromUserID)
        if err := row.Scan(&fromBalance); err != nil {
            return err
        }
        
        if fromBalance < amount {
            return fmt.Errorf("余额不足")
        }
        
        // 扣除发送方余额
        _, err := tx.Exec("UPDATE accounts SET balance = balance - ? WHERE user_id = ?", 
                         amount, fromUserID)
        if err != nil {
            return err
        }
        
        // 增加接收方余额
        _, err = tx.Exec("UPDATE accounts SET balance = balance + ? WHERE user_id = ?", 
                        amount, toUserID)
        if err != nil {
            return err
        }
        
        // 记录转账日志
        _, err = tx.Exec("INSERT INTO transfer_logs (from_user, to_user, amount) VALUES (?, ?, ?)", 
                        fromUserID, toUserID, amount)
        if err != nil {
            return err
        }
        
        return nil // 自动提交所有操作
    })
}
```

## 🌐 跨数据库兼容

### 自动占位符适配

```go
// 相同的Go代码
users, err := torm.Table("users").
    Where("status = ? AND age >= ?", "active", 18).
    Get()

// 自动生成的SQL：
// MySQL:      SELECT * FROM users WHERE status = ? AND age >= ?
// PostgreSQL: SELECT * FROM users WHERE status = $1 AND age >= $2  
// SQLite:     SELECT * FROM users WHERE status = ? AND age >= ?
```

### 数据库特定功能

```go
// JSON 查询（跨数据库兼容）
users, err := torm.Table("users").
    Where("settings->>'theme' = ?", "dark").  // MySQL/PostgreSQL
    Get()

// 全文搜索（自动适配）
users, err := torm.Table("users").
    Where("MATCH(name, bio) AGAINST(?)", "golang developer").  // MySQL
    Get()

// 日期函数（自动适配）
users, err := torm.Table("users").
    Where("DATE(created_at) = ?", "2024-01-01").
    Get()
```

### 连接切换

```go
// 同一查询，不同数据库
query := torm.Table("users").Where("status", "=", "active")

// MySQL
mysqlUsers, err := query.Connection("mysql").Get()

// PostgreSQL  
postgresUsers, err := query.Connection("postgres").Get()

// SQLite
sqliteUsers, err := query.Connection("sqlite").Get()
```

## 📝 CRUD操作

### INSERT 操作

```go
// 单条插入
id, err := torm.Table("users").Insert(map[string]interface{}{
    "name":  "张三",
    "email": "zhangsan@example.com",
    "age":   25,
})

// 批量插入
users := []map[string]interface{}{
    {"name": "李四", "email": "lisi@example.com", "age": 30},
    {"name": "王五", "email": "wangwu@example.com", "age": 28},
}
count, err := torm.Table("users").InsertBatch(users)
```

### UPDATE 操作

```go
// 基础更新
affected, err := torm.Table("users").
    Where("id", "=", 1).
    Update(map[string]interface{}{
        "name": "新名字",
        "age":  26,
    })

// 条件更新
affected, err := torm.Table("users").
    Where("status = ? AND last_login < ?", "inactive", "2023-01-01").
    Update(map[string]interface{}{
        "status": "archived",
    })

// 参数化更新
affected, err := torm.Table("users").
    Where("email = ?", "user@example.com").
    Update(map[string]interface{}{
        "name": "更新的名字",
        "updated_at": time.Now(),
    })
```

### DELETE 操作

```go
// 条件删除
affected, err := torm.Table("users").
    Where("status", "=", "deleted").
    Delete()

// 参数化删除
affected, err := torm.Table("users").
    Where("created_at < ? AND status = ?", "2022-01-01", "inactive").
    Delete()

// 批量删除
affected, err := torm.Table("users").
    Where("id IN (?)", []int{1, 2, 3, 4, 5}).
    Delete()
```

## 🔧 高级功能

### 原生SQL查询

```go
// 原生 SELECT
users, err := torm.Raw("SELECT * FROM users WHERE age > ? AND city = ?", 
                      18, "北京")

// 原生 INSERT
result, err := torm.Exec("INSERT INTO users (name, email) VALUES (?, ?)", 
                        "张三", "zhangsan@example.com")

// 原生查询与构建器结合
users, err := torm.Table("users").
    WhereRaw("YEAR(created_at) = ?", 2024).
    OrderByRaw("FIELD(status, ?, ?, ?)", "active", "pending", "inactive").
    Get()
```

### 查询调试

```go
// 获取生成的SQL（不执行）
sql, bindings := torm.Table("users").
    Where("status = ? AND age >= ?", "active", 18).
    ToSQL()

fmt.Printf("SQL: %s\n", sql)
fmt.Printf("参数: %v\n", bindings)
// 输出: SQL: SELECT * FROM users WHERE status = ? AND age >= ?
// 输出: 参数: [active 18]
```

### 性能优化

```go
// 只查询需要的字段
users, err := torm.Table("users").
    Select("id", "name", "email").
    Where("status", "=", "active").
    Get()

// 使用索引优化
users, err := torm.Table("users").
    Where("email", "=", email).  // email 字段应该有索引
    Where("status", "=", "active").
    Get()

// 分页避免大数据量
for page := 1; ; page++ {
    result, err := torm.Table("users").
        Where("status", "=", "active").
        Paginate(page, 100)
    
    if err != nil || len(result.Data) == 0 {
        break
    }
    
    // 处理当前页数据
    processBatch(result.Data)
}
```

## 📚 最佳实践

### 1. 安全性

```go
// ✅ 推荐：使用参数化查询
users, err := torm.Table("users").
    Where("name = ? AND age >= ?", userInput, minAge).
    Get()

// ❌ 危险：字符串拼接
// sql := "SELECT * FROM users WHERE name = '" + userInput + "'"
```

### 2. 性能优化

```go
// ✅ 推荐：利用数据库索引
users, err := torm.Table("users").
    Where("email", "=", email).      // email 应该有唯一索引
    Where("status", "=", "active").  // status 可以是复合索引
    Get()

// ✅ 推荐：只查询需要的字段
users, err := torm.Table("users").
    Select("id", "name", "email").
    Where("status", "=", "active").
    Get()
```

### 3. 数组参数

```go
// ✅ 推荐：使用数组参数
users, err := torm.Table("users").
    Where("status IN (?)", []string{"active", "premium"}).
    Where("id IN (?)", userIds).
    Get()

// ✅ 也支持：传统方式
users, err := torm.Table("users").
    WhereIn("status", []interface{}{"active", "premium"}).
    WhereIn("id", userIds).
    Get()
```

### 4. 事务使用

```go
// ✅ 推荐：使用自动事务管理
err := torm.Transaction(func(tx torm.TransactionInterface) error {
    // 所有数据库操作
    return performDatabaseOperations(tx)
})
```

## 🎨 Result 系统

### 访问器支持

TORM v2.0 引入了强大的 Result 系统，支持 ThinkPHP 风格的访问器/修改器：

```go
// 定义模型和访问器
type User struct {
    model.BaseModel
    ID       int    `json:"id"`
    Status   int    `json:"status"`
    Salary   int    `json:"salary"`  // 以分为单位存储
}

// 状态访问器
func (u *User) GetStatusAttr(value interface{}) interface{} {
    status := value.(int)
    statusMap := map[int]string{0: "禁用", 1: "正常", 2: "待审核"}
    return map[string]interface{}{
        "code": status,
        "name": statusMap[status],
        "color": []string{"red", "green", "orange"}[status],
    }
}

// 薪资访问器（分转元）
func (u *User) GetSalaryAttr(value interface{}) interface{} {
    cents := value.(int)
    yuan := float64(cents) / 100.0
    return map[string]interface{}{
        "cents":     cents,
        "yuan":      yuan,
        "formatted": fmt.Sprintf("¥%.2f", yuan),
    }
}
```

### Result 系统查询

```go
// 启用访问器的查询
users, err := torm.Table("users").Model(&User{}).Get()    // *ResultCollection
user, err := torm.Table("users").Model(&User{}).First()   // *Result

// 高性能原始数据查询
rawUsers, err := torm.Table("users").GetRaw()    // []map[string]interface{}
rawUser, err := torm.Table("users").FirstRaw()   // map[string]interface{}
```

### 数据处理

```go
// 单条记录处理
user, _ := torm.Table("users").Model(&User{}).Where("id", "=", 1).First()

// 通过访问器获取格式化数据
fmt.Printf("状态: %v\n", user.Get("status"))      // {"code": 1, "name": "正常", "color": "green"}
fmt.Printf("薪资: %v\n", user.Get("salary"))      // {"cents": 800000, "yuan": 8000.0, "formatted": "¥8000.00"}

// 获取原始数据（用于计算）
rawStatus := user.GetRaw("status").(int)          // 1
rawSalary := user.GetRaw("salary").(int)          // 800000

// JSON 输出
accessorJSON, _ := user.ToJSON()    // 包含访问器处理的完整JSON
rawJSON, _ := user.ToRawJSON()      // 原始数据JSON
```

### 集合操作

```go
users, _ := torm.Table("users").Model(&User{}).Get()

// 遍历处理
users.Each(func(index int, user *db.Result) bool {
    fmt.Printf("用户 %d: %v\n", index+1, user.Get("username"))
    return true  // 继续遍历
})

// 函数式过滤
activeUsers := users.Filter(func(user *db.Result) bool {
    status := user.Get("status").(map[string]interface{})
    return status["code"].(int) == 1  // 只要正常状态用户
})

// 映射操作
usernames := users.Map(func(user *db.Result) interface{} {
    return user.Get("username")
})

// 集合JSON输出
fmt.Printf("活跃用户数: %d\n", activeUsers.Count())
json, _ := activeUsers.ToJSON()
fmt.Printf("JSON: %s\n", json)
```

### API 选择指南

```go
// 🎯 显示层：使用 Model().Get()
func getUsersForDisplay() {
    users, _ := torm.Table("users").
        Model(&User{}).                    // 启用访问器
        Where("status", "=", 1).
        Get()
    
    // 自动格式化的数据，适合前端展示
    json, _ := users.ToJSON()
    return json
}

// ⚡ 计算层：使用 GetRaw()
func calculateStats() {
    users, _ := torm.Table("users").
        Where("status", "=", 1).
        GetRaw()                          // 高性能原始数据
    
    // 直接操作原始数据，性能最优
    var totalSalary int64
    for _, user := range users {
        totalSalary += user["salary"].(int64)
    }
    return totalSalary
}

// 🔄 混合使用
func processUsers() {
    users, _ := torm.Table("users").Model(&User{}).Get()
    
    users.Each(func(index int, user *db.Result) bool {
        // 显示数据
        statusInfo := user.Get("status")
        fmt.Printf("用户状态: %v\n", statusInfo)
        
        // 业务逻辑使用原始值
        rawStatus := user.GetRaw("status").(int)
        if rawStatus == 1 {
            // 执行业务逻辑
        }
        return true
    })
}
```