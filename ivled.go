package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	// drop-in replacement of stdlib 'encoding/json' that's way faster
	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigFastest
var ivleroot = os.ExpandEnv("$HOME/nus")

type HomoFolder struct {
	Results []HomoFolder `json:"Results"`

	Title string `json:"Title"`

	FolderName string `json:"FolderName"`
	Folders    []HomoFolder `json:"Folders"`
	Files      []HomoFolder `json:"Files"`

	FileName string `json:"FileName"`
	FileType string `json:"FIleType"`
	FileSize int `json:"FileSize"`
	ID       string `json:"ID"`
}

type PreStruct struct {
	Results []HomoFolder
}

func main() {
	fmt.Println(ivleroot)
	fmt.Println("$LAPIkey:", os.Getenv("LAPIkey"))
	fmt.Println("$AuthToken:", os.Getenv("AuthToken"))

	jsonbytes, _ := ioutil.ReadFile("module_jsons/eg2401.json")
	// cprint(string(jsonbytes))
	var bigfolder []HomoFolder
	json.Unmarshal(jsonbytes, &bigfolder)

	resp, _ := http.Get(os.ExpandEnv("https://ivle.nus.edu.sg/api/Lapi.svc/Workbins?APIKey=$LAPIkey&AuthToken=$AuthToken&CourseID=$MA1513"))
	body, _ := ioutil.ReadAll(resp.Body)
	cprint(string(body))
	fmt.Printf("\n\n\n")
	// var prestruct PreStruct
	var homofolders HomoFolder
	json.Unmarshal(body, &homofolders)
	// cprint(prestruct)
	// fmt.Printf("\n\n\n")
	// homofolders = prestruct.Results
	cprint(homofolders)

	// cprint(bigfolder[0])
	for _, hf := range bigfolder {
		walk(hf)
	}
	// walk(homofolders)

	// fileurl := os.ExpandEnv("https://ivle.nus.edu.sg/api/downloadfile.ashx?APIKey=$LAPIkey&AuthToken=$AuthToken&ID=5444db22-b035-406a-9c46-2cdac6e30bd3&target=workbin")
	// if err := DownloadFile("yee.pdf", fileurl); err != nil {
	// 	panic(err)
	// }
}

func walk(hf HomoFolder) {
	if len(hf.Results) > 0 {
		fmt.Println("Results:")
		for _, hf1 := range hf.Results {
			walk(hf1)
		}
	} else if hf.Title != "" {
		fmt.Println("Workbin:", hf.Title)
		for _, hf1 := range hf.Folders {
			walk(hf1)
		}
	} else if hf.FolderName != "" {
		fmt.Println("FolderName:", hf.FolderName)
		for _, hf1 := range hf.Folders {
			walk(hf1)
		}
		for _, hf1 := range hf.Files {
			walk(hf1)
		}
	} else if hf.FileName != "" {
		fmt.Println(hf.FileName)
		if !fileExists(hf.FileName) {
		}
		fmt.Println(hf.ID)
	}
}

func fileExists(filepath string) bool {
	return false
}

func DownloadFile(filepath string, url string) error {
	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	cprint(url)
	cprint(resp.Body)
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

func tprint(v interface{}) {
	fmt.Printf("%s\n", fmt.Sprintf("%T", v))
}

func cprint(v interface{}) {
	fmt.Printf("%#v\n", v)
}
