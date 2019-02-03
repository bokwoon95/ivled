package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
)

var ivleroot = os.ExpandEnv("$HOME/nus_ivled")

type HomoFolder struct {
	Results []HomoFolder

	Title string

	FolderName string
	Folders    []HomoFolder
	Files      []HomoFolder

	FileName string
	FileType string
	FileSize int
	ID       string
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
	// cprint(body)
	// fmt.Printf("\n\n\n")

	var homofolders HomoFolder
	fmt.Println("unmarshaling json...")
	json.Unmarshal(body, &homofolders)
	fmt.Println("json unmarshaled")
	// cprint(homofolders.Results)
	// fmt.Printf("\n\n\n")

	CreateDirIfNotExist(ivleroot)
	Walk(ivleroot, homofolders)

	// if err := DownloadFile("yee.pdf", fileurl); err != nil {
	// 	panic(err)
	// }
}

func Walk(filedir string, hf HomoFolder) {
	if len(hf.Results) > 0 {
		fmt.Println("Results:")
		for _, hf1 := range hf.Results {
			Walk(filedir, hf1)
		}
	} else if hf.Title != "" {
		disdir := filedir + "/" + hf.Title
		fmt.Println("dir:", disdir)
		CreateDirIfNotExist(disdir)
		for _, hf1 := range hf.Folders {
			Walk(disdir, hf1)
		}
	} else if hf.FolderName != "" {
		disdir := filedir + "/" + hf.FolderName
		fmt.Println("dir:", disdir)
		CreateDirIfNotExist(disdir)
		for _, hf1 := range hf.Folders {
			Walk(disdir, hf1)
		}
		for _, hf1 := range hf.Files {
			Walk(disdir, hf1)
		}
	} else if hf.FileName != "" {
		disfile := filedir + "/" + hf.FileName
		fmt.Println("fil:", disfile)
		fmt.Println("\t", hf.ID)
		if err := DownloadFileIfNotExist(disfile, hf.ID); err != nil {
		} else {
		}
	}
}

func FileExists(filepath string) bool {
	return false
}

func DownloadFileIfNotExist(filepath string, fileid string) error {
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		fmt.Println(filepath, "does not exist, GET-ting..")
		// Get the data
		url := os.ExpandEnv("https://ivle.nus.edu.sg/api/downloadfile.ashx?APIKey=$LAPIkey&AuthToken=$AuthToken&ID=" + fileid + "&target=workbin")
		resp, err := http.Get(url)
		if err != nil {
			return err
		}
		// cprint(url)
		// cprint(resp.Body)
		defer resp.Body.Close()
		fmt.Println(filepath, "GET completed")

		// Create the file
		fmt.Println("creating", filepath + "..")
		out, err := os.Create(filepath)
		if err != nil {
			return err
		}
		defer out.Close()
		fmt.Println(filepath, "created")

		// Write the body to file
		fmt.Println("Writing GET response to", filepath + "..")
		_, err = io.Copy(out, resp.Body)
		fmt.Println("Write completed")
		return err
	}
	return nil
}

func CreateDirIfNotExist(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			panic(err)
		}
	}
}

func tprint(v interface{}) {
	fmt.Printf("%s\n", fmt.Sprintf("%T", v))
}

func cprint(v interface{}) {
	fmt.Printf("%#v\n", v)
}
