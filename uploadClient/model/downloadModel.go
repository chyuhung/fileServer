package model

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

type DownloadModel struct {
	userId     string //用户id  "780002"
	uploadName string //文件名
	uploadPath string //远端文件路径
	targetUrl  string // Server url
}

func (self *DownloadModel) Init(user string, fileName string, uploadPath string, url string) {
	self.userId = user
	self.uploadName = fileName
	self.uploadPath = uploadPath
	self.targetUrl = url
}

func (self *DownloadModel) Download() error {
	u, _ := url.Parse(self.targetUrl + "/download")
	q := u.Query()
	q.Set("user_id", self.userId)
	q.Set("file_name", self.uploadName)
	q.Set("target_path", self.uploadPath)
	u.RawQuery = q.Encode()
	res, err := http.Get(u.String())
	if err != nil {
		fmt.Println("Download request error")
		return err
	}
	defer res.Body.Close()
	// 保存的文件
	dir, _ := os.Getwd()
	fileDir := filepath.Join(dir, "download")
	os.MkdirAll(fileDir, 0755)
	filePath := filepath.Join(fileDir, self.uploadName)
	out, err := os.Create(filePath)
	if err != nil {
		fmt.Printf("Create %s failed", filePath)
		return err
	}
	defer out.Close()
	io.Copy(out, res.Body)
	if err != nil {
		fmt.Println("Download read body error")
		return err
	}

	return nil
}
