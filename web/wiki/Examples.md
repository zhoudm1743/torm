# 示例代码

本文档提供了TORM现代化ORM的完整使用示例，涵盖了从基础操作到高级功能的各种场景。

## 🚀 v1.1.6 最新特性

### 🔍 增强WHERE查询方法

TORM v1.1.6 新增了完整的WHERE查询方法，完美对标ThinkORM：

```go
package main

import (
    "log"
    "github.com/zhoudm1743/torm/db"
)

func main() {
    // 配置数据库
    conf := &db.Config{
        Driver:   "sqlite",
        Database: ":memory:",
    }
    db.AddConnection("default", conf)

    // ===== 增强WHERE查询演示 =====
    
    query, _ := db.Table("users", "default")
    
    // NULL值查询
    activeUsersWithEmail, _ := query.
        WhereNotNull("email").
        WhereNull("deleted_at").
        Where("status", "=", "active").
        Get()
    
    // 范围查询
    adultUsers, _ := query.
        WhereBetween("age", []interface{}{18, 65}).
        WhereNotBetween("score", []interface{}{0, 60}).
        Get()
    
    // 子查询存在性检查
    usersWithOrders, _ := query.
        WhereExists("SELECT 1 FROM orders WHERE orders.user_id = users.id").
        WhereNotExists("SELECT 1 FROM banned_users WHERE banned_users.user_id = users.id").
        Get()
    
    // 高级排序功能
    randomUsers, _ := query.OrderRand().Limit(10).Get()
    
    // 按状态优先级排序
    priorityUsers, _ := query.
        OrderField("status", []interface{}{"premium", "active", "trial"}, "asc").
        Get()
    
    // 添加聚合字段
    userStats, _ := query.
        FieldRaw("COUNT(*) as total_count").
        FieldRaw("AVG(age) as avg_age").
        GroupBy("city").
        Get()
    
    log.Printf("增强查询功能演示完成")
}
```

## 🎯 v1.1.0 核心特性

### 🔍 查询构建器增强功能

TORM 提供了强大的查询构建器，支持链式调用和复杂查询：

```go
package main

import (
    "log"
    "github.com/zhoudm1743/torm/db"
)

func main() {
    // 配置数据库
    conf := &db.Config{
        Driver:   "mysql",
        Host:     "localhost",
        Port:     3306,
        Username: "root",
        Password: "123456",
        Database: "orm",
    }
    db.AddConnection("default", conf)

    // ===== 基础查询演示 =====
    
    // 创建查询构建器
    query, err := db.Table("users", "default")
    if err == nil {
        // 简单查询
        users, err := query.Select("id", "name", "email", "age").
            Where("status", "=", "active").
            OrderBy("created_at", "desc").
            Limit(5).
            Get()
        if err == nil {
            log.Printf("查询到 %d 个活跃用户", len(users))
        }

        // 条件统计
        count, err := query.Where("age", ">=", 18).
            Where("status", "=", "active").
            Count()
        if err == nil {
            log.Printf("成年活跃用户数量: %d", count)
        }
    }

    // ===== 高级查询演示 =====
    
    // 复杂条件查询
    complexQuery, err := db.Table("users", "default")
    if err == nil {
        result, err := complexQuery.
            Select("id", "name", "email").
            Where("age", "BETWEEN", []interface{}{20, 40}).
            WhereIn("status", []interface{}{"active", "pending"}).
            OrderBy("age", "ASC").
            OrderBy("name", "DESC").
            Limit(10).
            Get()
        if err == nil {
            log.Printf("复杂查询结果数量: %d", len(result))
        }
    }

    // 聚合查询
    aggregateQuery, err := db.Table("users", "default")
    if err == nil {
        // 统计数量
        totalCount, _ := aggregateQuery.Count()
        log.Printf("用户总数: %d", totalCount)
        
        // 求和
        totalAge, _ := aggregateQuery.Sum("age")
        log.Printf("年龄总和: %v", totalAge)
        
        // 平均值
        avgAge, _ := aggregateQuery.Avg("age")
        log.Printf("平均年龄: %v", avgAge)
    }

    // CRUD 操作
    crudQuery, err := db.Table("users", "default")
    if err == nil {
        // 插入
        userID, err := crudQuery.Insert(map[string]interface{}{
            "name":   "新用户",
            "email":  "newuser@example.com",
            "age":    25,
            "status": "active",
        })
        if err == nil {
            log.Printf("新用户ID: %v", userID)
        }

        // 更新
        affected, err := crudQuery.Where("id", "=", userID).
            Update(map[string]interface{}{
                "age": 26,
            })
        if err == nil {
            log.Printf("更新了 %d 条记录", affected)
        }

        // 删除
        deleted, err := crudQuery.Where("id", "=", userID).Delete()
        if err == nil {
            log.Printf("删除了 %d 条记录", deleted)
        }
    }
}

### 🔍 First/Find 增强功能

新的 First 和 Find 方法支持同时填充当前模型和传入的指针，并返回原始 map 数据：

```go
package main

import (
    "log"
    "github.com/zhoudm1743/torm/db"
    "github.com/zhoudm1743/torm/examples/models"
)

func main() {
    // 配置数据库
    conf := &db.Config{
        Driver:   "mysql",
        Host:     "localhost",
        Port:     3306,
        Username: "root",
        Password: "123456",
        Database: "orm",
    }
    db.AddConnection("default", conf)

    // First方法 - 只填充当前模型
    user1 := models.NewUser()
    result1, err := user1.Where("id", "=", 1).First()
    if err != nil {
        log.Printf("查询失败: %v", err)
    } else {
        log.Printf("当前模型: Name=%s, Age=%d", user1.Name, user1.Age)
        log.Printf("返回数据: %+v", result1)
    }

    // First方法 - 同时填充传入的指针
    user2 := models.NewUser()
    var anotherUser models.User
    result2, err := user2.Where("id", "=", 2).First(&anotherUser)
    if err != nil {
        log.Printf("查询失败: %v", err)
    } else {
        log.Printf("当前模型: %s", user2.Name)
        log.Printf("传入指针: %s", anotherUser.Name)
        log.Printf("返回数据: %+v", result2)
    }

    // Find方法 - 同时填充传入的指针  
    user3 := models.NewUser()
    var targetUser models.User
    result3, err := user3.Find(1, &targetUser)
    if err != nil {
        log.Printf("查询失败: %v", err)
    } else {
        log.Printf("当前模型: %s", user3.Name)
        log.Printf("传入指针: %s", targetUser.Name)
        log.Printf("返回数据: %+v", result3)
    }
}
```

### 🔑 自定义主键和复合主键

TORM 现在支持灵活的主键配置，包括 UUID、复合主键等：

```go
package main

import (
    "time"
    "github.com/zhoudm1743/torm/model"
)

// 默认主键模型
type User struct {
    model.BaseModel
    ID        interface{} `json:"id" db:"id"`
    Name      string      `json:"name" db:"name"`
    Email     string      `json:"email" db:"email"`
    CreatedAt time.Time   `json:"created_at" db:"created_at"`
}

// UUID主键模型
type Product struct {
    model.BaseModel
    UUID        string  `json:"uuid" db:"uuid" primary:"true"`    // UUID主键
    Name        string  `json:"name" db:"name"`
    Price       float64 `json:"price" db:"price"`
    CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// 复合主键模型（多租户场景）
type UserRole struct {
    model.BaseModel
    TenantID string `json:"tenant_id" db:"tenant_id" primary:"true"`  // 复合主键1
    UserID   string `json:"user_id" db:"user_id" primary:"true"`      // 复合主键2
    Role     string `json:"role" db:"role"`
    CreatedAt time.Time `json:"created_at" db:"created_at"`
}

func NewUser() *User {
    user := &User{BaseModel: *model.NewBaseModel()}
    user.SetTable("users")
    return user
}

func NewProduct() *Product {
    product := &Product{BaseModel: *model.NewBaseModel()}
    product.SetTable("products")
    // 自动检测主键标签
    product.DetectPrimaryKeysFromStruct(product)
    return product
}

func NewUserRole() *UserRole {
    userRole := &UserRole{BaseModel: *model.NewBaseModel()}
    userRole.SetTable("user_roles")
    // 自动检测复合主键标签
    userRole.DetectPrimaryKeysFromStruct(userRole)
    return userRole
}

func demonstratePrimaryKeys() {
    // 默认主键
    user := NewUser()
    log.Printf("默认主键: %v", user.PrimaryKeys())

    // UUID主键
    product := NewProduct()
    product.UUID = "550e8400-e29b-41d4-a716-446655440000"
    product.SetAttribute("uuid", product.UUID)
    log.Printf("UUID主键: %v, 值: %v", product.PrimaryKeys(), product.GetKey())

    // 复合主键
    userRole := NewUserRole()
    userRole.SetAttribute("tenant_id", "tenant-001")
    userRole.SetAttribute("user_id", "user-001")
    log.Printf("复合主键: %v, 值: %v", userRole.PrimaryKeys(), userRole.GetKey())

    // 手动设置主键
    customUser := NewUser()
    customUser.SetPrimaryKeys([]string{"tenant_id", "user_code"})
    log.Printf("手动设置复合主键: %v", customUser.PrimaryKeys())
}
```

### 📊 db包增强功能

底层db包的 First 和 Find 方法也支持了指针填充：

```go
package main

import (
    "log"
    "github.com/zhoudm1743/torm/db"
    "github.com/zhoudm1743/torm/examples/models"
)

func demonstrateDBPackage() {
    // db.Table().First() - 只返回map
    query1, err := db.Table("users", "default")
    if err == nil {
        result1, err := query1.Where("id", "=", 1).First()
        if err == nil {
            log.Printf("db.First() 结果: %s", result1["name"])
        }
    }

    // db.Table().First(&model) - 填充指针 + 返回map
    query2, err := db.Table("users", "default")
    if err == nil {
        var user models.User
        result2, err := query2.Where("id", "=", 1).First(&user)
        if err == nil {
            log.Printf("填充的模型: Name=%s", user.Name)
            log.Printf("返回的map: %+v", result2)
        }
    }

    // db.Table().Find(&model) - 同样支持指针填充
    query3, err := db.Table("users", "default")
    if err == nil {
        var user models.User
        result3, err := query3.Find(1, &user)
        if err == nil {
            log.Printf("Find填充的模型: Name=%s", user.Name)
            log.Printf("Find返回的map: %+v", result3)
        }
    }
}
```

## 🌟 v1.1.0 其他新功能示例

### 🔗 关联预加载 (Eager Loading)

解决 N+1 查询问题，大幅提升性能：

```go
package main

import (
    "context"
    "log"
    "github.com/zhoudm1743/torm/db"
    "github.com/zhoudm1743/torm/model"
)

// 用户模型
type User struct {
    *model.BaseModel
}

func NewUser() *User {
    user := &User{BaseModel: model.NewBaseModel()}
    user.SetTable("users")
    return user
}

// 定义关联关系
func (u *User) Profile() *model.HasOne {
    return u.HasOne(&Profile{}, "user_id", "id")
}

func (u *User) Posts() *model.HasMany {
    return u.HasMany(&Post{}, "user_id", "id")
}

func main() {
    // 初始化数据库连接
    config := &db.Config{
        Driver: "mysql",
        Host: "localhost", 
        Port: 3306,
        Database: "blog",
        Username: "root",
        Password: "password",
    }
    
    err := db.AddConnection("default", config)
    if err != nil {
        log.Fatal(err)
    }
    
    ctx := context.Background()
    
    // 获取用户数据
    query, _ := db.Table("users")
    userData, _ := query.Limit(10).Get()
    
    // 转换为模型
    users := make([]interface{}, len(userData))
    for i, data := range userData {
        user := NewUser()
        user.Fill(data)
        users[i] = user
    }
    
    // 创建模型集合并预载入关联
    collection := model.NewModelCollection(users)
    
    // 预载入用户资料和文章
    collection.With("profile", "posts")
    
    // 为文章关联添加条件
    collection.WithClosure("posts", func(q db.QueryInterface) db.QueryInterface {
        return q.Where("status", "=", "published").
            OrderBy("created_at", "desc").
            Limit(5)
    })
    
    // 执行预载入 - 只会执行3个查询而不是N+1个
    err = collection.Load(ctx)
    if err != nil {
        log.Fatal(err)
    }
    
    // 使用预载入的数据
    for _, userInterface := range collection.Models() {
        if user, ok := userInterface.(*User); ok {
            // 直接使用预载入的关联数据，无需额外查询
            if user.HasRelation("profile") {
                profile := user.GetRelation("profile")
                log.Printf("用户资料: %+v", profile)
            }
            
            if user.HasRelation("posts") {
                posts := user.GetRelation("posts")
                log.Printf("用户文章: %+v", posts)
            }
        }
    }
}
```

### 📄 分页功能

支持传统分页和高性能游标分页：

```go
package main

import (
    "context"
    "log"
    "github.com/zhoudm1743/torm/db"
    "github.com/zhoudm1743/torm/paginator"
)

func main() {
    // 初始化数据库连接
    config := &db.Config{
        Driver: "mysql",
        Host: "localhost",
        Port: 3306,
        Database: "blog", 
        Username: "root",
        Password: "password",
    }
    
    err := db.AddConnection("default", config)
    if err != nil {
        log.Fatal(err)
    }
    
    ctx := context.Background()
    
    // 1. 传统分页
    query, _ := db.Table("users")
    query = query.Where("status", "=", "active").OrderBy("created_at", "desc")
    
    // 使用内置分页方法
    result, err := query.Paginate(1, 10) // 第1页，每页10条
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("分页结果: %+v", result)
    
    // 2. 高级分页器
    queryPaginator := paginator.NewQueryPaginator(query, ctx)
    paginationResult, err := queryPaginator.
        SetPerPage(15).
        SetPage(2).
        Paginate()
    
    if err != nil {
        log.Fatal(err)
    }
    
    if pg, ok := paginationResult.(paginator.PaginatorInterface); ok {
        log.Printf("总记录数: %d", pg.Total())
        log.Printf("当前页: %d", pg.CurrentPage())
        log.Printf("总页数: %d", pg.LastPage())
        log.Printf("是否有下一页: %t", pg.HasMore())
        
        // 获取分页数据
        items := pg.Items()
        log.Printf("当前页数据: %+v", items)
        
        // 获取完整分页信息
        paginationData := pg.ToMap()
        log.Printf("分页信息: %+v", paginationData)
    }
    
    // 3. 游标分页 (适用于大数据量)
    items := []interface{}{
        map[string]interface{}{"id": 1, "name": "用户1"},
        map[string]interface{}{"id": 2, "name": "用户2"}, 
        map[string]interface{}{"id": 3, "name": "用户3"},
    }
    
    cursorPaginator := paginator.NewCursorPaginator(
        items, 
        10, 
        "eyJpZCI6MTB9", // next_cursor
        "eyJpZCI6MX0=", // prev_cursor
    )
    
    log.Printf("游标分页结果: %+v", cursorPaginator.ToMap())
}
```

### 🔍 JSON字段查询

跨数据库的JSON查询支持：

```go
package main

import (
    "log"
    "github.com/zhoudm1743/torm/db"
    "github.com/zhoudm1743/torm/query"
)

func main() {
    // 初始化数据库连接
    config := &db.Config{
        Driver: "mysql", // 支持 mysql, postgresql, sqlite
        Host: "localhost",
        Port: 3306,
        Database: "blog",
        Username: "root",
        Password: "password",
    }
    
    err := db.AddConnection("default", config)
    if err != nil {
        log.Fatal(err)
    }
    
    // 创建高级查询构建器
    baseQuery, _ := db.Table("users")
    advQuery := query.NewAdvancedQueryBuilder(baseQuery)
    
    // 1. JSON字段值查询
    result1 := advQuery.WhereJSON("profile", "$.age", ">", 25)
    
    // 2. JSON包含查询
    result2 := advQuery.WhereJSONContains("skills", "$.languages", "Go")
    
    // 3. JSON数组长度查询
    result3 := advQuery.WhereJSONLength("certifications", "$", ">=", 2)
    
    // 4. 复合JSON查询
    complexResult := advQuery.
        WhereJSON("metadata", "$.city", "=", "北京").
        WhereJSONContains("hobbies", "$.type", "技术").
        WhereJSONLength("projects", "$", ">", 5).
        OrderBy("created_at", "desc").
        Limit(20)
    
    data, err := complexResult.Get()
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("JSON查询结果: %+v", data)
    
    // 5. 跨数据库兼容的JSON查询
    // MySQL: 使用 JSON_EXTRACT、JSON_CONTAINS
    // PostgreSQL: 使用 jsonb 操作符 @>、->、->>
    // SQLite: 自动降级为 LIKE 查询
    universalQuery := advQuery.
        WhereJSON("settings", "$.theme", "=", "dark").
        WhereJSONContains("preferences", "$.notifications", true)
    
    universalData, err := universalQuery.Get()
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("跨数据库JSON查询结果: %+v", universalData)
}
```

### 🏗️ 高级查询功能

子查询和窗口函数支持：

```go
package main

import (
    "log"
    "github.com/zhoudm1743/torm/db"
    "github.com/zhoudm1743/torm/query"
)

func main() {
    // 初始化数据库连接
    config := &db.Config{
        Driver: "mysql",
        Host: "localhost",
        Port: 3306,
        Database: "company",
        Username: "root", 
        Password: "password",
    }
    
    err := db.AddConnection("default", config)
    if err != nil {
        log.Fatal(err)
    }
    
    // 创建高级查询构建器
    baseQuery, _ := db.Table("employees")
    advQuery := query.NewAdvancedQueryBuilder(baseQuery)
    
    // 1. EXISTS 子查询 - 查找有项目的员工
    employeesWithProjects := advQuery.WhereExists(func(q db.QueryInterface) db.QueryInterface {
        return q.Where("projects.employee_id", "=", "employees.id").
            Where("projects.status", "=", "active")
    })
    
    // 2. NOT EXISTS 子查询 - 查找没有迟到记录的员工
    punctualEmployees := advQuery.WhereNotExists(func(q db.QueryInterface) db.QueryInterface {
        return q.Where("attendances.employee_id", "=", "employees.id").
            Where("attendances.status", "=", "late")
    })
    
    // 3. IN 子查询 - 查找高绩效部门的员工
    highPerformers := advQuery.WhereInSubQuery("department_id", func(q db.QueryInterface) db.QueryInterface {
        return q.Where("performance_score", ">", 90).
            Where("year", "=", 2024)
    })
    
    // 4. 窗口函数 - 部门内薪资排名
    salaryRanking := advQuery.
        WithRowNumber("row_num", "department_id", "salary DESC").
        WithRank("salary_rank", "department_id", "salary DESC").
        WithDenseRank("dense_rank", "department_id", "salary DESC")
    
    // 5. 窗口聚合 - 部门统计
    departmentStats := advQuery.
        WithCountWindow("dept_employee_count", "department_id").
        WithSumWindow("salary", "dept_total_salary", "department_id").
        WithAvgWindow("salary", "dept_avg_salary", "department_id")
    
    // 6. LAG/LEAD 函数 - 获取前一个员工的薪资
    salaryComparison := advQuery.
        WithLag("salary", "prev_employee_salary", "department_id", "hire_date", 1, 0)
    
    // 7. 复合高级查询
    complexAnalysis := advQuery.
        Where("status", "=", "active").
        WhereJSON("skills", "$.level", ">=", "senior").
        WhereExists(func(q db.QueryInterface) db.QueryInterface {
            return q.Where("performance.employee_id", "=", "employees.id").
                Where("performance.score", ">", 85)
        }).
        WithRowNumber("performance_rank", "department_id", "hire_date").
        WithAvgWindow("salary", "dept_avg", "department_id").
        OrderBy("department_id", "asc").
        OrderBy("salary", "desc").
        Limit(50)
    
    // 执行查询
    data, err := complexAnalysis.Get()
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("复合分析结果: %+v", data)
    
    // 8. 分页 + 高级查询
    paginatedResult, err := complexAnalysis.Paginate(1, 20)
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("分页的高级查询结果: %+v", paginatedResult)
}
```

---

## 🚀 基础示例

### 连接数据库（现代化方式）

```go
package main

import (
    "log"
    "time"
    
    "github.com/zhoudm1743/torm/db"
)

func main() {
    config := &db.Config{
        Driver:          "mysql",
        Host:            "localhost",
        Port:            3306,
        Database:        "blog",
        Username:        "root",
        Password:        "password",
        Charset:         "utf8mb4",
        MaxOpenConns:    100,
        MaxIdleConns:    10,
        ConnMaxLifetime: time.Hour,
        LogQueries:      true,
    }
    
    err := db.AddConnection("default", config)
    if err != nil {
        log.Fatal(err)
    }
    
    conn, err := db.DB("default")
    if err != nil {
        log.Fatal(err)
    }
    
    // ✅ 现代化API - 无需context参数
    err = conn.Connect()
    if err != nil {
        log.Fatal(err)
    }
    
    // 可选的超时控制
    // err = conn.Ping() // 默认无超时
    
    log.Println("✅ 数据库连接成功！")
}
```

### 查询构建器基础用法

```go
package main

import (
    "fmt"
    "log"
    "time"
    
    "github.com/zhoudm1743/torm/db"
)

func main() {
    // 配置数据库...（省略）
    
    // ✅ 获取查询构建器 - 简洁的API
    query, err := db.Table("users")
    if err != nil {
        log.Fatal(err)
    }
    
    // 1. 插入数据
    userID, err := query.Insert(map[string]interface{}{
        "name":     "张三",
        "email":    "zhangsan@example.com",
        "age":      28,
        "status":   "active",
        "created_at": time.Now(),
    })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("✅ 插入成功，用户ID: %v\n", userID)
    
    // 2. 查询数据
    users, err := query.
        Where("status", "=", "active").
        Where("age", ">=", 18).
        OrderBy("created_at", "desc").
        Limit(10).
        Get()
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("✅ 找到 %d 个用户\n", len(users))
    
    // 3. 更新数据
    affected, err := query.
        Where("id", "=", userID).
        Update(map[string]interface{}{
            "age": 29,
            "updated_at": time.Now(),
        })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("✅ 更新了 %d 条记录\n", affected)
    
    // 4. 删除数据
    deleted, err := query.
        Where("id", "=", userID).
        Delete()
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("✅ 删除了 %d 条记录\n", deleted)
}
```

### 高级查询示例

```go
package main

import (
    "fmt"
    "log"
    "time"
    
    "github.com/zhoudm1743/torm/db"
)

func advancedQueryExamples() {
    // 1. 复杂条件查询
    query, _ := db.Table("users")
    
    results, err := query.
        Select("users.name", "profiles.avatar", "COUNT(posts.id) as post_count").
        LeftJoin("profiles", "users.id", "=", "profiles.user_id").
        LeftJoin("posts", "users.id", "=", "posts.user_id").
        Where("users.status", "=", "active").
        WhereIn("users.role", []interface{}{"admin", "editor"}).
        WhereBetween("users.age", 25, 65).
        WhereNotNull("profiles.avatar").
        GroupBy("users.id").
        Having("post_count", ">", 0).
        OrderBy("post_count", "desc").
        Limit(20).
        Get()
    
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("✅ 复杂查询返回 %d 条记录\n", len(results))
    
    // 2. 聚合查询
    stats, err := query.
        Select(
            "department", 
            "COUNT(*) as user_count", 
            "AVG(age) as avg_age",
            "MAX(salary) as max_salary",
            "MIN(created_at) as earliest_join",
        ).
        Where("status", "=", "active").
        GroupBy("department").
        Having("user_count", ">=", 5).
        OrderBy("avg_age", "desc").
        Get()
    
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("✅ 统计查询返回 %d 个部门\n", len(stats))
    
    // 3. 子查询示例
    subQuery, _ := db.Table("posts")
    subQuerySQL, bindings, _ := subQuery.
        Select("user_id").
        Where("status", "=", "published").
        GroupBy("user_id").
        Having("COUNT(*)", ">", 10).
        ToSQL()
    
    activeWriters, err := query.
        Where("status", "=", "active").
        WhereRaw("id IN ("+subQuerySQL+")", bindings...).
        Get()
    
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("✅ 找到 %d 个活跃作者\n", len(activeWriters))
}
```

### 批量操作示例

```go
package main

import (
    "fmt"
    "log"
    "time"
    
    "github.com/zhoudm1743/torm/db"
)

func batchOperationsExample() {
    query, _ := db.Table("users")
    
    // 1. 批量插入
    users := []map[string]interface{}{
        {
            "name":       "用户1",
            "email":      "user1@example.com",
            "age":        25,
            "status":     "active",
            "created_at": time.Now(),
        },
        {
            "name":       "用户2", 
            "email":      "user2@example.com",
            "age":        30,
            "status":     "active",
            "created_at": time.Now(),
        },
        {
            "name":       "用户3",
            "email":      "user3@example.com", 
            "age":        35,
            "status":     "pending",
            "created_at": time.Now(),
        },
    }
    
    affected, err := query.InsertBatch(users)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("✅ 批量插入 %d 条记录\n", affected)
    
    // 2. 批量更新
    affected, err = query.
        Where("status", "=", "pending").
        Update(map[string]interface{}{
            "status":     "active",
            "updated_at": time.Now(),
        })
    
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("✅ 批量更新 %d 条记录\n", affected)
    
    // 3. 条件删除
    affected, err = query.
        Where("status", "=", "inactive").
        Where("last_login", "<", time.Now().AddDate(0, -6, 0)). // 6个月未登录
        Delete()
    
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("✅ 清理了 %d 个非活跃用户\n", affected)
}
```

### 事务处理示例

```go
package main

import (
    "fmt"
    "log"
    "time"
    
    "github.com/zhoudm1743/torm/db"
)

func transactionExample() {
    // ✅ 现代化事务API - 无需context
    err := db.Transaction(func(tx db.TransactionInterface) error {
        // 1. 创建用户
        userResult, err := tx.Exec(`
            INSERT INTO users (name, email, age, status, created_at) 
            VALUES (?, ?, ?, ?, ?)
        `, "事务用户", "transaction@example.com", 28, "active", time.Now())
        
        if err != nil {
            return err // 自动回滚
        }
        
        userID, err := userResult.LastInsertId()
        if err != nil {
            return err
        }
        
        // 2. 创建用户资料
        _, err = tx.Exec(`
            INSERT INTO profiles (user_id, avatar, bio, created_at) 
            VALUES (?, ?, ?, ?)
        `, userID, "default-avatar.png", "新用户", time.Now())
        
        if err != nil {
            return err // 自动回滚
        }
        
        // 3. 记录操作日志
        _, err = tx.Exec(`
            INSERT INTO user_logs (user_id, action, details, created_at) 
            VALUES (?, ?, ?, ?)
        `, userID, "user_created", "用户注册", time.Now())
        
        if err != nil {
            return err // 自动回滚
        }
        
        fmt.Printf("✅ 事务中创建用户，ID: %d\n", userID)
        return nil // 自动提交
    })
    
    if err != nil {
        log.Printf("❌ 事务失败: %v", err)
        return
    }
    
    fmt.Println("✅ 事务执行成功！")
}

// 复杂事务示例：银行转账
func bankTransferExample() {
    err := db.Transaction(func(tx db.TransactionInterface) error {
        // 1. 检查转出账户余额
        var fromBalance float64
        err := tx.QueryRow(`
            SELECT balance FROM accounts WHERE id = ? FOR UPDATE
        `, 1).Scan(&fromBalance)
        
        if err != nil {
            return fmt.Errorf("查询转出账户失败: %v", err)
        }
        
        transferAmount := 1000.0
        if fromBalance < transferAmount {
            return fmt.Errorf("余额不足，当前余额: %.2f", fromBalance)
        }
        
        // 2. 扣除转出账户余额
        _, err = tx.Exec(`
            UPDATE accounts SET balance = balance - ?, updated_at = ? 
            WHERE id = ?
        `, transferAmount, time.Now(), 1)
        
        if err != nil {
            return fmt.Errorf("扣款失败: %v", err)
        }
        
        // 3. 增加转入账户余额
        _, err = tx.Exec(`
            UPDATE accounts SET balance = balance + ?, updated_at = ? 
            WHERE id = ?
        `, transferAmount, time.Now(), 2)
        
        if err != nil {
            return fmt.Errorf("入账失败: %v", err)
        }
        
        // 4. 记录转账日志
        _, err = tx.Exec(`
            INSERT INTO transfer_logs (from_account, to_account, amount, status, created_at) 
            VALUES (?, ?, ?, ?, ?)
        `, 1, 2, transferAmount, "completed", time.Now())
        
        if err != nil {
            return fmt.Errorf("记录日志失败: %v", err)
        }
        
        fmt.Printf("✅ 转账成功: %.2f 元\n", transferAmount)
        return nil
    })
    
    if err != nil {
        log.Printf("❌ 转账失败: %v", err)
        return
    }
    
    fmt.Println("✅ 转账事务完成！")
}
```

### 超时控制示例

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"
    
    "github.com/zhoudm1743/torm/db"
)

func timeoutControlExample() {
    query, _ := db.Table("users")
    
    // 1. 使用WithTimeout进行超时控制
    users, err := query.
        WithTimeout(5 * time.Second).  // 5秒超时
        Where("status", "=", "active").
        OrderBy("created_at", "desc").
        Limit(100).
        Get()
    
    if err != nil {
        log.Printf("❌ 查询超时: %v", err)
        return
    }
    fmt.Printf("✅ 在5秒内查询到 %d 个用户\n", len(users))
    
    // 2. 使用WithContext进行更精细的控制
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    largeDataset, err := query.
        WithContext(ctx).
        Select("*").
        OrderBy("id", "asc").
        Get()
    
    if err != nil {
        log.Printf("❌ 大数据查询失败: %v", err)
        return
    }
    fmt.Printf("✅ 查询大数据集: %d 条记录\n", len(largeDataset))
    
    // 3. 长时间运行的操作超时控制
    longRunningCtx, longCancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer longCancel()
    
    err = db.Transaction(func(tx db.TransactionInterface) error {
        // 在事务内部也可以检查context状态
        select {
        case <-longRunningCtx.Done():
            return longRunningCtx.Err()
        default:
        }
        
        // 执行长时间操作...
        result, err := tx.Exec(`
            UPDATE users SET status = 'verified' 
            WHERE email_verified = 1 AND status = 'pending'
        `)
        if err != nil {
            return err
        }
        
        affected, _ := result.RowsAffected()
        fmt.Printf("✅ 批量验证了 %d 个用户\n", affected)
        
        return nil
    })
    
    if err != nil {
        log.Printf("❌ 长时间操作失败: %v", err)
        return
    }
    
    fmt.Println("✅ 长时间操作完成！")
}
```

## 📦 数据迁移示例

```go
package main

import (
    "context"
    "log"
    
    "github.com/zhoudm1743/torm/db"
    "github.com/zhoudm1743/torm/migration"
)

func main() {
    // ... 数据库配置 ...
    
    conn, _ := db.DB("default")
    migrator := migration.NewMigrator(conn, nil)
    
    // 注册迁移
    registerMigrations(migrator)
    
    ctx := context.Background()
    
    // 执行迁移
    if err := migrator.Up(ctx); err != nil {
        log.Fatal("迁移失败:", err)
    }
    
    // 显示状态
    migrator.PrintStatus(ctx)
}

func registerMigrations(migrator *migration.Migrator) {
    // 创建用户表
    migrator.RegisterFunc(
        "20240101_000001",
        "创建用户表",
        func(ctx context.Context, conn db.ConnectionInterface) error {
            schema := migration.NewSchemaBuilder(conn)
            
            table := &migration.Table{
                Name: "users",
                Columns: []*migration.Column{
                    {
                        Name:          "id",
                        Type:          migration.ColumnTypeBigInt,
                        PrimaryKey:    true,
                        AutoIncrement: true,
                        NotNull:       true,
                    },
                    {
                        Name:    "name",
                        Type:    migration.ColumnTypeVarchar,
                        Length:  100,
                        NotNull: true,
                    },
                    {
                        Name:    "email",
                        Type:    migration.ColumnTypeVarchar,
                        Length:  100,
                        NotNull: true,
                    },
                    {
                        Name:    "age",
                        Type:    migration.ColumnTypeInt,
                        NotNull: true,
                    },
                    {
                        Name:    "created_at",
                        Type:    migration.ColumnTypeDateTime,
                        Default: "CURRENT_TIMESTAMP",
                    },
                },
                Indexes: []*migration.Index{
                    {
                        Name:    "idx_users_email",
                        Columns: []string{"email"},
                        Unique:  true,
                    },
                },
            }
            
            return schema.CreateTable(ctx, table)
        },
        func(ctx context.Context, conn db.ConnectionInterface) error {
            schema := migration.NewSchemaBuilder(conn)
            return schema.DropTable(ctx, "users")
        },
    )
    
    // 创建文章表
    migrator.RegisterFunc(
        "20240101_000002",
        "创建文章表",
        func(ctx context.Context, conn db.ConnectionInterface) error {
            schema := migration.NewSchemaBuilder(conn)
            
            table := &migration.Table{
                Name: "posts",
                Columns: []*migration.Column{
                    {
                        Name:          "id",
                        Type:          migration.ColumnTypeBigInt,
                        PrimaryKey:    true,
                        AutoIncrement: true,
                        NotNull:       true,
                    },
                    {
                        Name:    "title",
                        Type:    migration.ColumnTypeVarchar,
                        Length:  200,
                        NotNull: true,
                    },
                    {
                        Name: "content",
                        Type: migration.ColumnTypeText,
                    },
                    {
                        Name:    "user_id",
                        Type:    migration.ColumnTypeBigInt,
                        NotNull: true,
                    },
                    {
                        Name:    "created_at",
                        Type:    migration.ColumnTypeDateTime,
                        Default: "CURRENT_TIMESTAMP",
                    },
                },
                ForeignKeys: []*migration.ForeignKey{
                    {
                        Name:              "fk_posts_user_id",
                        Columns:           []string{"user_id"},
                        ReferencedTable:   "users",
                        ReferencedColumns: []string{"id"},
                        OnDelete:          "CASCADE",
                    },
                },
            }
            
            return schema.CreateTable(ctx, table)
        },
        func(ctx context.Context, conn db.ConnectionInterface) error {
            schema := migration.NewSchemaBuilder(conn)
            return schema.DropTable(ctx, "posts")
        },
    )
}
```

## 🔄 事务处理示例

```go
package main

import (
    "context"
    "log"
    
    "github.com/zhoudm1743/torm/db"
)

func main() {
    // ... 数据库配置 ...
    
    ctx := context.Background()
    conn, _ := db.DB("default")
    
    // 简单事务
    simpleTransaction(ctx, conn)
    
    // 复杂事务
    complexTransaction(ctx, conn)
}

func simpleTransaction(ctx context.Context, conn db.ConnectionInterface) {
    tx, err := conn.Begin(ctx)
    if err != nil {
        log.Fatal("开始事务失败:", err)
    }
    defer tx.Rollback() // 确保回滚
    
    // 执行操作
    _, err = tx.Exec(ctx, "INSERT INTO users (name, email, age) VALUES (?, ?, ?)", "事务用户", "tx@example.com", 30)
    if err != nil {
        log.Fatal("插入失败:", err)
    }
    
    // 提交事务
    if err = tx.Commit(); err != nil {
        log.Fatal("提交事务失败:", err)
    }
    
    log.Println("简单事务完成")
}

func complexTransaction(ctx context.Context, conn db.ConnectionInterface) {
    tx, err := conn.Begin(ctx)
    if err != nil {
        log.Fatal("开始事务失败:", err)
    }
    defer func() {
        if r := recover(); r != nil {
            tx.Rollback()
            log.Printf("事务回滚: %v", r)
        }
    }()
    
    // 插入用户
    result, err := tx.Exec(ctx, "INSERT INTO users (name, email, age) VALUES (?, ?, ?)", "作者", "author@example.com", 35)
    if err != nil {
        tx.Rollback()
        log.Fatal("插入用户失败:", err)
    }
    
    userID, _ := result.LastInsertId()
    
    // 插入文章
    _, err = tx.Exec(ctx, "INSERT INTO posts (title, content, user_id) VALUES (?, ?, ?)", "事务文章", "这是在事务中创建的文章", userID)
    if err != nil {
        tx.Rollback()
        log.Fatal("插入文章失败:", err)
    }
    
    // 提交事务
    if err = tx.Commit(); err != nil {
        log.Fatal("提交事务失败:", err)
    }
    
    log.Println("复杂事务完成")
}
```

## 🗄️ 多数据库示例

```go
package main

import (
    "context"
    "log"
    "time"
    
    "go.mongodb.org/mongo-driver/bson"
    "github.com/zhoudm1743/torm/db"
)

func main() {
    // 配置多个数据库
    setupDatabases()
    
    ctx := context.Background()
    
    // MySQL 操作
    mysqlOperations(ctx)
    
    // MongoDB 操作
    mongodbOperations(ctx)
    
    // SQLite 操作
    sqliteOperations(ctx)
}

func setupDatabases() {
    // MySQL 连接
    mysqlConfig := &db.Config{
        Driver:   "mysql",
        Host:     "localhost",
        Port:     3306,
        Database: "mysql_db",
        Username: "root",
        Password: "password",
    }
    db.AddConnection("mysql", mysqlConfig)
    
    // MongoDB 连接
    mongoConfig := &db.Config{
        Driver:   "mongodb",
        Host:     "localhost",
        Port:     27017,
        Database: "mongo_db",
    }
    db.AddConnection("mongodb", mongoConfig)
    
    // SQLite 连接
    sqliteConfig := &db.Config{
        Driver:   "sqlite",
        Database: "sqlite_db.db",
    }
    db.AddConnection("sqlite", sqliteConfig)
}

func mysqlOperations(ctx context.Context) {
    conn, err := db.DB("mysql")
    if err != nil {
        log.Printf("MySQL连接失败: %v", err)
        return
    }
    
    // 创建表
    _, err = conn.Exec(ctx, `
        CREATE TABLE IF NOT EXISTS mysql_users (
            id BIGINT AUTO_INCREMENT PRIMARY KEY,
            name VARCHAR(100),
            email VARCHAR(100),
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        )
    `)
    if err != nil {
        log.Printf("MySQL创建表失败: %v", err)
        return
    }
    
    // 插入数据
    _, err = conn.Exec(ctx, "INSERT INTO mysql_users (name, email) VALUES (?, ?)", "MySQL用户", "mysql@example.com")
    if err != nil {
        log.Printf("MySQL插入失败: %v", err)
        return
    }
    
    log.Println("MySQL操作完成")
}

func mongodbOperations(ctx context.Context) {
    conn, err := db.DB("mongodb")
    if err != nil {
        log.Printf("MongoDB连接失败: %v", err)
        return
    }
    
    mongoConn := db.GetMongoConnection(conn)
    if mongoConn == nil {
        log.Println("MongoDB连接转换失败")
        return
    }
    
    collection := mongoConn.GetCollection("users")
    query := db.NewMongoQuery(collection, nil)
    
    // 插入文档
    user := bson.M{
        "name":       "MongoDB用户",
        "email":      "mongo@example.com",
        "created_at": time.Now(),
    }
    
    _, err = query.InsertOne(ctx, user)
    if err != nil {
        log.Printf("MongoDB插入失败: %v", err)
        return
    }
    
    // 查询文档
    count, err := query.Count(ctx)
    if err != nil {
        log.Printf("MongoDB查询失败: %v", err)
        return
    }
    
    log.Printf("MongoDB操作完成，文档数量: %d", count)
}

func sqliteOperations(ctx context.Context) {
    conn, err := db.DB("sqlite")
    if err != nil {
        log.Printf("SQLite连接失败: %v", err)
        return
    }
    
    // 创建表
    _, err = conn.Exec(ctx, `
        CREATE TABLE IF NOT EXISTS sqlite_users (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            name TEXT,
            email TEXT,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP
        )
    `)
    if err != nil {
        log.Printf("SQLite创建表失败: %v", err)
        return
    }
    
    // 插入数据
    _, err = conn.Exec(ctx, "INSERT INTO sqlite_users (name, email) VALUES (?, ?)", "SQLite用户", "sqlite@example.com")
    if err != nil {
        log.Printf("SQLite插入失败: %v", err)
        return
    }
    
    log.Println("SQLite操作完成")
}
```

## 🎯 实际应用示例

### 博客系统

```go
package main

import (
    "context"
    "log"
    "time"
    
    "github.com/zhoudm1743/torm/db"
)

type User struct {
    ID        int64     `db:"id" json:"id"`
    Username  string    `db:"username" json:"username"`
    Email     string    `db:"email" json:"email"`
    Password  string    `db:"password" json:"-"`
    CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type Post struct {
    ID        int64     `db:"id" json:"id"`
    Title     string    `db:"title" json:"title"`
    Content   string    `db:"content" json:"content"`
    UserID    int64     `db:"user_id" json:"user_id"`
    Status    string    `db:"status" json:"status"`
    CreatedAt time.Time `db:"created_at" json:"created_at"`
    UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

type BlogService struct {
    conn db.ConnectionInterface
}

func NewBlogService(conn db.ConnectionInterface) *BlogService {
    return &BlogService{conn: conn}
}

func (s *BlogService) CreateUser(ctx context.Context, username, email, password string) (*User, error) {
    sql := `INSERT INTO users (username, email, password) VALUES (?, ?, ?)`
    
    result, err := s.conn.Exec(ctx, sql, username, email, password)
    if err != nil {
        return nil, err
    }
    
    id, _ := result.LastInsertId()
    
    return &User{
        ID:        id,
        Username:  username,
        Email:     email,
        CreatedAt: time.Now(),
    }, nil
}

func (s *BlogService) GetUserByEmail(ctx context.Context, email string) (*User, error) {
    sql := `SELECT id, username, email, created_at FROM users WHERE email = ?`
    
    row := s.conn.QueryRow(ctx, sql, email)
    
    user := &User{}
    err := row.Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt)
    if err != nil {
        return nil, err
    }
    
    return user, nil
}

func (s *BlogService) CreatePost(ctx context.Context, title, content string, userID int64) (*Post, error) {
    sql := `INSERT INTO posts (title, content, user_id, status, updated_at) VALUES (?, ?, ?, 'draft', CURRENT_TIMESTAMP)`
    
    result, err := s.conn.Exec(ctx, sql, title, content, userID)
    if err != nil {
        return nil, err
    }
    
    id, _ := result.LastInsertId()
    
    return &Post{
        ID:        id,
        Title:     title,
        Content:   content,
        UserID:    userID,
        Status:    "draft",
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }, nil
}

func (s *BlogService) PublishPost(ctx context.Context, postID int64) error {
    sql := `UPDATE posts SET status = 'published', updated_at = CURRENT_TIMESTAMP WHERE id = ?`
    
    _, err := s.conn.Exec(ctx, sql, postID)
    return err
}

func (s *BlogService) GetPostsByUser(ctx context.Context, userID int64) ([]*Post, error) {
    sql := `SELECT id, title, content, user_id, status, created_at, updated_at FROM posts WHERE user_id = ? ORDER BY created_at DESC`
    
    rows, err := s.conn.Query(ctx, sql, userID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var posts []*Post
    for rows.Next() {
        post := &Post{}
        err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.UserID, &post.Status, &post.CreatedAt, &post.UpdatedAt)
        if err != nil {
            return nil, err
        }
        posts = append(posts, post)
    }
    
    return posts, nil
}

func main() {
    // ... 数据库配置 ...
    
    conn, _ := db.DB("default")
    blogService := NewBlogService(conn)
    
    ctx := context.Background()
    
    // 创建用户
    user, err := blogService.CreateUser(ctx, "johndoe", "john@example.com", "password123")
    if err != nil {
        log.Fatal("创建用户失败:", err)
    }
    log.Printf("创建用户: %+v", user)
    
    // 创建文章
    post, err := blogService.CreatePost(ctx, "我的第一篇文章", "这是文章内容...", user.ID)
    if err != nil {
        log.Fatal("创建文章失败:", err)
    }
    log.Printf("创建文章: %+v", post)
    
    // 发布文章
    err = blogService.PublishPost(ctx, post.ID)
    if err != nil {
        log.Fatal("发布文章失败:", err)
    }
    
    // 获取用户的文章
    posts, err := blogService.GetPostsByUser(ctx, user.ID)
    if err != nil {
        log.Fatal("获取文章失败:", err)
    }
    log.Printf("用户文章数量: %d", len(posts))
}
```

---

**📚 更多示例请参考 [GitHub仓库](https://github.com/zhoudm1743/torm) 中的 examples 目录。** 