/*
Package db fdn cfg
Copyright © 2022 hobbymarks ihobbymarks@gmail.com
*/
package db

import (
	"github.com/hobbymarks/fdn/utils"
	"gorm.io/gorm"
)

const (
	BuiltinSource = "builtin"
	UserSource    = "user"
)

type TermWord struct {
	gorm.Model
	KeyHash       string `gorm:"unique"`
	OriginalLower string
	TargetWord    string
	Source        string `gorm:"default:user"`
}

type ToSepWord struct {
	gorm.Model
	KeyHash string `gorm:"unique"`
	Value   string
	Source  string `gorm:"default:user"`
}

type Separator struct {
	gorm.Model
	KeyHash string `gorm:"unique"`
	Value   string
	Source  string `gorm:"default:user"`
}

func ConnectCFGDB(path ...string) (*gorm.DB, error) {
	conn := GetDB()
	if conn != nil {
		return conn, nil
	}
	return ConnectDB(path...)
}

func RekeySeparator(conn *gorm.DB) error {
	var sep Separator
	if err := conn.First(&sep).Error; err != nil {
		return err
	}
	if sep.Source != BuiltinSource {
		sep.KeyHash = utils.KeyHash(sep.Value)
		return conn.Save(&sep).Error
	}
	return nil
}
