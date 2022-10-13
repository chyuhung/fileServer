package model

import (
	"flag"
	"sync"
)

var (
	myArgs    *Args
	pathMutex sync.Mutex
)

type Args struct {
	User       string // 上传用户名
	FilePath   string // 本地文件路径
	UploadPath string // 指定远端文件目录
	TargetUrl  string // 服务端URL 测试：http://127.0.0.1:27149/fileServer http://10.191.22.9:8001/27149/fileServer
	IsCover    bool   // 覆盖文件
	IsDownload bool   // 下载文件
	FileName   string //文件名
}

func InitArgs() *Args {
	if myArgs != nil {
		return myArgs
	}
	pathMutex.Lock()
	defer pathMutex.Unlock()
	// check again
	if myArgs != nil {
		return myArgs
	}
	myArgs = &Args{}
	//命令行参数
	flag.StringVar(&myArgs.User, "u", "public", "user.")
	flag.StringVar(&myArgs.FilePath, "f", "", "upload file path.")
	flag.StringVar(&myArgs.UploadPath, "r", "", "upload remote dir path. e.g. -r /newdir")
	flag.StringVar(&myArgs.TargetUrl, "l", "http://10.191.22.9:8001/27149/fileServer", "server url.")
	flag.StringVar(&myArgs.FileName, "n", "", "download remote filename.")
	flag.BoolVar(&myArgs.IsCover, "o", false, "upload overwrite.")
	flag.BoolVar(&myArgs.IsDownload, "d", false, "download.")
	flag.Parse()

	return myArgs
}
