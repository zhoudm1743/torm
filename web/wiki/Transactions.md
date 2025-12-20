# 事务处理

TORM 提供了完整的事务支持，包括自动事务管理、手动事务控制、嵌套事务等功能，确保数据的一致性和完整性。

## 📋 目录

- [快速开始](#快速开始)
- [闭包事务](#闭包事务)
- [手动事务](#手动事务)
- [查询构建器事务](#查询构建器事务)
- [事务回滚](#事务回滚)
- [最佳实践](#最佳实践)

## 🚀 快速开始

### 基础事务使用

```go
import "github.com/zhoudm1743/torm/db"

// 使用闭包事务（推荐）
err := db.Transaction(func(tx db.TransactionInterface) error {
    // 在事务中执行操作
    _, err := tx.Exec("INSERT INTO users (name, email) VALUES (?, ?)", 
        "张三", "zhangsan@example.com")
    if err != nil {
        return err // 返回错误自动回滚
    }
    
    _, err = tx.Exec("UPDATE accounts SET balance = balance - 100 WHERE user_id = ?", 1)
    if err != nil {
        return err // 返回错误自动回滚
    }
    
    return nil // 返回 nil 自动提交
})

if err != nil {
    log.Printf("事务失败: %v", err)
}
```

## 🔄 闭包事务

闭包事务是最推荐的使用方式，自动处理提交和回滚。

### 基础用法

```go
import "github.com/zhoudm1743/torm/db"

// 默认连接
err := db.Transaction(func(tx db.TransactionInterface) error {
    // 事务操作
    return nil
})

// 指定连接
err := db.Transaction(func(tx db.TransactionInterface) error {
    // 事务操作
    return nil
}, "mysql_connection")
```

### 转账示例

```go
func transfer(fromUserID, toUserID int64, amount float64) error {
    return db.Transaction(func(tx db.TransactionInterface) error {
        // 1. 检查发送方余额
        var balance float64
        row := tx.QueryRow("SELECT balance FROM accounts WHERE user_id = ?", fromUserID)
        if err := row.Scan(&balance); err != nil {
            return fmt.Errorf("查询余额失败: %w", err)
        }
        
        if balance < amount {
            return fmt.Errorf("余额不足")
        }
        
        // 2. 扣除发送方余额
        _, err := tx.Exec("UPDATE accounts SET balance = balance - ? WHERE user_id = ?", 
            amount, fromUserID)
        if err != nil {
            return fmt.Errorf("扣款失败: %w", err)
        }
        
        // 3. 增加接收方余额
        _, err = tx.Exec("UPDATE accounts SET balance = balance + ? WHERE user_id = ?", 
            amount, toUserID)
        if err != nil {
            return fmt.Errorf("入账失败: %w", err)
        }
        
        // 4. 记录转账日志
        _, err = tx.Exec(`INSERT INTO transfer_logs 
            (from_user_id, to_user_id, amount, created_at) 
            VALUES (?, ?, ?, NOW())`, 
            fromUserID, toUserID, amount)
        if err != nil {
            return fmt.Errorf("记录日志失败: %w", err)
        }
        
        return nil // 成功，自动提交
    })
}
```

### 批量操作

```go
func batchCreateUsers(users []map[string]interface{}) error {
    return db.Transaction(func(tx db.TransactionInterface) error {
        for i, user := range users {
            _, err := tx.Exec(`INSERT INTO users (username, email, status) 
                VALUES (?, ?, ?)`,
                user["username"], user["email"], "active")
            if err != nil {
                return fmt.Errorf("创建用户 %d 失败: %w", i, err)
            }
        }
        return nil
    })
}
```

## ✋ 手动事务

当需要更精细的控制时，可以使用手动事务。

### 基础用法

```go
import "github.com/zhoudm1743/torm/db"

// 开始事务
tx, err := db.BeginTransaction()
if err != nil {
    return fmt.Errorf("开始事务失败: %w", err)
}

// 指定连接
tx, err := db.BeginTransaction("mysql_connection")
```

### 完整示例

```go
func createUserWithProfile(userData, profileData map[string]interface{}) error {
    // 开始事务
    tx, err := db.BeginTransaction()
    if err != nil {
        return fmt.Errorf("开始事务失败: %w", err)
    }
    
    // 确保事务被正确处理
    defer func() {
        if r := recover(); r != nil {
            tx.Rollback()
            panic(r) // 重新抛出 panic
        }
    }()
    
    // 创建用户
    result, err := tx.Exec(`INSERT INTO users (username, email) VALUES (?, ?)`,
        userData["username"], userData["email"])
    if err != nil {
        tx.Rollback()
        return fmt.Errorf("创建用户失败: %w", err)
    }
    
    userID, _ := result.LastInsertId()
    
    // 创建用户档案
    _, err = tx.Exec(`INSERT INTO profiles (user_id, avatar, bio) VALUES (?, ?, ?)`,
        userID, profileData["avatar"], profileData["bio"])
    if err != nil {
        tx.Rollback()
        return fmt.Errorf("创建档案失败: %w", err)
    }
    
    // 提交事务
    if err := tx.Commit(); err != nil {
        return fmt.Errorf("提交事务失败: %w", err)
    }
    
    return nil
}
```

### 使用 defer 简化

```go
func createOrder(orderData map[string]interface{}, items []map[string]interface{}) error {
    tx, err := db.BeginTransaction()
    if err != nil {
        return err
    }
    
    // 使用 defer 确保事务被处理
    var txErr error
    defer func() {
        if txErr != nil {
            tx.Rollback()
        }
    }()
    
    // 创建订单
    result, txErr := tx.Exec(`INSERT INTO orders (user_id, total, status) VALUES (?, ?, ?)`,
        orderData["user_id"], orderData["total"], "pending")
    if txErr != nil {
        return txErr
    }
    
    orderID, _ := result.LastInsertId()
    
    // 创建订单项
    for _, item := range items {
        _, txErr = tx.Exec(`INSERT INTO order_items (order_id, product_id, quantity, price) 
            VALUES (?, ?, ?, ?)`,
            orderID, item["product_id"], item["quantity"], item["price"])
        if txErr != nil {
            return txErr
        }
    }
    
    // 提交
    return tx.Commit()
}
```

## 🔧 查询构建器事务

在事务中使用查询构建器。

### 基础用法

```go
err := db.Transaction(func(tx db.TransactionInterface) error {
    // 创建事务查询构建器
    userQuery := tx.Table("users")
    
    // 插入
    userID, err := userQuery.Insert(map[string]interface{}{
        "username": "newuser",
        "email":    "newuser@example.com",
        "status":   "active",
    })
    if err != nil {
        return err
    }
    
    // 创建档案
    profileQuery := tx.Table("profiles")
    _, err = profileQuery.Insert(map[string]interface{}{
        "user_id": userID,
        "avatar":  "default.png",
    })
    if err != nil {
        return err
    }
    
    return nil
})
```

### 复杂查询

```go
err := db.Transaction(func(tx db.TransactionInterface) error {
    // 查询
    userQuery := tx.Table("users")
    user, err := userQuery.Where("id", "=", 1).First()
    if err != nil {
        return err
    }
    
    // 更新
    _, err = userQuery.Where("id", "=", 1).Update(map[string]interface{}{
        "login_count": user["login_count"].(int) + 1,
        "last_login":  time.Now(),
    })
    if err != nil {
        return err
    }
    
    // 条件删除
    logQuery := tx.Table("login_logs")
    _, err = logQuery.Where("created_at", "<", time.Now().AddDate(0, -1, 0)).Delete()
    
    return err
})
```

## ⏪ 事务回滚

### 自动回滚

闭包事务在返回错误时自动回滚：

```go
err := db.Transaction(func(tx db.TransactionInterface) error {
    _, err := tx.Exec("INSERT INTO users (name) VALUES (?)", "test")
    if err != nil {
        return err // 自动回滚
    }
    
    // 业务逻辑错误
    if someCondition {
        return errors.New("业务逻辑错误") // 自动回滚
    }
    
    return nil // 自动提交
})
```

### 手动回滚

```go
tx, err := db.BeginTransaction()
if err != nil {
    return err
}

// 执行操作
_, err = tx.Exec("INSERT INTO users (name) VALUES (?)", "test")
if err != nil {
    tx.Rollback() // 手动回滚
    return err
}

// 业务验证
if !isValid {
    tx.Rollback() // 手动回滚
    return errors.New("验证失败")
}

// 提交
return tx.Commit()
```

### Panic 恢复

```go
err := db.Transaction(func(tx db.TransactionInterface) error {
    defer func() {
        if r := recover(); r != nil {
            // panic 时自动回滚
            log.Printf("事务 panic: %v", r)
        }
    }()
    
    // 可能 panic 的操作
    riskyOperation(tx)
    
    return nil
})
```

## 📊 实际应用场景

### 订单处理

```go
func processOrder(orderID int64) error {
    return db.Transaction(func(tx db.TransactionInterface) error {
        // 1. 获取订单
        orderQuery := tx.Table("orders")
        order, err := orderQuery.Where("id", "=", orderID).First()
        if err != nil {
            return fmt.Errorf("订单不存在: %w", err)
        }
        
        if order["status"] != "pending" {
            return fmt.Errorf("订单状态不正确")
        }
        
        // 2. 获取订单项
        itemQuery := tx.Table("order_items")
        items, err := itemQuery.Where("order_id", "=", orderID).Get()
        if err != nil {
            return err
        }
        
        // 3. 检查并扣减库存
        for _, item := range items {
            productID := item["product_id"]
            quantity := item["quantity"].(int)
            
            // 检查库存
            productQuery := tx.Table("products")
            product, err := productQuery.Where("id", "=", productID).First()
            if err != nil {
                return fmt.Errorf("商品 %v 不存在", productID)
            }
            
            stock := product["stock"].(int)
            if stock < quantity {
                return fmt.Errorf("商品 %v 库存不足", productID)
            }
            
            // 扣减库存
            _, err = productQuery.Where("id", "=", productID).Update(map[string]interface{}{
                "stock": stock - quantity,
            })
            if err != nil {
                return err
            }
        }
        
        // 4. 更新订单状态
        _, err = orderQuery.Where("id", "=", orderID).Update(map[string]interface{}{
            "status":     "confirmed",
            "updated_at": time.Now(),
        })
        if err != nil {
            return err
        }
        
        // 5. 记录日志
        logQuery := tx.Table("order_logs")
        _, err = logQuery.Insert(map[string]interface{}{
            "order_id":   orderID,
            "action":     "confirmed",
            "created_at": time.Now(),
        })
        
        return err
    })
}
```

### 用户注册

```go
func registerUser(userData map[string]interface{}) (int64, error) {
    var userID int64
    
    err := db.Transaction(func(tx db.TransactionInterface) error {
        // 1. 检查用户名是否存在
        userQuery := tx.Table("users")
        existing, _ := userQuery.Where("username", "=", userData["username"]).First()
        if existing != nil {
            return fmt.Errorf("用户名已存在")
        }
        
        // 2. 检查邮箱是否存在
        existing, _ = userQuery.Where("email", "=", userData["email"]).First()
        if existing != nil {
            return fmt.Errorf("邮箱已被注册")
        }
        
        // 3. 创建用户
        id, err := userQuery.Insert(map[string]interface{}{
            "username":   userData["username"],
            "email":      userData["email"],
            "password":   hashPassword(userData["password"].(string)),
            "status":     "active",
            "created_at": time.Now(),
        })
        if err != nil {
            return err
        }
        userID = id
        
        // 4. 创建用户档案
        profileQuery := tx.Table("profiles")
        _, err = profileQuery.Insert(map[string]interface{}{
            "user_id":    userID,
            "avatar":     "default.png",
            "created_at": time.Now(),
        })
        if err != nil {
            return err
        }
        
        // 5. 分配默认角色
        roleQuery := tx.Table("user_roles")
        _, err = roleQuery.Insert(map[string]interface{}{
            "user_id":    userID,
            "role_id":    1, // 默认角色
            "created_at": time.Now(),
        })
        
        return err
    })
    
    return userID, err
}
```

## 💡 最佳实践

### 1. 优先使用闭包事务

```go
// ✅ 推荐：闭包事务自动处理提交和回滚
err := db.Transaction(func(tx db.TransactionInterface) error {
    // 操作
    return nil
})

// ⚠️ 谨慎：手动事务需要自己处理提交和回滚
tx, _ := db.BeginTransaction()
// 必须确保 Commit 或 Rollback 被调用
```

### 2. 保持事务简短

```go
// ✅ 好：事务只包含必要的数据库操作
err := db.Transaction(func(tx db.TransactionInterface) error {
    _, err := tx.Exec("UPDATE accounts SET balance = balance - ? WHERE id = ?", amount, fromID)
    if err != nil {
        return err
    }
    _, err = tx.Exec("UPDATE accounts SET balance = balance + ? WHERE id = ?", amount, toID)
    return err
})

// ❌ 避免：在事务中执行耗时操作
err := db.Transaction(func(tx db.TransactionInterface) error {
    // 不要在事务中调用外部 API
    response, _ := http.Get("https://api.example.com/verify")
    // ...
    return nil
})
```

### 3. 正确处理错误

```go
// ✅ 完整的错误处理
func createUserWithProfile(userData, profileData map[string]interface{}) error {
    err := db.Transaction(func(tx db.TransactionInterface) error {
        userID, err := createUser(tx, userData)
        if err != nil {
            return fmt.Errorf("创建用户失败: %w", err)
        }
        
        err = createProfile(tx, userID, profileData)
        if err != nil {
            return fmt.Errorf("创建档案失败: %w", err)
        }
        
        return nil
    })
    
    if err != nil {
        log.Printf("事务执行失败: %v", err)
        return err
    }
    
    return nil
}
```

### 4. 避免死锁

```go
// ✅ 按固定顺序访问资源
err := db.Transaction(func(tx db.TransactionInterface) error {
    // 总是按 ID 升序锁定账户
    ids := []int64{fromAccountID, toAccountID}
    sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
    
    for _, id := range ids {
        _, err := tx.Exec("SELECT * FROM accounts WHERE id = ? FOR UPDATE", id)
        if err != nil {
            return err
        }
    }
    
    // 执行转账
    return performTransfer(tx, fromAccountID, toAccountID, amount)
})
```

### 5. 事务粒度控制

```go
// ✅ 适当的事务粒度
func processOrders(orderIDs []int64) error {
    for _, orderID := range orderIDs {
        // 每个订单一个事务，失败不影响其他订单
        err := db.Transaction(func(tx db.TransactionInterface) error {
            return processOrder(tx, orderID)
        })
        if err != nil {
            log.Printf("订单 %d 处理失败: %v", orderID, err)
            // 继续处理其他订单
        }
    }
    return nil
}

// ❌ 避免：一个大事务处理所有订单
func processAllOrders(orderIDs []int64) error {
    return db.Transaction(func(tx db.TransactionInterface) error {
        for _, orderID := range orderIDs {
            if err := processOrder(tx, orderID); err != nil {
                return err // 一个失败，全部回滚
            }
        }
        return nil
    })
}
```

## 🔗 相关文档

- [查询构建器](Query-Builder) - 在事务中使用查询构建器
- [模型系统](Model-System) - 模型的事务支持
- [配置指南](Configuration) - 数据库连接配置
- [错误处理](API-Reference#错误处理) - 错误处理最佳实践
