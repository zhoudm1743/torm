package models

import (
	"log"

	"github.com/zhoudm1743/torm/model"
)

type Admin struct {
	model.BaseModel
	ID        string   `json:"id" torm:"primary_key,type:varchar,size:32,comment:管理员ID"`
	Phone     string   `json:"phone" torm:"type:varchar,size:11,comment:手机号"`
	Password  string   `json:"password" torm:"type:varchar,size:32,comment:密码"`
	Nickname  string   `json:"nickname" torm:"type:varchar,size:32,comment:昵称"`
	Avatar    string   `json:"avatar" torm:"type:varchar,size:255,comment:头像"`
	Status    int      `json:"status" torm:"type:int,default:1,comment:状态"`
	Role      []string `json:"role" torm:"type:varchar,size:255,comment:角色（JSON存储）"`
	CreatedAt int64    `json:"created_at" torm:"auto_create_time,comment:创建时间"`
	UpdatedAt int64    `json:"updated_at" torm:"auto_update_time,comment:更新时间"`
}

func NewAdmin() *Admin {
	admin := &Admin{BaseModel: *model.NewBaseModel()}
	admin.SetTable("admin")
	admin.SetPrimaryKey("id")
	admin.SetConnection("default")

	// 关键：必须先调用 DetectConfigFromStruct
	admin.DetectConfigFromStruct(admin)

	// 然后再调用 AutoMigrate
	err := admin.AutoMigrate()
	if err != nil {
		log.Printf("AutoMigrate failed: %v", err)
	} else {
		log.Println("AutoMigrate completed successfully")
	}

	return admin
}

// 或者使用更简洁的方式（推荐）
func NewAdminSimple() *Admin {
	admin := &Admin{}
	// 使用 NewBaseModelWithAutoDetect 自动调用 DetectConfigFromStruct
	admin.BaseModel = *model.NewBaseModelWithAutoDetect(admin)
	admin.SetTable("admin")
	admin.SetPrimaryKey("id")
	admin.SetConnection("default")

	err := admin.AutoMigrate()
	if err != nil {
		log.Printf("AutoMigrate failed: %v", err)
	} else {
		log.Println("AutoMigrate completed successfully")
	}

	return admin
}
