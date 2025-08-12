# 事务处理

TORM 提供了完整的事务支持，包括自动事务管理、嵌套事务、保存点等功能，确保数据的一致性和完整性。

## 📋 目录

- [基础事务](#基础事务)
- [自动事务](#自动事务)
- [手动事务](#手动事务)
- [嵌套事务](#嵌套事务)
- [保存点](#保存点)
- [事务回滚](#事务回滚)
- [分布式事务](#分布式事务)
- [最佳实践](#最佳实践)

## 🚀 快速开始

### 基础事务使用

```go
import "github.com/zhoudm1743/torm/db"

// 基础事务
err := db.Transaction(func(tx db.TransactionInterface) error {
    // 在事务中执行操作
    _, err := tx.Exec("INSERT INTO users (name, email) VALUES (?, ?)", "张三", "zhangsan@example.com")
    if err != nil {
        return err // 自动回滚
    }
    
    _, err = tx.Exec("UPDATE accounts SET balance = balance - 100 WHERE user_id = ?", 1)
    if err != nil {
        return err // 自动回滚
    }
    
    return nil // 自动提交
})

if err != nil {
    log.Printf("事务失败: %v", err)
}
```

## 🔄 基础事务

### 使用闭包事务

```go
// 简单的转账事务
err := db.Transaction(func(tx db.TransactionInterface) error {
    // 检查发送方余额
    var fromBalance float64
    err := tx.QueryRow("SELECT balance FROM accounts WHERE user_id = ?", fromUserID).Scan(&fromBalance)
    if err != nil {
        return fmt.Errorf("查询发送方余额失败: %w", err)
    }
    
    if fromBalance < amount {
        return fmt.Errorf("余额不足")
    }
    
    // 扣除发送方余额
    _, err = tx.Exec("UPDATE accounts SET balance = balance - ? WHERE user_id = ?", amount, fromUserID)
    if err != nil {
        return fmt.Errorf("扣除发送方余额失败: %w", err)
    }
    
    // 增加接收方余额
    _, err = tx.Exec("UPDATE accounts SET balance = balance + ? WHERE user_id = ?", amount, toUserID)
    if err != nil {
        return fmt.Errorf("增加接收方余额失败: %w", err)
    }
    
    // 记录转账日志
    _, err = tx.Exec("INSERT INTO transfer_logs (from_user_id, to_user_id, amount) VALUES (?, ?, ?)", 
        fromUserID, toUserID, amount)
    if err != nil {
        return fmt.Errorf("记录转账日志失败: %w", err)
    }
    
    return nil
})
```

### 使用查询构建器事务

```go
err := db.Transaction(func(tx db.TransactionInterface) error {
    // 创建事务查询构建器
    txQuery := tx.Table("users")
    
    // 插入用户
    userID, err := txQuery.Insert(map[string]interface{}{
        "name":  "新用户",
        "email": "newuser@example.com",
        "status": "active",
    })
    if err != nil {
        return err
    }
    
    // 创建用户档案
    profileQuery := tx.Table("profiles")
    _, err = profileQuery.Insert(map[string]interface{}{
        "user_id": userID,
        "avatar":  "default.png",
        "bio":     "新用户",
    })
    if err != nil {
        return err
    }
    
    return nil
})
```

## 🤖 自动事务

### 模型事务

```go
// 模型自动事务
user := models.NewUser()
err := user.Transaction(func() error {
    user.Name = "张三"
    user.Email = "zhangsan@example.com"
    err := user.Save()
    if err != nil {
        return err
    }
    
    // 创建用户档案
    profile := models.NewProfile()
    profile.UserID = user.ID.(int64)
    profile.Avatar = "avatar.png"
    err = profile.Save()
    if err != nil {
        return err
    }
    
    return nil
})
```

### 批量操作事务

```go
users := []*models.User{
    {Name: "用户1", Email: "user1@example.com"},
    {Name: "用户2", Email: "user2@example.com"},
    {Name: "用户3", Email: "user3@example.com"},
}

err := db.Transaction(func(tx db.TransactionInterface) error {
    for _, user := range users {
        err := user.WithTransaction(tx).Save()
        if err != nil {
            return fmt.Errorf("保存用户 %s 失败: %w", user.Name, err)
        }
    }
    return nil
})
```

## ✋ 手动事务

### 手动开始和提交

```go
// 手动开始事务
tx, err := db.Begin()
if err != nil {
    log.Fatal(err)
}

// 确保事务会被处理
defer func() {
    if p := recover(); p != nil {
        tx.Rollback()
        panic(p) // 重新抛出panic
    } else if err != nil {
        tx.Rollback()
    } else {
        err = tx.Commit()
    }
}()

// 执行操作
_, err = tx.Exec("INSERT INTO users (name, email) VALUES (?, ?)", "手动事务用户", "manual@example.com")
if err != nil {
    return // defer中会回滚
}

_, err = tx.Exec("UPDATE statistics SET user_count = user_count + 1")
if err != nil {
    return // defer中会回滚
}

// 成功执行，defer中会提交
```

### 手动控制事务生命周期

```go
type UserService struct {
    db db.DatabaseInterface
}

func (s *UserService) CreateUserWithProfile(userData, profileData map[string]interface{}) error {
    tx, err := s.db.Begin()
    if err != nil {
        return fmt.Errorf("开始事务失败: %w", err)
    }
    
    // 创建用户
    userID, err := s.createUser(tx, userData)
    if err != nil {
        tx.Rollback()
        return fmt.Errorf("创建用户失败: %w", err)
    }
    
    // 创建档案
    profileData["user_id"] = userID
    err = s.createProfile(tx, profileData)
    if err != nil {
        tx.Rollback()
        return fmt.Errorf("创建档案失败: %w", err)
    }
    
    // 提交事务
    err = tx.Commit()
    if err != nil {
        return fmt.Errorf("提交事务失败: %w", err)
    }
    
    return nil
}

func (s *UserService) createUser(tx db.TransactionInterface, data map[string]interface{}) (int64, error) {
    return tx.Table("users").Insert(data)
}

func (s *UserService) createProfile(tx db.TransactionInterface, data map[string]interface{}) error {
    _, err := tx.Table("profiles").Insert(data)
    return err
}
```

## 🔄 嵌套事务

### 基础嵌套事务

```go
err := db.Transaction(func(tx db.TransactionInterface) error {
    // 外层事务：创建订单
    orderID, err := tx.Table("orders").Insert(map[string]interface{}{
        "user_id": userID,
        "total":   totalAmount,
        "status":  "pending",
    })
    if err != nil {
        return err
    }
    
    // 嵌套事务：处理订单项
    err = tx.Transaction(func(innerTx db.TransactionInterface) error {
        for _, item := range orderItems {
            // 检查库存
            var stock int
            err := innerTx.QueryRow("SELECT stock FROM products WHERE id = ?", item.ProductID).Scan(&stock)
            if err != nil {
                return err
            }
            
            if stock < item.Quantity {
                return fmt.Errorf("商品 %d 库存不足", item.ProductID)
            }
            
            // 创建订单项
            _, err = innerTx.Table("order_items").Insert(map[string]interface{}{
                "order_id":   orderID,
                "product_id": item.ProductID,
                "quantity":   item.Quantity,
                "price":      item.Price,
            })
            if err != nil {
                return err
            }
            
            // 减少库存
            _, err = innerTx.Exec("UPDATE products SET stock = stock - ? WHERE id = ?", 
                item.Quantity, item.ProductID)
            if err != nil {
                return err
            }
        }
        return nil
    })
    
    if err != nil {
        return fmt.Errorf("处理订单项失败: %w", err)
    }
    
    // 更新订单状态
    _, err = tx.Table("orders").Where("id", "=", orderID).Update(map[string]interface{}{
        "status": "confirmed",
    })
    
    return err
})
```

### 条件嵌套事务

```go
err := db.Transaction(func(tx db.TransactionInterface) error {
    // 主要操作
    userID, err := tx.Table("users").Insert(userData)
    if err != nil {
        return err
    }
    
    // 条件性嵌套事务
    if shouldCreateProfile {
        err = tx.Transaction(func(profileTx db.TransactionInterface) error {
            _, err := profileTx.Table("profiles").Insert(map[string]interface{}{
                "user_id": userID,
                "avatar":  "default.png",
            })
            return err
        })
        if err != nil {
            return fmt.Errorf("创建档案失败: %w", err)
        }
    }
    
    if shouldSendEmail {
        err = tx.Transaction(func(emailTx db.TransactionInterface) error {
            // 记录邮件发送日志
            _, err := emailTx.Table("email_logs").Insert(map[string]interface{}{
                "user_id": userID,
                "type":    "welcome",
                "status":  "pending",
            })
            return err
        })
        if err != nil {
            return fmt.Errorf("记录邮件日志失败: %w", err)
        }
    }
    
    return nil
})
```

## 💾 保存点

### 使用保存点

```go
err := db.Transaction(func(tx db.TransactionInterface) error {
    // 创建用户
    userID, err := tx.Table("users").Insert(userData)
    if err != nil {
        return err
    }
    
    // 创建保存点
    savepointName := "after_user_creation"
    err = tx.Savepoint(savepointName)
    if err != nil {
        return fmt.Errorf("创建保存点失败: %w", err)
    }
    
    // 尝试创建档案
    _, err = tx.Table("profiles").Insert(profileData)
    if err != nil {
        // 回滚到保存点
        rollbackErr := tx.RollbackToSavepoint(savepointName)
        if rollbackErr != nil {
            return fmt.Errorf("回滚到保存点失败: %w", rollbackErr)
        }
        
        // 记录警告但不阻止事务
        log.Printf("创建档案失败，已回滚: %v", err)
    }
    
    // 释放保存点
    err = tx.ReleaseSavepoint(savepointName)
    if err != nil {
        return fmt.Errorf("释放保存点失败: %w", err)
    }
    
    return nil
})
```

### 复杂保存点场景

```go
err := db.Transaction(func(tx db.TransactionInterface) error {
    // 主要操作
    orderID, err := tx.Table("orders").Insert(orderData)
    if err != nil {
        return err
    }
    
    // 处理每个订单项，使用保存点确保部分失败不影响整个订单
    for i, item := range orderItems {
        savepointName := fmt.Sprintf("item_%d", i)
        
        err = tx.Savepoint(savepointName)
        if err != nil {
            return err
        }
        
        // 尝试处理订单项
        err = processOrderItem(tx, orderID, item)
        if err != nil {
            // 回滚这个订单项
            tx.RollbackToSavepoint(savepointName)
            log.Printf("订单项 %d 处理失败: %v", i, err)
            
            // 记录失败的订单项
            _, logErr := tx.Table("failed_order_items").Insert(map[string]interface{}{
                "order_id":    orderID,
                "product_id":  item.ProductID,
                "error":       err.Error(),
                "created_at":  time.Now(),
            })
            if logErr != nil {
                return logErr
            }
        }
        
        tx.ReleaseSavepoint(savepointName)
    }
    
    return nil
})

func processOrderItem(tx db.TransactionInterface, orderID int64, item OrderItem) error {
    // 检查库存
    var stock int
    err := tx.QueryRow("SELECT stock FROM products WHERE id = ?", item.ProductID).Scan(&stock)
    if err != nil {
        return err
    }
    
    if stock < item.Quantity {
        return fmt.Errorf("库存不足")
    }
    
    // 创建订单项
    _, err = tx.Table("order_items").Insert(map[string]interface{}{
        "order_id":   orderID,
        "product_id": item.ProductID,
        "quantity":   item.Quantity,
    })
    if err != nil {
        return err
    }
    
    // 减少库存
    _, err = tx.Exec("UPDATE products SET stock = stock - ? WHERE id = ?", 
        item.Quantity, item.ProductID)
    
    return err
}
```

## ⏪ 事务回滚

### 自动回滚

```go
// 当函数返回错误时自动回滚
err := db.Transaction(func(tx db.TransactionInterface) error {
    _, err := tx.Exec("INSERT INTO users (name) VALUES (?)", "测试用户")
    if err != nil {
        return err // 自动回滚
    }
    
    // 模拟错误
    if someCondition {
        return errors.New("业务逻辑错误") // 自动回滚
    }
    
    return nil // 自动提交
})
```

### 手动回滚

```go
tx, err := db.Begin()
if err != nil {
    return err
}

defer func() {
    if err != nil {
        tx.Rollback()
    }
}()

// 执行操作
_, err = tx.Exec("INSERT INTO users (name) VALUES (?)", "用户")
if err != nil {
    return err
}

// 业务逻辑检查
valid, err := validateBusinessLogic(tx)
if err != nil {
    return err
}

if !valid {
    tx.Rollback()
    return errors.New("业务逻辑验证失败")
}

// 手动提交
err = tx.Commit()
return err
```

### 部分回滚处理

```go
err := db.Transaction(func(tx db.TransactionInterface) error {
    // 批量处理，部分失败不影响整体
    var successCount, failCount int
    
    for _, user := range users {
        err := processUser(tx, user)
        if err != nil {
            failCount++
            log.Printf("处理用户 %s 失败: %v", user.Name, err)
            continue
        }
        successCount++
    }
    
    // 记录处理结果
    _, err := tx.Table("batch_results").Insert(map[string]interface{}{
        "success_count": successCount,
        "fail_count":    failCount,
        "processed_at":  time.Now(),
    })
    
    if err != nil {
        return err
    }
    
    // 如果失败率太高，回滚整个批次
    if failCount > successCount {
        return errors.New("失败率过高，回滚整个批次")
    }
    
    return nil
})
```

## 🌐 分布式事务

### 两阶段提交 (2PC)

```go
// 分布式事务管理器
type DistributedTransactionManager struct {
    databases []db.DatabaseInterface
}

func (dtm *DistributedTransactionManager) Execute(operations []DistributedOperation) error {
    // 阶段1：准备阶段
    transactions := make([]db.TransactionInterface, len(dtm.databases))
    
    for i, database := range dtm.databases {
        tx, err := database.Begin()
        if err != nil {
            // 回滚已开始的事务
            dtm.rollbackAll(transactions[:i])
            return fmt.Errorf("开始事务失败: %w", err)
        }
        transactions[i] = tx
    }
    
    // 执行操作
    for i, op := range operations {
        err := op.Execute(transactions[i])
        if err != nil {
            // 回滚所有事务
            dtm.rollbackAll(transactions)
            return fmt.Errorf("执行操作 %d 失败: %w", i, err)
        }
    }
    
    // 阶段2：提交阶段
    for i, tx := range transactions {
        err := tx.Commit()
        if err != nil {
            // 如果提交失败，尝试回滚剩余事务
            dtm.rollbackAll(transactions[i+1:])
            return fmt.Errorf("提交事务 %d 失败: %w", i, err)
        }
    }
    
    return nil
}

func (dtm *DistributedTransactionManager) rollbackAll(transactions []db.TransactionInterface) {
    for _, tx := range transactions {
        if tx != nil {
            tx.Rollback()
        }
    }
}

type DistributedOperation interface {
    Execute(tx db.TransactionInterface) error
}
```

### 补偿事务模式 (Saga)

```go
type SagaStep struct {
    Execute    func(tx db.TransactionInterface) error
    Compensate func(tx db.TransactionInterface) error
}

type SagaTransaction struct {
    steps []SagaStep
}

func (saga *SagaTransaction) Execute() error {
    executedSteps := make([]int, 0)
    
    // 执行所有步骤
    for i, step := range saga.steps {
        err := db.Transaction(func(tx db.TransactionInterface) error {
            return step.Execute(tx)
        })
        
        if err != nil {
            // 执行补偿操作
            saga.compensate(executedSteps)
            return fmt.Errorf("步骤 %d 执行失败: %w", i, err)
        }
        
        executedSteps = append(executedSteps, i)
    }
    
    return nil
}

func (saga *SagaTransaction) compensate(executedSteps []int) {
    // 逆序执行补偿操作
    for i := len(executedSteps) - 1; i >= 0; i-- {
        stepIndex := executedSteps[i]
        step := saga.steps[stepIndex]
        
        err := db.Transaction(func(tx db.TransactionInterface) error {
            return step.Compensate(tx)
        })
        
        if err != nil {
            log.Printf("补偿步骤 %d 失败: %v", stepIndex, err)
            // 继续执行其他补偿操作
        }
    }
}

// 使用示例
func transferWithSaga(fromAccount, toAccount int64, amount float64) error {
    saga := &SagaTransaction{
        steps: []SagaStep{
            {
                Execute: func(tx db.TransactionInterface) error {
                    // 扣除发送方余额
                    _, err := tx.Exec("UPDATE accounts SET balance = balance - ? WHERE id = ?", 
                        amount, fromAccount)
                    return err
                },
                Compensate: func(tx db.TransactionInterface) error {
                    // 恢复发送方余额
                    _, err := tx.Exec("UPDATE accounts SET balance = balance + ? WHERE id = ?", 
                        amount, fromAccount)
                    return err
                },
            },
            {
                Execute: func(tx db.TransactionInterface) error {
                    // 增加接收方余额
                    _, err := tx.Exec("UPDATE accounts SET balance = balance + ? WHERE id = ?", 
                        amount, toAccount)
                    return err
                },
                Compensate: func(tx db.TransactionInterface) error {
                    // 恢复接收方余额
                    _, err := tx.Exec("UPDATE accounts SET balance = balance - ? WHERE id = ?", 
                        amount, toAccount)
                    return err
                },
            },
        },
    }
    
    return saga.Execute()
}
```

## 📚 最佳实践

### 1. 事务粒度控制

```go
// 好的做法：事务粒度适中
err := db.Transaction(func(tx db.TransactionInterface) error {
    // 只包含相关的操作
    userID, err := tx.Table("users").Insert(userData)
    if err != nil {
        return err
    }
    
    _, err = tx.Table("profiles").Insert(map[string]interface{}{
        "user_id": userID,
        "avatar":  "default.png",
    })
    return err
})

// 避免：事务过大
// 不要在一个事务中包含不相关的操作或长时间运行的操作
```

### 2. 错误处理

```go
// 好的做法：完整的错误处理
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
    // 处理事务失败的情况
}
```

### 3. 超时控制

```go
// 设置事务超时
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

err := db.TransactionWithContext(ctx, func(tx db.TransactionInterface) error {
    // 长时间运行的操作
    return processLargeDataset(tx)
})

if err != nil {
    if errors.Is(err, context.DeadlineExceeded) {
        log.Println("事务超时")
    } else {
        log.Printf("事务失败: %v", err)
    }
}
```

### 4. 死锁预防

```go
// 好的做法：按固定顺序访问资源
err := db.Transaction(func(tx db.TransactionInterface) error {
    // 总是按 ID 升序锁定账户
    accountIDs := []int64{fromAccountID, toAccountID}
    sort.Slice(accountIDs, func(i, j int) bool {
        return accountIDs[i] < accountIDs[j]
    })
    
    for _, accountID := range accountIDs {
        _, err := tx.Exec("SELECT balance FROM accounts WHERE id = ? FOR UPDATE", accountID)
        if err != nil {
            return err
        }
    }
    
    // 执行转账操作
    return performTransfer(tx, fromAccountID, toAccountID, amount)
})
```

### 5. 资源清理

```go
// 确保资源得到正确清理
func processWithTransaction() error {
    tx, err := db.Begin()
    if err != nil {
        return err
    }
    
    // 确保事务被正确处理
    defer func() {
        if p := recover(); p != nil {
            tx.Rollback()
            panic(p)
        } else if err != nil {
            tx.Rollback()
        } else {
            err = tx.Commit()
        }
    }()
    
    // 执行业务逻辑
    err = performBusinessLogic(tx)
    return err
}
```

## 🔗 相关文档

- [查询构建器](Query-Builder) - 在事务中使用查询构建器
- [模型系统](Model-System) - 模型的事务支持
- [性能优化](Performance) - 事务性能优化
- [故障排除](Troubleshooting) - 事务相关问题解决 