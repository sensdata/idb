package model

import "time"

type Fingerprint struct {
	ID           uint      `gorm:"primarykey;AUTO_INCREMENT" json:"id"`
	CreatedAt    time.Time `gorm:"type:timestamp;not null;default:current_timestamp" json:"created_at"`
	UpdatedAt    time.Time `gorm:"type:timestamp;not null;default:current_timestamp" json:"updated_at"`
	IP           string    `gorm:"type:varchar(16);not null" json:"ip"`
	MAC          string    `gorm:"type:varchar(20);not null" json:"mac"`
	HasPublicIP  bool      `gorm:"type:bool;not null;default:false" json:"has_public_ip"`
	Fingerprint  string    `gorm:"type:varchar(64);not null" json:"fingerprint"`
	VerifyResult int32     `gorm:"type:int;not null;default:0" json:"verify_result"`
	VerifyTime   time.Time `gorm:"type:timestamp;default:null" json:"verify_time"`
	ExpireTime   time.Time `gorm:"type:timestamp;default:null" json:"expire_time"`
}

type VerifyRequest struct {
	Fingerprint string `json:"fingerprint"`
	IP          string `json:"ip"`
	MAC         string `json:"mac"`
}

type VerifyResponse struct {
	Fingerprint string `json:"fingerprint"`
	Result      int32  `json:"result"`
	VerifyTime  int64  `json:"verify_time"`
	ExpireTime  int64  `json:"expire_time"`
}
