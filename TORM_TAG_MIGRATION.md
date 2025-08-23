# TORM 标签迁移指南

TORM v1.1.6 引入了统一的 `torm` 标签，大大简化了模型定义。新标签完全支持**大小写不敏感**，本文档展示如何从旧标签迁移到新标签。

## 为什么使用统一标签？

### 旧方式的问题
- 标签过多：`primaryKey`, `autoIncrement`, `size`, `unique`, `default`, `comment` 等
- 语法不一致：有些用 `="true"`，有些用 `="值"`
- 可读性差：一个字段可能有 5-6 个不同的标签

### 新方式的优势
- **统一语法**：只需要一个 `torm` 标签
- **更简洁**：用逗号分隔多个属性
- **更易读**：所有属性都在一个标签内
- **大小写不敏感**：支持任何大小写组合
- **向后兼容**：旧标签依然支持

## 迁移对照表

### 主键和约束

| 旧标签 | 新标签 | 说明 |
|--------|--------|------|
| `primaryKey:"true"` | `torm:"primary_key"` | 主键 |
| `pk:""` | `torm:"pk"` | 主键简写 |
| `autoIncrement:"true"` | `torm:"auto_increment"` | 自增 |
| `unique:"true"` | `torm:"unique"` | 唯一约束 |
| `nullable:"true"` | `torm:"nullable"` | 允许NULL |
| `not_null:"true"` | `torm:"not_null"` | 不允许NULL |

### 数据类型

| 旧标签 | 新标签 | 说明 |
|--------|--------|------|
| `type:"varchar" size:"100"` | `torm:"type:varchar,size:100"` | 字符串类型+长度 |
| `type:"decimal" precision:"10" scale:"2"` | `torm:"type:decimal,precision:10,scale:2"` | 精确数值 |
| `size:"100"` | `torm:"size:100"` | 字段长度 |

### 索引

| 旧标签 | 新标签 | 说明 |
|--------|--------|------|
| `index:"true"` | `torm:"index"` | 普通索引 |
| `index:"custom_name"` | `torm:"index:custom_name"` | 自定义索引名 |

### 默认值和时间戳

| 旧标签 | 新标签 | 说明 |
|--------|--------|------|
| `default:"value"` | `torm:"default:value"` | 默认值 |
| `autoCreateTime:"true"` | `torm:"auto_create_time"` | 创建时间 |
| `autoUpdateTime:"true"` | `torm:"auto_update_time"` | 更新时间 |
| `comment:"描述"` | `torm:"comment:描述"` | 字段注释 |

## 实际迁移示例

### 旧方式
```go
type User struct {
    model.BaseModel
    ID        int64     `json:"id" db:"id" primaryKey:"true" autoIncrement:"true" comment:"用户ID"`
    Email     string    `json:"email" db:"email" size:"100" unique:"true" comment:"邮箱"`
    Name      string    `json:"name" db:"name" size:"50" comment:"姓名"`
    Balance   float64   `json:"balance" db:"balance" type:"decimal" precision:"10" scale:"2" default:"0.00"`
    Status    string    `json:"status" db:"status" default:"active" comment:"状态"`
    Avatar    *string   `json:"avatar" db:"avatar" nullable:"true" comment:"头像"`
    CreatedAt int64     `json:"created_at" db:"created_at" autoCreateTime:"true" comment:"创建时间"`
    UpdatedAt int64     `json:"updated_at" db:"updated_at" autoUpdateTime:"true" comment:"更新时间"`
}
```

### 新方式
```go
type User struct {
    model.BaseModel
    ID        int64     `json:"id" db:"id" torm:"primary_key,auto_increment,comment:用户ID"`
    Email     string    `json:"email" db:"email" torm:"size:100,unique,comment:邮箱"`
    Name      string    `json:"name" db:"name" torm:"size:50,comment:姓名"`
    Balance   float64   `json:"balance" db:"balance" torm:"type:decimal,precision:10,scale:2,default:0.00"`
    Status    string    `json:"status" db:"status" torm:"default:active,comment:状态"`
    Avatar    *string   `json:"avatar" db:"avatar" torm:"nullable,comment:头像"`
    CreatedAt int64     `json:"created_at" db:"created_at" torm:"auto_create_time,comment:创建时间"`
    UpdatedAt int64     `json:"updated_at" db:"updated_at" torm:"auto_update_time,comment:更新时间"`
}
```

### 对比优势

| 对比项 | 旧方式 | 新方式 | 改进 |
|--------|--------|--------|------|
| **标签数量** | 16 个标签 | 8 个标签 | 减少 50% |
| **字符数** | ~400 字符 | ~280 字符 | 减少 30% |
| **可读性** | 分散，难以快速理解 | 集中，一目了然 | 显著提升 |
| **维护性** | 修改多个标签 | 修改一个标签 | 更容易维护 |

## 渐进式迁移

TORM 支持向后兼容，您可以渐进式迁移：

### 1. 混合使用
```go
type User struct {
    model.BaseModel
    // 新字段使用新标签
    ID        int64  `db:"id" torm:"primary_key,auto_increment,comment:用户ID"`
    
    // 旧字段保持原样
    Email     string `db:"email" size:"100" unique:"true" comment:"邮箱"`
    
    // 逐步迁移
    Name      string `db:"name" torm:"size:50,comment:姓名"`
}
```

### 2. 批量迁移脚本

```bash
# 简单的查找替换（仅作参考，请根据实际情况调整）
sed -i 's/primaryKey:"true"/torm:"primary_key"/g' *.go
sed -i 's/autoIncrement:"true"/torm:"auto_increment"/g' *.go
sed -i 's/unique:"true"/torm:"unique"/g' *.go
```

## 最佳实践

### 1. 新项目
- 直接使用新的 `torm` 标签
- 享受简洁、统一的语法

### 2. 现有项目
- 优先迁移新增字段
- 重构时同步迁移旧字段
- 利用向后兼容性平滑过渡

### 3. 团队协作
- 统一代码风格，建议全团队使用新标签
- 在代码规范中明确标签使用方式
- 通过 Code Review 确保一致性

## 类型长度和精度详解

### VARCHAR 长度控制

```go
type StringFields struct {
    ShortCode   string `torm:"type:varchar,size:10"`     // VARCHAR(10)   - 短编码
    Name        string `torm:"type:varchar,size:50"`     // VARCHAR(50)   - 普通名称  
    Description string `torm:"type:varchar,size:200"`    // VARCHAR(200)  - 长描述
    LongText    string `torm:"type:text"`                // TEXT          - 超长文本
}
```

### CHAR 固定长度

```go
type FixedFields struct {
    CountryCode string `torm:"type:char,size:2"`         // CHAR(2)       - 国家代码
    FixedCode   string `torm:"type:char,size:8"`         // CHAR(8)       - 固定编码
}
```

### DECIMAL 精度控制

```go
type NumericFields struct {
    // DECIMAL(precision, scale) - precision总位数，scale小数位数
    Price      float64 `torm:"type:decimal,precision:10,scale:2"`  // DECIMAL(10,2) - 最大8位整数,2位小数
    Rate       float64 `torm:"type:decimal,precision:5,scale:4"`   // DECIMAL(5,4)  - 最大1位整数,4位小数
    Amount     float64 `torm:"type:decimal,precision:15,scale:2"`  // DECIMAL(15,2) - 最大13位整数,2位小数
    Percentage float64 `torm:"type:decimal,precision:6,scale:3"`   // DECIMAL(6,3)  - 最大3位整数,3位小数
}
```

### 实际业务场景

| 业务需求 | 数据范例 | 推荐类型 | TORM 标签 |
|----------|----------|----------|-----------|
| **电商价格** | 99.99 | DECIMAL(10,2) | `torm:"type:decimal,precision:10,scale:2"` |
| **银行利率** | 0.0325 | DECIMAL(5,4) | `torm:"type:decimal,precision:5,scale:4"` |
| **大额转账** | 1234567.89 | DECIMAL(15,2) | `torm:"type:decimal,precision:15,scale:2"` |
| **商品编码** | "P12345" | VARCHAR(10) | `torm:"type:varchar,size:10"` |
| **手机号码** | "13812345678" | VARCHAR(15) | `torm:"type:varchar,size:15"` |
| **国家代码** | "CN" | CHAR(2) | `torm:"type:char,size:2"` |
| **身份证号** | "123456789012345678" | CHAR(18) | `torm:"type:char,size:18"` |

### 大小写不敏感示例

所有这些写法都完全等效，会生成相同的数据库结构：

```go
type CaseExample struct {
    model.BaseModel
    // 1. 全小写（推荐风格）
    Field1 float64 `torm:"type:decimal,precision:10,scale:2,default:0.00,comment:价格"`
    
    // 2. 全大写
    Field2 float64 `torm:"TYPE:DECIMAL,PRECISION:10,SCALE:2,DEFAULT:0.00,COMMENT:价格"`
    
    // 3. 首字母大写
    Field3 float64 `torm:"Type:Decimal,Precision:10,Scale:2,Default:0.00,Comment:价格"`
    
    // 4. 混合大小写
    Field4 float64 `torm:"TYPE:decimal,PRECISION:10,scale:2,DEFAULT:0.00,comment:价格"`
}
```

**支持的大小写组合：**
- 标志位：`primary_key`, `PRIMARY_KEY`, `Primary_Key`, `pRiMaRy_KeY`
- 类型：`varchar`, `VARCHAR`, `VarChar`, `varchar`, `VaRcHaR`
- 属性：`size`, `SIZE`, `Size`, `precision`, `PRECISION`, `Precision`
- 默认值：`true`, `TRUE`, `True`, `null`, `NULL`, `Null`

### 组合使用示例

```go
type Product struct {
    model.BaseModel
    // 完整的字段定义
    ID          int64   `db:"id" torm:"primary_key,auto_increment,comment:产品ID"`
    SKU         string  `db:"sku" torm:"type:varchar,size:20,unique,comment:产品编码"`
    Name        string  `db:"name" torm:"type:varchar,size:100,comment:产品名称"`
    Price       float64 `db:"price" torm:"type:decimal,precision:10,scale:2,default:0.00,comment:售价"`
    Weight      float64 `db:"weight" torm:"type:decimal,precision:8,scale:3,comment:重量(kg)"`
    CategoryID  int64   `db:"category_id" torm:"index,comment:分类ID"`
    IsActive    bool    `db:"is_active" torm:"default:true,comment:是否上架"`
    CreatedAt   int64   `db:"created_at" torm:"auto_create_time,comment:创建时间"`
}
```

## 总结

TORM v1.1.6 的统一标签是一个重大改进：

- ✅ **简化语法**：从多个标签到一个标签
- ✅ **提升可读性**：所有属性集中显示
- ✅ **精确控制**：完整支持类型长度和精度
- ✅ **向后兼容**：旧项目无需立即迁移
- ✅ **更易维护**：减少标签数量，降低维护成本
- ✅ **业务友好**：贴近实际业务场景的类型控制

推荐所有新项目都使用新的 `torm` 标签语法！
