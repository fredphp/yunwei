package server

import (
	"time"

	"gorm.io/gorm"
)

// SshKey SSH 密钥
type SshKey struct {
	ID          uint           `json:"id" gorm:"primarykey"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	// 基本信息
	Name        string `json:"name" gorm:"type:varchar(64);not null;comment:密钥名称"`
	Filename    string `json:"filename" gorm:"type:varchar(128);comment:原始文件名"`
	KeyContent  string `json:"-" gorm:"type:text;not null;comment:PEM私钥内容"`
	Passphrase  string `json:"-" gorm:"type:varchar(128);comment:密钥密码"`
	Fingerprint string `json:"fingerprint" gorm:"type:varchar(64);comment:密钥指纹"`
	Description string `json:"description" gorm:"type:varchar(255);comment:描述"`
}

func (SshKey) TableName() string {
	return "ssh_keys"
}
