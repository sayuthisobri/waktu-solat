package services

import "github.com/sayuthisobri/waktu-solat/common"

type UserConfig struct {
	ID    string
	Value string
}

func GetUserConfig(ctx *common.Ctx, key string, fallback string) string {
	db, _ := OpenDb(ctx)
	uc := UserConfig{
		ID: key,
	}
	if tx := db.First(&uc); tx.Error != nil && len(uc.Value) == 0 {
		return uc.Value
	}
	return fallback
}

func SetUserConfig(ctx *common.Ctx, key string, value string) {
	db, _ := OpenDb(ctx)
	uc := UserConfig{
		ID:    key,
		Value: value,
	}
	db.Save(&uc)
}
