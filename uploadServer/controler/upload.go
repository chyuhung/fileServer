package controler

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"uploadServer/arsHash"
	"uploadServer/global"

	"github.com/gin-gonic/gin"
)

var (
	myPath = global.InitPath()
)

type fileInfo struct {
	FileName string `json:"fileName"`
	FileSize int64  `json:"fileSize"`
	FileHash string `json:"fileHash"`
}

func getFileInfo(filename string) (fileSize int64, fileHash string, err error) {
	return arsHash.FileHash(filename)
}

func getFileSize(filename string) int64 {
	fileInfo, err := os.Stat(filename)
	if err != nil {
		panic(err)
	}
	return fileInfo.Size()
}
func pathExists(path string) bool {
	_, err := os.Stat(path) //os.Stat获取文件信息
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

func createFilePath(path string) error {
	if !pathExists(path) {
		err := os.MkdirAll(path, os.ModePerm)
		return err
	}
	return nil
}

func uploadPathToLocalPath(user, uploadPath string) string {
	if user == "public" {
		return filepath.Join(myPath.UserDataPath, "Common", uploadPath) + "/"
	}
	return filepath.Join(myPath.UserDataPath, "User", user, uploadPath) + "/"
}

// 名称重复时，获取最新的名称
func getFileNameFormRepeatName(filePath, fileName string) (bool, string, error) {
	i := 0
	var file string
	fileSuffix := path.Ext(fileName)
	filenameOnly := strings.TrimSuffix(fileName, fileSuffix)
	for {
		if i == 0 {
			file = fileName
		} else {
			file = fmt.Sprintf("%s(%d)%s", filenameOnly, i, fileSuffix)
		}
		isExis := pathExists(filePath + file)
		if !isExis {
			break
		}
		i++
	}
	return i > 0, file, nil
}

func getFileInfoListFormRepeatName(filePath, fileName string) ([]fileInfo, string) {
	i := 0
	var fileList []fileInfo
	var file string
	fileSuffix := path.Ext(fileName)
	filenameOnly := strings.TrimSuffix(fileName, fileSuffix)
	for {
		if i == 0 {
			file = fileName
		} else {
			file = fmt.Sprintf("%s(%d)%s", filenameOnly, i, fileSuffix)
		}

		fileSize, fileHash, err := arsHash.FileHash(filePath + file)
		if err != nil {
			break
		}
		fileMode := fileInfo{file, fileSize, fileHash}
		fileList = append(fileList, fileMode)
		i++
	}
	return fileList, file
}

func GetProgress(c *gin.Context) {
	user := c.Query("user_id")
	fileName := c.Query("file_name")
	filePath := c.Query("target_path")
	fileHash := c.Query("task_hash") //可以用uuid
	fileTmp := myPath.TempDataPath + user + "/" + fileHash
	if fileHash == "" {
		c.JSON(http.StatusOK, gin.H{"code": -1, "description": "file_hash is null"})
		return
	}
	var progress int64 = 0
	isExists := pathExists(fileTmp)
	if isExists {
		progress = getFileSize(fileTmp)
	}
	localPath := uploadPathToLocalPath(user, filePath)
	repeatFileList, newName := getFileInfoListFormRepeatName(localPath, fileName)
	if len(repeatFileList) > 0 {
		//fileSize,fileHash ,err:= getFileInfo(localPath+fileName)
		//if err != nil{
		//	c.JSON(http.StatusOK, gin.H{"code": -1,"description":"get fileInfo error"})
		//	return
		//}
		c.JSON(http.StatusOK, gin.H{"code": 0, "description": "", "data": gin.H{"progress": progress, "newName": newName, "fileInfoList": repeatFileList}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "description": "", "data": gin.H{"progress": progress}})
}
func GetFile(c *gin.Context) {
	fileName := c.Query("file_name")
	user := c.Query("user_id")
	filePath := c.Query("target_path")
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", "attachment; filename="+fileName)
	c.Header("Content-Transfer-Encoding", "binary")

	localPath := uploadPathToLocalPath(user, filePath)
	c.File(localPath + fileName)
}
func UploadDelete(c *gin.Context) {
	user := c.PostForm("user_id")
	fileHash := c.PostForm("task_hash") //可以用uuid
	fileTmp := myPath.TempDataPath + user + "/" + fileHash
	if fileHash == "" {
		c.JSON(http.StatusOK, gin.H{"code": -1, "description": "file_hash is null"})
		return
	}
	isExists := pathExists(fileTmp)
	if isExists {
		err := os.Remove(fileTmp)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"code": -1, "description": "remove temp file failed"})
		}
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "description": "success"})
}

func AppendHandle(c *gin.Context) {
	user := c.Query("user_id")
	fileName := c.Query("file_name")
	filePath := c.Query("target_path")
	fileHash := c.Query("task_hash") //可以用uuid
	fileSizeStr := c.Query("file_size")
	isCoverStr := c.Query("cover")
	fileSize, err := strconv.ParseInt(fileSizeStr, 10, 64)
	if err != nil || fileSize < 0 {
		c.JSON(http.StatusOK, gin.H{"code": -1, "description": "file_size must be uint64"})
		return
	}
	if fileHash == "" {
		c.JSON(http.StatusOK, gin.H{"code": -1, "description": "file_hash is null"})
		return
	}
	fileTempPath := myPath.TempDataPath + user + "/" + fileHash
	err = createFilePath(myPath.TempDataPath + user)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusOK, gin.H{"code": -1, "description": "creat tmp folder failed"})
		return
	}
	fileTemp, err := os.OpenFile(fileTempPath, os.O_CREATE|os.O_RDWR, 0666)
	defer fileTemp.Close()
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusOK, gin.H{"code": -1, "description": "creat tmp file failed"})
		return
	}
	curSize, _ := fileTemp.Seek(0, 2) //文件句柄跳到最后，并返回偏移量(临时文件大小)
	//fileUpload, err := c.FormFile("file")
	//fileUpload, header, err := c.Request.FormFile("file")   //读取header,可以从range 鉴定临时文件大小是否一致
	fileUpload := c.Request.Body
	defer fileUpload.Close()

	buf := make([]byte, 2<<20)
	for {
		n, err := fileUpload.Read(buf) //网络原因,每次读不一定是1024
		if n > 0 {
			fileTemp.Write(buf[0:n])
		}
		curSize = curSize + int64(n)
		if err == io.EOF { //结束
			fileTemp.Close()
			fmt.Println("finish ?")
			break
		}
	}
	if curSize < fileSize {
		c.JSON(http.StatusOK, gin.H{"code": 0, "description": "incomplete", "data": gin.H{"fileName": fileName, "progress": curSize, "complete": false}})
		return
	}
	if curSize > fileSize {
		c.JSON(http.StatusOK, gin.H{"code": -1, "description": "Temp file is bigger then file size"})
		return
	}
	localPath := uploadPathToLocalPath(user, filePath)
	err = createFilePath(localPath)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusOK, gin.H{"code": -1, "description": "Create User Upload Path failed"})
		return
	}
	if strings.EqualFold(isCoverStr, "1") {
		os.Rename(fileTempPath, localPath+fileName)
		c.JSON(http.StatusOK, gin.H{"code": 0, "description": "success", "data": gin.H{"fileName": fileName, "progress": curSize, "complete": true}})
		return
	}
	_, newFileName, err := getFileNameFormRepeatName(localPath, fileName)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusOK, gin.H{"code": -1, "description": "Get RepeatName failed"})
		return
	}
	os.Rename(fileTempPath, localPath+newFileName)
	c.JSON(http.StatusOK, gin.H{"code": 0, "description": "success", "data": gin.H{"fileName": newFileName, "progress": curSize, "complete": true}})
}

func UploadNewFile(c *gin.Context) {
	user := c.Query("user_id")
	fileName := c.Query("file_name")
	filePath := c.Query("target_path")
	fileHash := c.Query("task_hash") //可以用uuid
	fileSizeStr := c.Query("file_size")
	isCoverStr := c.Query("cover")
	fileSize, err := strconv.ParseInt(fileSizeStr, 10, 64)
	if err != nil || fileSize < 0 {
		c.JSON(http.StatusOK, gin.H{"code": -1, "description": "file_size must be uint64"})
		return
	}
	if fileHash == "" {
		c.JSON(http.StatusOK, gin.H{"code": -1, "description": "file_hash is null"})
		return
	}
	fileTempPath := myPath.TempDataPath + user + "/" + fileHash
	err = createFilePath(myPath.TempDataPath + user)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusOK, gin.H{"code": -1, "description": "creat tmp folder failed"})
		return
	}
	fileTemp, err := os.OpenFile(fileTempPath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666)
	defer fileTemp.Close()
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusOK, gin.H{"code": -1, "description": "creat tmp file failed"})
		return
	}
	curSize := int64(0)
	fileUpload := c.Request.Body

	defer fileUpload.Close()

	buf := make([]byte, 1024)
	for {
		n, err := fileUpload.Read(buf) //网络原因,每次读不一定是1024
		if n > 0 {
			fileTemp.Write(buf[0:n])
		}
		curSize = curSize + int64(n)
		if err == io.EOF { //结束
			fileTemp.Close()
			fmt.Println("finish ?")
			break
		}
	}
	if curSize < fileSize {
		c.JSON(http.StatusOK, gin.H{"code": 0, "description": "incomplete", "data": gin.H{"fileName": fileName, "progress": curSize, "complete": false}})
		return
	}
	if curSize > fileSize {
		c.JSON(http.StatusOK, gin.H{"code": -1, "description": "Temp file is bigger then file size"})
		return
	}
	localPath := uploadPathToLocalPath(user, filePath)
	err = createFilePath(localPath)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusOK, gin.H{"code": -1, "description": "Create User Upload Path failed"})
		return
	}
	if strings.EqualFold(isCoverStr, "1") {
		os.Rename(fileTempPath, localPath+fileName)
		c.JSON(http.StatusOK, gin.H{"code": 0, "description": "success", "data": gin.H{"fileName": fileName, "progress": curSize, "complete": true}})
		return
	}
	_, newFileName, err := getFileNameFormRepeatName(localPath, fileName)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusOK, gin.H{"code": -1, "description": "Get RepeatName failed"})
		return
	}
	os.Rename(fileTempPath, localPath+newFileName)
	c.JSON(http.StatusOK, gin.H{"code": 0, "description": "success", "data": gin.H{"fileName": newFileName, "progress": curSize, "complete": true}})
}

func init() {
	exist := pathExists(myPath.TempDataPath)
	if !exist {
		// 创建文件夹
		err := os.Mkdir(myPath.TempDataPath, os.ModePerm)
		if err != nil {
			fmt.Printf("mkdir failed![%v]\n", err)
			os.Exit(2)
		}
	}
}
