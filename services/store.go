package services

import (
	"github.com/sayuthisobri/waktu-solat/common"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func OpenDb(ctx *common.Ctx) (*gorm.DB, error) {
	loggerMode := logger.Silent
	if ctx.Config.IsDebug && !ctx.Config.IsAlfred() {
		loggerMode = logger.Warn
	}
	db, err := gorm.Open(sqlite.Open(ctx.Config.DbPath), &gorm.Config{
		Logger: logger.Default.LogMode(loggerMode),
	})
	if err != nil {
		return nil, err
	}
	err = db.AutoMigrate(&PrayerDate{}, &UserConfig{})
	return db, err
}
