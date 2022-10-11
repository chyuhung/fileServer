package main

import (
	"flag"
	"fmt"
	"uploadClient/model"
)

func main() {
	var (
		user       string // 上传用户名
		filePath   string // 本地文件路径
		uploadPath string // 指定远端文件目录
		targetUrl  string // 服务端URL 测试：http://127.0.0.1:27149/fileServer http://10.191.22.9:8001/27149
		isCover    bool
	)
	//命令行参数
	flag.StringVar(&user, "u", "public", "upload user.")
	flag.StringVar(&filePath, "f", "", "file path.")
	flag.StringVar(&uploadPath, "d", "/*home*", "remote dir path. option: /*home*, /*public*.\neg: -d /*home*/newdir.\n")
	flag.StringVar(&targetUrl, "l", "http://127.0.0.1:27149/fileServer", "server url.")
	flag.BoolVar(&isCover, "o", false, "overwrite")
	flag.Parse()

	uploadModel := model.UploadModel{}
	err := uploadModel.Init(user, filePath, uploadPath)
	if err != nil {
		fmt.Println(err)
		return
	}
	// set url
	uploadModel.SetUrl(targetUrl)
	uploadModel.IsCover = isCover //是否覆盖上传
	//fileHash := arsHash.FileHash(filePath)
	//fmt.Println(fileHash)

	progressInfo, err := uploadModel.GetProgressFromServer()
	if err != nil {
		fmt.Println(err)
		fmt.Println("get Progress failed")
		return
	}
	fmt.Printf("获取上传进度(缓存文件大小): %d\n", progressInfo.Progress)

	if len(progressInfo.FileInfoList) > 0 {
		fmt.Printf("服务端预定义文件名: %s\n", progressInfo.NewName)
		i := 0
		for i < len(progressInfo.FileInfoList) {
			item := progressInfo.FileInfoList[i]
			fmt.Printf("\n远端路径有重名文件\n")
			fmt.Printf("文件名称: %s\n", item.FileName)
			fmt.Printf("文件大小: %d\n", item.FileSize)
			fmt.Printf("文件哈希: %s\n", item.FileHash)
			fmt.Printf("\n")
			i++
		}

	}
	err = uploadModel.UploadStart()
	//err = uploadModel.UploadDelete()
	if err != nil {
		fmt.Println(err)
		fmt.Println("upload failed")
		return
	}
}
