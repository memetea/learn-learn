package models

import (
	"time"
)

type UserStatus int

const (
	StatusInactive UserStatus = iota
	StatusActive
	StatusPending
	StatusSuspended
)

func (s UserStatus) String() string {
	switch s {
	case StatusInactive:
		return "inactive"
	case StatusActive:
		return "active"
	case StatusPending:
		return "pending"
	case StatusSuspended:
		return "suspended"
	default:
		return "unknown"
	}
}

// 定义一个你自己的基础模型结构体
type BaseModel struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type User struct {
	BaseModel
	Username     string     `gorm:"unique;not null"`
	Password     string     `gorm:"not null"`
	Roles        []Role     `gorm:"many2many:user_roles;"`
	TokenVersion uint       `gorm:"default:1"`          // 添加 TokenVersion 字段
	Status       UserStatus `gorm:"not null,default:0"` // 新增字段，用于表示用户是否激活
}

type Role struct {
	ID          uint         `gorm:"primarykey"`
	Name        string       `gorm:"unique;not null"`
	Permissions []Permission `gorm:"many2many:roles_permissions;"`
}

// Permission represents a permission in the system
type Permission struct {
	ID          uint   `gorm:"primaryKey"`
	Name        string `gorm:"unique;not null"`
	Description string `gorm:""`
}

// GetRoles 返回用户角色名称的数组
func (u *User) GetRoles() []string {
	var roleNames []string
	for _, role := range u.Roles {
		roleNames = append(roleNames, role.Name)
	}
	return roleNames
}

func (r *Role) GetPermissions() []string {
	var permissionNames []string
	// for _, permission := range r.Permissions {
	// 	permissionNames = append(permissionNames, permission.Name)
	// }
	return permissionNames
}
