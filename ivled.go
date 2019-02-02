package main

import (
	"fmt"
	"net/http"
	"os"
)

type File struct {
	ID       string
	FileName string
	FileType string
	FileSize int
}

type Folder struct {
	FolderName string
	Folders    []Folder
	Files      []File
}

type Workbin struct {
	Title   string
	Folders []Folder
}

func workbins() {
	resp, err := http.Get(os.ExpandEnv("https://ivle.nus.edu.sg/api/Lapi.svc/Modules_Search?APIKey=$LAPIkey&AuthToken=$AuthToken&IncludeAllInfo=false&ModuleCode=$1"))
	if err != nil {
		// handle err
	}
	defer resp.Body.Close()
}

func main() {
	fmt.Println("hello ivled")
}
