package model

import (
	"crypto/sha1"
	"encoding/base32"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
)

type UploadModel struct {
	userId      string //用户id  "780002"
	filePath    string //文件路径，包含文件名
	uploadPath  string //上传路径
	uploadName  string //上传文件名
	IsCover     bool   //是否覆盖上传
	fileSize    int64  //文件总大小
	fileSizeStr string //文件总大小 字符串类型
	fileHash    string //文件哈希，由上传路径 + 上传文件名 + 大小  计算得到
	progress    int64  //进度，已经传了多少
	isReady     bool   //如果续传 是否准备好
	targetUrl   string // Server url
}

func (self *UploadModel) Init(userId, filePath, uploadPath string) error {
	self.isReady = false
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return err
	}
	self.userId = userId
	self.filePath = filePath
	self.fileSize = fileInfo.Size()
	self.uploadPath = uploadPath
	self.uploadName = path.Base(filePath)
	self.progress = 0
	self.IsCover = false
	fileSizeStr := strconv.FormatInt(self.fileSize, 10)
	self.fileSizeStr = fileSizeStr
	Sha1Inst := sha1.New()
	Sha1Inst.Write([]byte(fmt.Sprintf("%s-%s-%s", uploadPath, self.uploadName, fileSizeStr)))
	result := Sha1Inst.Sum([]byte(""))
	self.fileHash = base32.StdEncoding.EncodeToString(result) //上传路径 + 上传文件名 + 大小 计算hash 再使用base32编码转字符串
	return nil
}
func (self *UploadModel) SetUrl(url string) {
	self.targetUrl = url
}

func (self *UploadModel) GetProgressFromServer() (info *progressData, err error) {
	u, _ := url.Parse(self.targetUrl + "/getProgress")
	q := u.Query()
	q.Set("user_id", self.userId)
	q.Set("file_name", self.uploadName)
	q.Set("target_path", self.uploadPath)
	q.Set("file_size", self.fileSizeStr)
	q.Set("task_hash", self.fileHash)
	u.RawQuery = q.Encode()
	res, err := http.Get(u.String())
	if err != nil {
		fmt.Println("GetProgress request error")
		return
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println("GetProgress read body error")
		return
	}
	resData := &progressResponse{}
	err = json.Unmarshal(body, resData)
	if err != nil {
		fmt.Println("GetProgress json decode error")
		fmt.Println(string(body))
		return
	}
	if resData.Code >= 0 {
		self.progress = resData.Data.Progress
		info = &resData.Data
	} else {
		err = errors.New(resData.Description)
	}
	return
}

func (self *UploadModel) UploadStart() error {
	fh, err := os.Open(self.filePath)
	if err != nil {
		fmt.Println("Error opening file")
		return err
	}
	writer := Writer{fh, self.progress}
	u, _ := url.Parse(self.targetUrl + "/uploadAppend")
	q := u.Query()
	q.Set("user_id", self.userId)
	q.Set("file_name", self.uploadName)
	q.Set("target_path", self.uploadPath)
	q.Set("file_size", self.fileSizeStr)
	q.Set("task_hash", self.fileHash)
	if self.IsCover { //不覆盖时，不传这个值就可以了
		q.Set("cover", "1")
	}
	u.RawQuery = q.Encode()
	apizUrl := u.String()
	r, w := io.Pipe()
	go writer.doWrite(w)
	resp, err := http.Post(apizUrl, "binary/octet-stream", r)
	defer resp.Body.Close()
	if err != nil {
		fmt.Println("UploadStart request error")
		return err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("UploadStart Read Body error")
		return err
	}
	resData := &uploadResponse{}
	err = json.Unmarshal(body, resData)
	if err != nil {
		fmt.Println("UploadStart json decode error")
		fmt.Println(string(body))
		return err
	}
	if resData.Code >= 0 {
		self.progress = resData.Data.Progress
		fmt.Printf("上传成功\n文件名: %s\n上传了: %d字节\n是否完成: %t \n", resData.Data.FileName, resData.Data.Progress, resData.Data.Complete)
	} else {
		fmt.Println(resData.Description)
	}
	return nil
}

func (self *UploadModel) UploadDelete() error {
	resp, err := http.PostForm(self.targetUrl+"/uploadDelete", url.Values{"user_id": {self.userId}, "task_hash": {self.fileHash}})

	defer resp.Body.Close()
	if err != nil {
		fmt.Println("UploadDelete request error")
		return err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("UploadDelete Read Body error")
		return err
	}
	resData := &uploadResponse{}
	err = json.Unmarshal(body, resData)
	if err != nil {
		fmt.Println("UploadDelete json decode error")
		fmt.Println(string(body))
		return err
	}
	fmt.Println(resData.Description)
	return nil
}
