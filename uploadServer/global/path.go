package global

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

var (
	myPath    *path
	pathMutex sync.Mutex
)

type path struct {
	UserDataPath string
	TempDataPath string
	delimiter    string
}

func InitPath() *path {
	if myPath != nil {
		return myPath
	}
	pathMutex.Lock()
	defer pathMutex.Unlock()
	// check again
	if myPath != nil {
		return myPath
	}
	myPath = &path{}
	dirPath, err := os.Getwd()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	switch runtime.GOOS {
	case "linux":
		myPath.delimiter = "/"
	case "windows":
		myPath.delimiter = "\\"
	default:
		fmt.Println("runtime.GOOS failed, unknown OS")
		os.Exit(1)
	}
	myPath.UserDataPath = dirPath + myPath.delimiter
	myPath.TempDataPath = filepath.Join(myPath.UserDataPath, "temp") + myPath.delimiter

	fmt.Println("UserDataPath:" + myPath.UserDataPath)
	fmt.Println("TempDataPath:" + myPath.TempDataPath)
	return myPath
}
