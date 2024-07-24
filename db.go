package main

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db, err = gorm.Open(sqlite.Open("./config/data.db"), &gorm.Config{})

type FavourFolder struct {
	gorm.Model
}

type DB interface {
	GetLocalFolderContents(folderId int) ([]LocalFavourContent, error)
	DeleteLocalFavour(content LocalFavourContent) error
	AddLocalFavour(content FavourContent) error
}

// GetLocalFolderContents 获取本地对应的文件夹中已有的内容
func (s *BiliMusicDBStruct) GetLocalFolderContents(folderId int) ([]LocalFavourContent, error) {
	return nil, nil
}
