package model

import (
	"fmt"
	"time"

	"github.com/zhoudm1743/torm/db"
)

// AccessorDemoUser 演示访问器系统的用户模型
type AccessorDemoUser struct {
	BaseModel
	ID        uint       `json:"id" torm:"primary_key,auto_increment"`
	Username  string     `json:"username" torm:"type:varchar,size:50,unique"`
	Status    int        `json:"status" torm:"type:int,default:1"`
	Age       int        `json:"age" torm:"type:int,default:0"`
	Balance   int        `json:"balance" torm:"type:int,default:0"` // 以分为单位存储
	CreatedAt time.Time  `json:"created_at" torm:"auto_create_time"`
	UpdatedAt time.Time  `json:"updated_at" torm:"auto_update_time"`
	DeletedAt *time.Time `json:"deleted_at" torm:"soft_delete"`
}

func (u *AccessorDemoUser) GetTableName() string {
	return "accessor_demo_users"
}

// GetStatusAttr 状态访问器 - 将数字转换为可读状态
func (u *AccessorDemoUser) GetStatusAttr(value interface{}) interface{} {
	status := 1
	if v, ok := value.(int); ok {
		status = v
	} else if v, ok := value.(int64); ok {
		status = int(v)
	} else if v, ok := value.(float64); ok {
		status = int(v)
	}

	statusMap := map[int]map[string]interface{}{
		0: {"code": 0, "name": "禁用", "color": "red", "can_login": false},
		1: {"code": 1, "name": "正常", "color": "green", "can_login": true},
		2: {"code": 2, "name": "待审核", "color": "orange", "can_login": false},
	}

	if result, exists := statusMap[status]; exists {
		return result
	}
	return statusMap[1]
}

// SetStatusAttr 状态设置器 - 支持多种输入格式
func (u *AccessorDemoUser) SetStatusAttr(value interface{}) interface{} {
	if str, ok := value.(string); ok {
		switch str {
		case "禁用", "disabled":
			return 0
		case "正常", "active":
			return 1
		case "待审核", "pending":
			return 2
		}
	}

	if v, ok := value.(int); ok {
		return v
	}
	if v, ok := value.(int64); ok {
		return int(v)
	}
	if v, ok := value.(float64); ok {
		return int(v)
	}

	return 1
}

// GetBalanceAttr 余额访问器 - 智能金额格式化
func (u *AccessorDemoUser) GetBalanceAttr(value interface{}) interface{} {
	cents := 0
	if v, ok := value.(int); ok {
		cents = v
	} else if v, ok := value.(int64); ok {
		cents = int(v)
	} else if v, ok := value.(float64); ok {
		cents = int(v)
	}

	yuan := float64(cents) / 100.0

	return map[string]interface{}{
		"cents":       cents,
		"yuan":        yuan,
		"formatted":   fmt.Sprintf("¥%.2f", yuan),
		"level":       getBalanceLevel(yuan),
		"is_positive": cents > 0,
	}
}

// SetBalanceAttr 余额设置器 - 支持元和分的输入
func (u *AccessorDemoUser) SetBalanceAttr(value interface{}) interface{} {
	if str, ok := value.(string); ok {
		// 如果是字符串，尝试解析（简单实现）
		if str == "0" || str == "" {
			return 0
		}
		// 这里可以添加更复杂的字符串解析逻辑
		return 0
	}

	if v, ok := value.(float64); ok {
		// 如果是浮点数，假设是元，转换为分
		return int(v * 100)
	}

	if v, ok := value.(int); ok {
		return v // 假设已经是分
	}

	return 0
}

// GetAgeAttr 年龄访问器 - 添加年龄组信息
func (u *AccessorDemoUser) GetAgeAttr(value interface{}) interface{} {
	age := 0
	if v, ok := value.(int); ok {
		age = v
	} else if v, ok := value.(int64); ok {
		age = int(v)
	} else if v, ok := value.(float64); ok {
		age = int(v)
	}

	var ageGroup string
	if age < 18 {
		ageGroup = "未成年"
	} else if age < 30 {
		ageGroup = "青年"
	} else if age < 50 {
		ageGroup = "中年"
	} else {
		ageGroup = "老年"
	}

	return map[string]interface{}{
		"value":    age,
		"group":    ageGroup,
		"is_adult": age >= 18,
	}
}

// getBalanceLevel 获取余额等级（辅助函数）
func getBalanceLevel(yuan float64) string {
	if yuan <= 0 {
		return "无余额"
	} else if yuan < 100 {
		return "低余额"
	} else if yuan < 1000 {
		return "中余额"
	} else {
		return "高余额"
	}
}

// DemoAccessorUsage 演示访问器系统的使用
func DemoAccessorUsage() {
	// 1. 创建模型实例
	user := NewModel(&AccessorDemoUser{})
	user.SetConnection("default")

	// 2. 使用设置器设置数据
	user.SetAttributeWithAccessor(&AccessorDemoUser{}, "status", "待审核")
	user.SetAttributeWithAccessor(&AccessorDemoUser{}, "balance", 150.50) // 150.50元
	user.SetAttribute("username", "demo_user")
	user.SetAttribute("age", 25)

	// 3. 模拟查询结果数据
	mockData := []map[string]interface{}{
		{
			"id":       1,
			"username": "user1",
			"status":   1,
			"age":      25,
			"balance":  15050, // 150.50元 = 15050分
		},
		{
			"id":       2,
			"username": "user2",
			"status":   0,
			"age":      17,
			"balance":  5000, // 50.00元 = 5000分
		},
	}

	// 4. 应用访问器处理
	processor := db.NewAccessorProcessor(&AccessorDemoUser{})
	processedData := processor.ProcessDataSlice(mockData)

	// 5. 查看处理后的结果
	for i, record := range processedData {
		println(fmt.Sprintf("用户 %d:", i+1))
		println(fmt.Sprintf("  用户名: %s", record["username"]))

		if statusData, ok := record["status"].(map[string]interface{}); ok {
			println(fmt.Sprintf("  状态: %s (%s)", statusData["name"], statusData["color"]))
			println(fmt.Sprintf("  可登录: %v", statusData["can_login"]))
		}

		if ageData, ok := record["age"].(map[string]interface{}); ok {
			println(fmt.Sprintf("  年龄: %d岁 (%s)", ageData["value"], ageData["group"]))
			println(fmt.Sprintf("  是否成年: %v", ageData["is_adult"]))
		}

		if balanceData, ok := record["balance"].(map[string]interface{}); ok {
			println(fmt.Sprintf("  余额: %s (%s)", balanceData["formatted"], balanceData["level"]))
		}

		println("")
	}
}
