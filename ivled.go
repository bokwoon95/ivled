package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	// drop-in replacement of stdlib 'encoding/json' that's way faster
	// jsoniter "github.com/json-iterator/go"
	"encoding/json"
)

// var json = jsoniter.ConfigCompatibleWithStandardLibrary
var ivleroot = os.ExpandEnv("$HOME/nus")

type HomoFolder struct {
	Results []HomoFolder `json:"Results,omitempty"`

	Title string `json:"Title,omitempty"`

	FolderName string       `json:"FolderName,omitempty"`
	Folders    []HomoFolder `json:"Folders,omitempty"`
	Files      []HomoFolder `json:"Files,omitempty"`

	FileName string `json:"FileName,omitempty"`
	FileType string `json:"FIleType,omitempty"`
	FileSize int    `json:"FileSize,omitempty"`
	ID       string `json:"ID,omitempty"`
}

type Results struct {
	Results []*json.RawMessage
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

	fmt.Println("sending GET..")
	resp, _ := http.Get(os.ExpandEnv("https://ivle.nus.edu.sg/api/Lapi.svc/Workbins?APIKey=$LAPIkey&AuthToken=$AuthToken&CourseID=$MA1513"))
	fmt.Println("GET completed")
	fmt.Println("reading response body..")
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("read finished")
	cprint(body)
	fmt.Printf("\n\n\n")
	// var prestruct PreStruct

	var homofolders HomoFolder
	fmt.Println("unmarshaling json...")
	json.Unmarshal(body, &homofolders)
	fmt.Println("json unmarshaled")
	cprint(homofolders.Results)
	fmt.Printf("\n\n\n")

	// var results Results
	// json.Unmarshal(body, &results)
	// cprint(results)
	// fmt.Printf("\n\n\n")

	// cprint(prestruct)
	// fmt.Printf("\n\n\n")
	// homofolders = prestruct.Results
	// cprint(homofolders)

	// cprint(bigfolder[0])
	// for _, hf := range bigfolder {
	// 	walk(hf)
	// }
	walk(homofolders)

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
