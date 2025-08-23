# TORM 标签全面使用指南

## 🎯 概述

经过全面扩展和测试，TORM现在支持超过**44种不同的标签写法**，涵盖所有常见的数据库字段定义需求。您说得对，是`not null`而不是`not nil`，我们现在同时支持多种写法以提供最大的灵活性。

## 📋 完整的标签支持列表

### 🔑 键值对标签 (16种)

#### 基础属性
- `type:varchar` - 字段类型
- `size:255` - 字段长度  
- `length:100` - 长度别名
- `len:50` - 长度简写
- `width:200` - 宽度(长度别名)

#### 数值精度
- `precision:10` - 数值精度
- `prec:8` - 精度别名
- `scale:2` - 小数位数
- `digits:3` - 小数位别名

#### 特殊属性
- `default:1` - 默认值
- `comment:用户名` - 字段注释
- `column:custom_name` - 自定义列名

#### 时间相关
- `auto_update:current_timestamp` - 自动更新时间
- `on_update:current_timestamp` - 更新时动作
- `auto_create:current_timestamp` - 自动创建时间  
- `on_create:current_timestamp` - 创建时动作

### 🚩 标志标签 (28种)

#### 主键标志 (4种写法)
- `primary_key` ✅
- `pk` ✅
- `primary` ✅
- `primarykey` ✅

#### 自增标志 (4种写法)
- `auto_increment` ✅
- `autoincrement` ✅
- `auto_inc` ✅
- `autoinc` ✅

#### 唯一约束 (2种写法)
- `unique` ✅
- `uniq` ✅

#### 非空约束 (5种写法) - 您说得对！
- `not_null` ✅ (标准写法)
- `not null` ✅ (SQL标准空格写法)
- `notnull` ✅ (连写)
- `not_nil` ✅ (Go风格)
- `notnil` ✅ (Go风格连写)

#### 可空约束 (2种写法)
- `nullable` ✅
- `null` ✅

#### 时间戳标志 (10种写法)
- `auto_create_time` ✅
- `create_time` ✅
- `created_at` ✅
- `auto_created_at` ✅
- `auto_update_time` ✅
- `update_time` ✅
- `updated_at` ✅
- `auto_updated_at` ✅
- `timestamp` ✅
- `current_timestamp` ✅

#### JSON字段标志 (2种写法)
- `json` ✅
- `is_json` ✅

## 🎨 支持的数据类型 (50+种)

### 字符串类型
```go
Field1 string `torm:"type:varchar,size:255"`      // VARCHAR
Field2 string `torm:"type:string,size:100"`       // VARCHAR别名
Field3 string `torm:"type:char,size:10"`          // CHAR
Field4 string `torm:"type:character,size:5"`      // CHAR别名
Field5 string `torm:"type:text"`                  // TEXT
Field6 string `torm:"type:longtext"`              // LONGTEXT
Field7 string `torm:"type:mediumtext"`            // TEXT
Field8 string `torm:"type:tinytext"`              // TEXT
```

### 整数类型
```go
Field1  int   `torm:"type:int"`           // INT
Field2  int   `torm:"type:integer"`       // INT别名
Field3  int32 `torm:"type:int32"`         // INT
Field4  int8  `torm:"type:tinyint"`       // TINYINT
Field5  int8  `torm:"type:int8"`          // TINYINT别名
Field6  int8  `torm:"type:byte"`          // TINYINT别名
Field7  int16 `torm:"type:smallint"`      // SMALLINT
Field8  int16 `torm:"type:int16"`         // SMALLINT别名
Field9  int16 `torm:"type:short"`         // SMALLINT别名
Field10 int64 `torm:"type:bigint"`        // BIGINT
Field11 int64 `torm:"type:int64"`         // BIGINT别名
Field12 int64 `torm:"type:long"`          // BIGINT别名
Field13 int   `torm:"type:mediumint"`     // INT
```

### 浮点类型
```go
Field1 float32 `torm:"type:float"`                    // FLOAT
Field2 float32 `torm:"type:float32"`                  // FLOAT
Field3 float32 `torm:"type:real"`                     // FLOAT
Field4 float64 `torm:"type:double"`                   // DOUBLE
Field5 float64 `torm:"type:float64"`                  // DOUBLE
Field6 float64 `torm:"type:double_precision"`         // DOUBLE
Field7 float64 `torm:"type:decimal,precision:10,scale:2"` // DECIMAL
Field8 float64 `torm:"type:numeric,precision:12,scale:4"` // DECIMAL
Field9 float64 `torm:"type:money,precision:8,scale:2"`    // DECIMAL
```

### 布尔类型
```go
Field1 bool `torm:"type:boolean"`     // BOOLEAN
Field2 bool `torm:"type:bool"`        // BOOLEAN
Field3 bool `torm:"type:bit"`         // BOOLEAN
```

### 日期时间类型
```go
Field1 string `torm:"type:date"`          // DATE
Field2 string `torm:"type:datetime"`      // DATETIME
Field3 string `torm:"type:datetime2"`     // DATETIME
Field4 string `torm:"type:timestamp"`     // TIMESTAMP
Field5 string `torm:"type:timestamptz"`   // TIMESTAMP
Field6 string `torm:"type:time"`          // TIME
Field7 string `torm:"type:timetz"`        // TIME
Field8 int    `torm:"type:year"`          // INT(年份)
```

### 二进制类型
```go
Field1 []byte `torm:"type:blob"`         // BLOB
Field2 []byte `torm:"type:binary"`       // BLOB
Field3 []byte `torm:"type:varbinary"`    // BLOB
Field4 []byte `torm:"type:tinyblob"`     // BLOB
Field5 []byte `torm:"type:mediumblob"`   // BLOB
Field6 []byte `torm:"type:longblob"`     // BLOB
```

### 特殊类型
```go
// JSON类型
Field1 interface{} `torm:"type:json"`        // JSON
Field2 interface{} `torm:"type:jsonb"`       // JSON

// UUID类型
Field3 string `torm:"type:uuid"`             // VARCHAR(36)
Field4 string `torm:"type:guid"`             // VARCHAR(36)

// 枚举类型  
Field5 string `torm:"type:enum"`             // VARCHAR(255)
Field6 string `torm:"type:set"`              // VARCHAR(255)

// 几何类型
Field7 string `torm:"type:geometry"`         // TEXT
Field8 string `torm:"type:point"`            // TEXT
Field9 string `torm:"type:linestring"`       // TEXT
Field10 string `torm:"type:polygon"`         // TEXT

// 其他类型
Field11 string      `torm:"type:xml"`         // TEXT
Field12 string      `torm:"type:inet"`        // VARCHAR
Field13 string      `torm:"type:cidr"`        // VARCHAR
Field14 string      `torm:"type:macaddr"`     // VARCHAR
Field15 interface{} `torm:"type:array"`       // JSON
```

## 🏆 完整示例模型

```go
type ComprehensiveUser struct {
    model.BaseModel
    
    // 主键 - 多种写法
    ID string `json:"id" torm:"primary_key,type:varchar,size:32,comment:用户ID"`
    
    // 字符串字段 - 各种约束
    Username string `json:"username" torm:"column:user_name,type:varchar,size:50,unique,not_null,comment:用户名"`
    Email    string `json:"email" torm:"type:varchar,length:255,uniq,not null,comment:邮箱地址"`
    Phone    string `json:"phone" torm:"type:varchar,len:11,notnull,comment:手机号"`
    Nickname string `json:"nickname" torm:"type:varchar,width:100,nullable,comment:昵称"`
    
    // 数值字段 - 精度控制
    Age      int     `json:"age" torm:"type:int,default:18,comment:年龄"`
    Salary   float64 `json:"salary" torm:"type:decimal,precision:10,scale:2,comment:薪资"`
    Score    float64 `json:"score" torm:"type:numeric,prec:5,digits:2,default:0.00,comment:评分"`
    
    // 布尔字段
    IsActive bool `json:"is_active" torm:"type:boolean,default:true,comment:是否激活"`
    IsVIP    bool `json:"is_vip" torm:"type:bool,default:false,comment:是否VIP"`
    
    // 自增字段 - 多种写法
    SerialNo int64 `json:"serial_no" torm:"type:bigint,auto_increment,comment:序列号"`
    OrderNum int64 `json:"order_num" torm:"type:int64,autoinc,comment:订单号"`
    
    // 时间字段 - 多种写法
    CreatedAt int64 `json:"created_at" torm:"auto_create_time,comment:创建时间"`
    UpdatedAt int64 `json:"updated_at" torm:"auto_update_time,comment:更新时间"`
    LoginAt   int64 `json:"login_at" torm:"timestamp,comment:登录时间"`
    DeletedAt int64 `json:"deleted_at" torm:"auto_create:current_timestamp,nullable,comment:删除时间"`
    
    // JSON字段
    Profile  interface{} `json:"profile" torm:"type:json,comment:用户资料"`
    Settings interface{} `json:"settings" torm:"json,comment:用户设置"`
    Tags     []string    `json:"tags" torm:"type:varchar,size:500,comment:用户标签"`
    
    // 特殊类型
    Avatar    string `json:"avatar" torm:"type:text,comment:头像URL"`
    UUID      string `json:"uuid" torm:"type:uuid,comment:全局唯一标识"`
    Metadata  string `json:"metadata" torm:"type:jsonb,comment:元数据"`
    
    // 二进制字段
    Photo []byte `json:"photo" torm:"type:blob,comment:用户照片"`
    
    // 自定义列名
    InternalCode string `json:"internal_code" torm:"column:internal_user_code,type:varchar,size:64,comment:内部编码"`
}
```

## 🚀 使用建议

### 1. 推荐的标签组合
```go
// 主键字段
ID string `torm:"primary_key,type:varchar,size:32,comment:主键ID"`

// 唯一字段  
Email string `torm:"type:varchar,size:255,unique,not_null,comment:邮箱"`

// 必填字段
Name string `torm:"type:varchar,size:100,not null,comment:姓名"`

// 可选字段
Bio string `torm:"type:text,nullable,comment:个人简介"`

// 数值字段
Price float64 `torm:"type:decimal,precision:10,scale:2,default:0.00,comment:价格"`

// 时间字段
CreatedAt int64 `torm:"auto_create_time,comment:创建时间"`
UpdatedAt int64 `torm:"auto_update_time,comment:更新时间"`
```

### 2. 类型选择指南
- **主键**: `varchar(32)` 或 `bigint auto_increment`
- **用户名/邮箱**: `varchar(255) unique not_null`  
- **手机号**: `varchar(11) not_null`
- **密码**: `varchar(255) not_null` (加密后)
- **金额**: `decimal(10,2)` 
- **状态**: `int default:1`
- **布尔值**: `boolean default:false`
- **JSON数据**: `json` 或 `text`
- **时间戳**: `auto_create_time` / `auto_update_time`

### 3. 兼容性说明
- ✅ 支持 `not_null` 和 `not null`(带空格)
- ✅ 支持 `auto_increment` 和 `autoinc` 等简写
- ✅ 支持 `size`、`length`、`len`、`width` 等长度别名
- ✅ 支持 `precision`/`prec` 和 `scale`/`digits` 等精度别名
- ✅ 支持多种时间戳写法
- ✅ 支持自定义列名映射

## 🎉 总结

现在TORM支持**44种标签写法**，涵盖：
- **16种键值对标签** - 属性设置
- **28种标志标签** - 约束和特性
- **50+种数据类型** - 完整类型覆盖

您可以使用最符合您习惯的写法，无论是SQL标准的`not null`，Go风格的`not_nil`，还是简洁的`notnull`，TORM都能正确识别和处理！

所有这些标签的修改都会被自动迁移系统正确检测并生成相应的ALTER TABLE语句。🚀
