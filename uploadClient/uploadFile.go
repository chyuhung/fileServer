package main

import (
	"fmt"
	"uploadClient/model"
)

func main() {
	myArgs := model.InitArgs()
	// download file
	if myArgs.IsDownload {
		if myArgs.FileName != "" && !myArgs.IsCover {
			downloadModel := model.DownloadModel{}
			downloadModel.Init(myArgs.User, myArgs.FileName, myArgs.UploadPath, myArgs.TargetUrl)
			err := downloadModel.Download()
			if err != nil {
				fmt.Println(err)
				fmt.Println("dowload failed")
			}
		} else {
			fmt.Println("use -d with -n and without -o.")
		}
		return
	}
	// upload file
	uploadModel := model.UploadModel{}
	err := uploadModel.Init(myArgs.User, myArgs.FilePath, myArgs.UploadPath)
	if err != nil {
		fmt.Println(err)
		return
	}
	// set url
	uploadModel.SetUrl(myArgs.TargetUrl)
	uploadModel.IsCover = myArgs.IsCover //是否覆盖上传
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
