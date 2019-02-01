package main

import "fmt"

type File struct {
	ID string
	FileName string
	FileType string
	FileSize int
}

type Folder struct {
	FolderName string
	Folders []Folder
	Files []File
}

type Workbin struct {
	Title string
	Folders []Folder
}

func main() {
	fmt.Println("hello ivled")
}
