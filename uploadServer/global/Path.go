package global

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

var (
	myPath    *Path
	pathMutex sync.Mutex
)

type Path struct {
	UserDataPath string
	TempDataPath string
}

func GetPath() *Path {
	if myPath != nil {
		return myPath
	}
	pathMutex.Lock()
	defer pathMutex.Unlock()
	// check again
	if myPath != nil {
		return myPath
	}
	myPath = &Path{}
	dirPath, err := os.Getwd()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	myPath.UserDataPath = dirPath
	myPath.TempDataPath = filepath.Join(myPath.UserDataPath, "temp")

	fmt.Println("UserDataPath:" + myPath.UserDataPath)
	fmt.Println("TempDataPath:" + myPath.TempDataPath)
	return myPath
}
