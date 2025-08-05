package repo

import "github.com/sensdata/idb/core/model"

type FingerprintRepo struct{}

type IFingerprintRepo interface {
	Get() (*model.Fingerprint, error)
	Set(fingerprint *model.Fingerprint) error
	Verify(fingerprint *model.Fingerprint) error
}

func NewFingerprintRepo() IFingerprintRepo {
	return &FingerprintRepo{}
}

func (f *FingerprintRepo) Get() (*model.Fingerprint, error) {
	var fingerprint model.Fingerprint
	db := getDb().Model(&model.Fingerprint{})
	if err := db.First(&fingerprint).Error; err != nil {
		return &fingerprint, err
	}
	return &fingerprint, nil
}

func (f *FingerprintRepo) Set(fingerprint *model.Fingerprint) error {
	db := getDb().Model(&model.Fingerprint{})
	oldFingerprint, err := f.Get()
	if err != nil || oldFingerprint.ID == 0 {
		return db.Create(&fingerprint).Error
	}
	upMap := make(map[string]interface{})
	upMap["ip"] = fingerprint.IP
	upMap["mac"] = fingerprint.MAC
	upMap["has_public_ip"] = fingerprint.HasPublicIP
	upMap["fingerprint"] = fingerprint.Fingerprint
	if err := db.Where("id = ?", oldFingerprint.ID).Updates(upMap).Error; err != nil {
		return err
	}
	return nil
}

func (f *FingerprintRepo) Verify(fingerprint *model.Fingerprint) error {
	db := getDb().Model(&model.Fingerprint{})
	oldFingerprint, err := f.Get()
	if err != nil || oldFingerprint.ID == 0 {
		return db.Create(&fingerprint).Error
	}
	upMap := make(map[string]interface{})
	upMap["verify_result"] = fingerprint.VerifyResult
	upMap["verify_time"] = fingerprint.VerifyTime
	if err := db.Where("id = ?", oldFingerprint.ID).Updates(upMap).Error; err != nil {
		return err
	}
	return nil
}
