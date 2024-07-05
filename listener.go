package main

import (
	"github.com/CuteReimu/bilibili/v2"
	"os"
	"strings"
)

// BiliListener 监听模块
type BiliListener interface {
	InitListener()
	GetTargetFavourFolders()
	LoginByQRCode() bool
	GetFavourFolderContents(folderId int) FavourContents
}

func (b *BiliListenerStruct) LoginByQRCode() bool {
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
			b.logger.Errorf("创建 ./configs 文件夹失败！报错：%v", err)
			return false
		}
		b.logger.Debug("创建./configs文件夹成功！")
	}

	err = os.WriteFile("./configs/cookies.txt", []byte(cookiesString), os.ModePerm)
	if err != nil {
		b.logger.Errorf("写入文件失败：%v", err)
		return false
	}
	b.logger.Debug("写入cookies成功！")
	return true
}

// InitListener 初始化监听器
func (b *BiliListenerStruct) InitListener() {
	b.BiliClient.Resty().SetLogger(b.logger)
	b.logger.Info("开始初始化收藏夹监听模块")
	b.logger.Debug("尝试获取./configs/cookies.txt以登录b站")
	cookies, err := os.ReadFile("./configs/cookies.txt")
	if err != nil {
		b.logger.Info("读取cookie文件失败，开始执行扫码登录流程！")
		res := b.LoginByQRCode()
		if !res {
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
			res := b.LoginByQRCode()
			if !res {
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
	b.logger.Debugf("%+v", b.FavourFolderList)
	b.logger.Info("初始化完成！开始启动监听！")
	b.GetFavourFolderContents((*b.FavourFolderList)[0].Id)
}

// GetTargetFavourFolders 获取用户的收藏夹列表，并根据规则剔除掉一些（收藏夹标题包含[nf]，即no follow标记）
func (b *BiliListenerStruct) GetTargetFavourFolders() {
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
		if !strings.Contains(folder.Title, "[nf]") {
			newFolderList = append(newFolderList, folder)
		} else {
			b.logger.Debugf("在「%v」中检测到[nf]标签，停止监听此文件夹！", folder.Title)
		}
	}
	b.logger.Debugf("筛选后剩余的收藏夹有%v个", len(newFolderList))
	b.FavourFolderList = &newFolderList
}

// GetFavourFolderContents 获取某个收藏夹内的收藏内容
func (b *BiliListenerStruct) GetFavourFolderContents(folderId int) FavourContents {
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

// InitLocalFolder 初始化本地对应的收藏文件夹，可以选择是否全量同步一次bilibili端内容
func (b *BiliListenerStruct) InitLocalFolder(targetDir string, folderId int, doFullSync bool) bool {
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		err = os.MkdirAll(targetDir, os.ModePerm)
		if err != nil {
			b.logger.Warnf("在创建文件夹「%v」时出现错误：%v", targetDir, err)
			return false
		}
	}
	favours := b.GetFavourFolderContents(folderId)
	if favours == nil {
		b.logger.Debugf("favours是nil，但上层函数已经打印过log了，这里直接跳过处理返回false")
		return false
	}
	return true
}
