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
	License      string    `gorm:"type:text;default:null" json:"license"`
	Signature    string    `gorm:"type:text;default:null" json:"signature"`
	LastVerifyAt time.Time `gorm:"type:timestamp;default:null" json:"last_verify_at"`
}

type LicensePayload struct {
	Fingerprint string    `json:"fingerprint"`
	IP          string    `json:"ip"`
	MAC         string    `json:"mac"`
	LicenseType int32     `json:"license_type"`
	IssuedAt    time.Time `json:"issued_at"`
	ExpireAt    time.Time `json:"expire_at"`
}

type RegisterFingerprintReq struct {
	Ip          string `json:"ip"`
	Mac         string `json:"mac"`
	Fingerprint string `json:"fingerprint"`
}

type RegisterFingerprintRsp struct {
	License   string `json:"license"`
	Signature string `json:"signature"`
}

type VerifyLicenseRequest struct {
	License   string `json:"license" binding:"required"`   // base64 encoded JSON
	Signature string `json:"signature" binding:"required"` // base64 encoded Ed25519 signature
}

type VerifyLicenseResponse struct {
	Valid       bool   `json:"valid"`
	LicenseType int32  `json:"license_type,omitempty"`
	ExpireAt    string `json:"expire_at,omitempty"`
}
