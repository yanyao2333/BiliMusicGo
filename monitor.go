package main

import (
	"fmt"
	"github.com/CuteReimu/bilibili/v2"
	"os"
	"strings"
)

// BiliMonitor 监听模块
type BiliMonitor interface {
	InitListener()
	GetTargetFavourFolders()
	LoginByQRCode() bool
	GetFavourFolderContents(folderId int) FavourContents
}

// LoginByQRCode 扫码登录
func (b *BiliMonitorStruct) LoginByQRCode() error {
	qrCode, _ := b.BiliClient.GetQRCode()
	qrCode.Print()
	result, err := b.BiliClient.LoginWithQRCode(bilibili.LoginWithQRCodeParam{
		QrcodeKey: qrCode.QrcodeKey,
	})
	if err == nil && result.Code == 0 {
		b.logger.Info("登录成功")
	}

	b.logger.Debugf("resty cookies: %+v", b.BiliClient.Resty().Cookies)

	cookiesString := b.BiliClient.GetCookiesString()
	if _, err := os.Stat("./configs"); os.IsNotExist(err) {
		err := os.MkdirAll("./configs", os.ModePerm)
		if err != nil {
			return fmt.Errorf("创建 ./configs 文件夹失败！报错：%w", err)
		}
		b.logger.Debug("创建./configs文件夹成功！")
	}

	err = os.WriteFile("./configs/cookies.txt", []byte(cookiesString), os.ModePerm)
	if err != nil {
		return fmt.Errorf("写入文件失败：%w", err)
	}
	b.logger.Debug("写入cookies成功！")
	return nil
}

// InitListener 初始化监听器
func (b *BiliMonitorStruct) InitListener() {
	b.BiliClient.Resty().SetLogger(b.logger)
	b.logger.Info("开始初始化收藏夹监听模块")
	b.logger.Debug("尝试获取./configs/cookies.txt以登录b站")
	cookies, err := os.ReadFile("./configs/cookies.txt")
	if err != nil {
		b.logger.Info("读取cookie文件失败，开始执行扫码登录流程！")
		err := b.LoginByQRCode()
		if err != nil {
			b.logger.Panic("登录时出现了问题！程序退出！")
		}
	} else {
		b.BiliClient.SetCookiesString(string(cookies))
	}
	b.logger.Info("成功获取到cookie！")
	info, err := b.BiliClient.GetAccountInformation()
	if err != nil {
		if strings.Contains(err.Error(), "-101") {
			b.logger.Info("登录失效，重新开始扫码登录流程！")
			err := b.LoginByQRCode()
			if err != nil {
				b.logger.Panic("登录时出现了问题！程序退出！")
			}
			info, err = b.BiliClient.GetAccountInformation()
			if err != nil {
				b.logger.Panicf("我治不了你了！%v", err)
			}
			b.UserInfo = info
		} else {
			b.logger.Panicf("获取用户信息时出现错误：%v，程序退出！", err)
		}
	}
	b.UserInfo = info
	b.logger.Info(b.UserInfo.Uname, "，欢迎回来！")
	b.logger.Info("开始查询并筛选需要监听的收藏夹")
	b.GetTargetFavourFolders()
	b.logger.Info("初始化完成！开始启动监听！")
	b.GetFavourFolderContents((*b.FavourFolderList)[0].Id)
}

// GetTargetFavourFolders 获取用户的收藏夹列表，只监听收藏夹包含[em]的(意为enable monitor)
func (b *BiliMonitorStruct) GetTargetFavourFolders() {
	info, err := b.BiliClient.GetAllFavourFolderInfo(bilibili.GetAllFavourFolderInfoParam{
		UpMid: b.UserInfo.Mid,
	})
	if err != nil {
		b.logger.Panicf("查询收藏列表时发生错误：%v，程序退出！", err)
	}
	var newFolderList []struct {
		Id         int    `json:"id"`
		Fid        int    `json:"fid"`
		Mid        int    `json:"mid"`
		Attr       int    `json:"attr"`
		Title      string `json:"title"`
		FavState   int    `json:"fav_state"`
		MediaCount int    `json:"media_count"`
	}
	for _, folder := range info.List {
		if strings.Contains(folder.Title, "[em]") {
			b.logger.Debugf("在「%v」中检测到[em]标签，监听此文件夹", folder.Title)
			newFolderList = append(newFolderList, folder)
		}
	}
	b.logger.Debugf("筛选后剩余的收藏夹有%v个", len(newFolderList))
	b.FavourFolderList = &newFolderList
}

// GetFavourFolderContents 获取某个收藏夹内的收藏内容
func (b *BiliMonitorStruct) GetFavourFolderContents(folderId int) FavourContents {
	res, err := b.BiliClient.GetFavourIds(bilibili.GetFavourIdsParam{
		MediaId: folderId,
	})
	if err != nil {
		b.logger.Warnf("在获取收藏夹 %v 的收藏内容时出现错误：%v", folderId, err)
		return nil
	}
	var favours FavourContents
	for _, favourId := range res {
		favours = append(favours, FavourContent{
			Type: favourId.Type,
			Id:   favourId.Id,
			Bvid: favourId.Bvid,
		})
	}
	b.logger.Debugf("处理后的 %v 收藏夹的内容为：%v", folderId, favours)
	return favours
}

// InitLocalFolder 初始化本地对应的收藏文件夹
func (b *BiliMonitorStruct) InitLocalFolder(targetDir string, folderId int) bool {
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		err = os.MkdirAll(targetDir, os.ModePerm)
		if err != nil {
			b.logger.Warnf("在创建文件夹「%v」时出现错误：%v", targetDir, err)
			return false
		}
	}
	favours := b.GetFavourFolderContents(folderId)
	if favours == nil {
		return false
	}
	return true
}

func (b *BiliMonitorStruct) StartMonitor() {
	b.logger.Debug("启动监听")

}

func calculateDiff(local []LocalFavourContent, remote []FavourContent) (toDelete []LocalFavourContent, toDownload []FavourContent) {
	remoteMap := make(map[int]FavourContent)
	for _, file := range remote {
		remoteMap[file.Id] = file
	}

	for _, localFile := range local {
		if _, exists := remoteMap[localFile.Id]; !exists {
			toDelete = append(toDelete, localFile)
		} else {
			delete(remoteMap, localFile.Id)
		}
	}

	for _, file := range remoteMap {
		toDownload = append(toDownload, file)
	}

	return toDelete, toDownload
}

func (b *BiliMonitorStruct) DeleteLocalFavour(content LocalFavourContent) error {
	err := b.DB.DeleteLocalFavour(content)
	if err != nil {
		return fmt.Errorf("在数据库中删除 id: %v 时出错: %w", content.Id, err)
	}
	err = os.Remove(content.FilePath)
	if err != nil {
		return fmt.Errorf("删除文件: %v 时出错: %w", content.FilePath, err)
	}
	return nil
}

func (b *BiliMonitorStruct) GetDownloadUrl(content FavourContent) (url string, err error) {
	return "", err
}

func (b *BiliMonitorStruct) AddLocalFavour(content FavourContent) error {
	err := b.DB.AddLocalFavour(content)
	if err != nil {
		return fmt.Errorf("向数据库中添加新的内容失败: %w", err)
	}
	url, err := b.GetDownloadUrl(content)
	if err != nil {
		return fmt.Errorf("获取 %v 对应的下载链接失败: %s", content.Id, err)
	}
	err = b.Downloader.CreateDownloadTask(url)
	if err != nil {
		return fmt.Errorf("创建 %v 的下载任务失败: %s", content.Id, err)
	}
	return nil
}

func (b *BiliMonitorStruct) SyncOneFavourFolder(folderId int, targetDir string) error {
	remoteContents := b.GetFavourFolderContents(folderId)
	localFavourContents, err := b.DB.GetLocalFolderContents(folderId)
	if err != nil {
		return fmt.Errorf("获取数据库中已有内容失败: %s", err)
	}
	toDelete, toDownload := calculateDiff(localFavourContents, remoteContents)
	for _, content := range toDelete {
		err := b.DeleteLocalFavour(content)
		if err != nil {
			b.logger.Errorf("[sync]在同步删除本地内容时出现错误: %s", err)
		}
	}
	for _, content := range toDownload {
		err = b.AddLocalFavour(content)
		if err != nil {
			b.logger.Errorf("[sync]在同步下载本地内容时出现错误: %s", err)
		}
	}
	return nil
}
