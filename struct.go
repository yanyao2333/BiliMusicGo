package main

import (
	"github.com/CuteReimu/bilibili/v2"
	"go.uber.org/zap"
)

type FavourContent struct {
	Type int    // 内容类型，2：视频稿件，12：音频，21：视频合集
	Id   int    // 内容id，视频稿件：视频稿件avid，音频：音频auid，视频合集：视频合集id
	Bvid string // 视频才有的bv号
}

type LocalFavourContent struct {
	FavourContent
	FilePath       string
	FavourFolderId int // 所在远端收藏夹的id
}

type BiliMonitorStruct struct {
	BiliClient       *bilibili.Client
	logger           *zap.SugaredLogger
	UserInfo         *bilibili.AccountInformation
	FavourFolderList *[]struct {
		Id         int    `json:"id"`
		Fid        int    `json:"fid"`
		Mid        int    `json:"mid"`
		Attr       int    `json:"attr"`
		Title      string `json:"title"`
		FavState   int    `json:"fav_state"`
		MediaCount int    `json:"media_count"`
	}
	DB         DB
	Downloader Downloader
}

type FavourContents []FavourContent

type BiliMusicDBStruct struct {
}
