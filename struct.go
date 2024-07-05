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

type BiliListenerStruct struct {
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
}

type FavourContents []FavourContent

// LocalFavourMusicRecordStruct 本地储存的json文件中单个记录对象（我听不懂我在写什么...）
type LocalFavourMusicRecordStruct struct {
	Id       int    `json:"id"`     // 内容id，视频稿件：视频稿件avid，音频：音频auid，视频合集：视频合集id
	Status   string `json:"status"` // 状态，saved|removed|error (已保存在本地|之前同步过了，但被用户删除/重命名/移动，不会再次同步|我不知道，预留的)
	FilePath string `json:"file_path"`
}
