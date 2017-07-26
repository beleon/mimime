package mimime

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
)

var (
	homePath      string
	cachePath     string
	cacheOrigPath string
)

var fileLocks map[string]*sync.Mutex
var fileLocksLock sync.Mutex

func init() {
	homePath = os.Getenv("HOME")
	cachePath = filepath.Join(homePath, ".cache", applicationName)
	cacheOrigPath = filepath.Join(cachePath, "orig")
	fileLocks = make(map[string]*sync.Mutex)

	err := os.MkdirAll(homePath, os.ModePerm)
	if err != nil {
		fmt.Println(err)
	}
	err = os.MkdirAll(cachePath, os.ModePerm)
	if err != nil {
		fmt.Println(err)
	}
	err = os.MkdirAll(cacheOrigPath, os.ModePerm)
	if err != nil {
		fmt.Println(err)
	}
}

func retrieveOriginal(req *request) error {
	path := filepath.Join(cacheOrigPath, req.imgId())
	lockFile(req.imgId())
	defer unlockFile(req.imgId())

	if !req.reqOpts.setOpts[forceReloadOption] {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			return nil
		} else if err != nil {
			return err
		}
	}

	return downloadOriginal(req, path)
}

//todo: add cleanup routine when sth goes wrong
func downloadOriginal(req *request, path string) error {
	protocol := "http"
	if req.reqOpts.setOpts[sslOption] {
		protocol = "https"
	}

	response, err := http.Get(protocol + "://" + req.imgUrl)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, response.Body)
	if err != nil {
		return err
	}

	return nil
}

func lockFile(key string) {
	fileLocksLock.Lock()
	lock, ok := fileLocks[key]
	if !ok {
		lock = &sync.Mutex{}
		fileLocks[key] = lock
	}
	lock.Lock()
	fileLocksLock.Unlock()
}

func unlockFile(key string) {
	fileLocksLock.Lock()
	fileLocks[key].Unlock()
	fileLocksLock.Unlock()
}
