package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
)

var ivleroot = os.ExpandEnv("$HOME/nus_ivled")

type IVLEresponse struct {
	Results  json.RawMessage
	Comments string
}

type ModuleInfo struct {
	ModuleCode      string
	ModuleTitle     string
	AcadYear        string
	SemesterDisplay string
	ID              string
}

type CourseInfo struct {
	CourseCode     string
	CourseName     string
	CourseAcadYear string
	CourseSemester string
	ID             string
}

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
	fmt.Println("$StudentID:", os.Getenv("StudentID"))

	moduleinfos := GetModulesTaken("modtak.json")
	cprint(moduleinfos)

	for _, module := range moduleinfos {
		DownloadWorkbin(module.ModuleCode, module.ID)
	}
}

func GetModulesTaken(filename string) (moduleinfos []ModuleInfo) {
	LAPIrequestmodules := true

	if _, err := os.Stat(filename); err == nil {
		jsonbytes, _ := ioutil.ReadFile(filename)
		err := json.Unmarshal(jsonbytes, &moduleinfos)
		if err != nil {
			panic(err)
		}
		if len(moduleinfos) >= 0 {
			LAPIrequestmodules = false
		}
	}

	if LAPIrequestmodules {
		fmt.Println("GET-ting your modules this semester..")
		//TODO Check if LAPIkey and AuthToken and StudentID are not empty
		resp, _ := http.Get(os.ExpandEnv("https://ivle.nus.edu.sg/api/Lapi.svc/Modules_Taken?APIKey=$LAPIkey&AuthToken=$AuthToken&StudentID=$StudentID"))
		fmt.Println("GET completed")
		body, _ := ioutil.ReadAll(resp.Body)
		var ivleresponse IVLEresponse
		json.Unmarshal(body, &ivleresponse)
		tprint(ivleresponse.Results)
		json.Unmarshal(ivleresponse.Results, &moduleinfos)
		moduleinfos = FilterModuleInfo(moduleinfos, func(mi ModuleInfo) bool {
			return mi.AcadYear == "2018/2019" && mi.SemesterDisplay == "Semester 2"
		})
		moduleinfos = MapModuleInfo(moduleinfos, func(mi ModuleInfo) ModuleInfo {
			fmt.Println("GET-ting the module ID of", mi.ModuleCode+"..")
			resp, _ := http.Get(os.ExpandEnv("https://ivle.nus.edu.sg/api/Lapi.svc/Modules_Search?APIKey=$LAPIkey&AuthToken=$AuthToken&IncludeAllInfo=false&ModuleCode=" + mi.ModuleCode))
			fmt.Println("GET completed")
			body, _ := ioutil.ReadAll(resp.Body)
			var ivleresponse IVLEresponse
			json.Unmarshal(body, &ivleresponse)
			var courseinfos []CourseInfo
			json.Unmarshal(ivleresponse.Results, &courseinfos)
			courseinfos = FilterCourseInfo(courseinfos, func(ci CourseInfo) bool {
				return ci.CourseAcadYear == mi.AcadYear && ci.CourseSemester == mi.SemesterDisplay
			})
			//TODO check that courseinfos[] has at least one element else next line will fail
			mi.ID = courseinfos[0].ID
			return mi
		})
		json, _ := JSONMarshalIndent(moduleinfos, true)
		_ = ioutil.WriteFile("modtak.json", json, 0666)
	}
	return
}

func DownloadWorkbin(ModuleCode string, ModuleID string) {
	fmt.Println("sending GET..")
	resp, _ := http.Get(os.ExpandEnv("https://ivle.nus.edu.sg/api/Lapi.svc/Workbins?APIKey=$LAPIkey&AuthToken=$AuthToken&CourseID=" + ModuleID))
	fmt.Println(os.ExpandEnv("https://ivle.nus.edu.sg/api/Lapi.svc/Workbins?APIKey=$LAPIkey&AuthToken=$AuthToken&CourseID=" + ModuleID))
	fmt.Println("GET completed")
	fmt.Println("reading response body..")
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(ModuleCode, ModuleID)
	cprint(string(body))
	fmt.Println("read finished")

	var homofolders HomoFolder
	fmt.Println("unmarshaling json...")
	json.Unmarshal(body, &homofolders)
	fmt.Println("json unmarshaled")
	// cprint(homofolders)

	CreateDirIfNotExist(ivleroot)
	Walk(ModuleCode, ivleroot, homofolders)
}

func Walk(modulecode string, filedir string, hf HomoFolder) {
	if len(hf.Results) > 0 {
		fmt.Println("Results:")
		for _, hf1 := range hf.Results {
			Walk(modulecode, filedir, hf1)
		}
	} else if hf.Title != "" {
		disdir := filedir + "/" + modulecode + " " + hf.Title
		fmt.Println("dir:", disdir)
		CreateDirIfNotExist(disdir)
		for _, hf1 := range hf.Folders {
			Walk(modulecode, disdir, hf1)
		}
	} else if hf.FolderName != "" {
		disdir := filedir + "/" + hf.FolderName
		fmt.Println("dir:", disdir)
		CreateDirIfNotExist(disdir)
		for _, hf1 := range hf.Folders {
			Walk(modulecode, disdir, hf1)
		}
		for _, hf1 := range hf.Files {
			Walk(modulecode, disdir, hf1)
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
		fmt.Println("creating", filepath+"..")
		out, err := os.Create(filepath)
		if err != nil {
			return err
		}
		defer out.Close()
		fmt.Println(filepath, "created")

		// Write the body to file
		fmt.Println("Writing GET response to", filepath+"..")
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

func JSONMarshalIndent(v interface{}, safeEncoding bool) ([]byte, error) {
	b, err := json.MarshalIndent(v, "", "  ")

	if safeEncoding {
		b = bytes.Replace(b, []byte("\\u003c"), []byte("<"), -1)
		b = bytes.Replace(b, []byte("\\u003e"), []byte(">"), -1)
		b = bytes.Replace(b, []byte("\\u0026"), []byte("&"), -1)
	}
	return b, err
}

func FilterModuleInfo(ss []ModuleInfo, test func(ModuleInfo) bool) (ret []ModuleInfo) {
	for _, s := range ss {
		if test(s) {
			ret = append(ret, s)
		}
	}
	return
}

func FilterCourseInfo(ss []CourseInfo, test func(CourseInfo) bool) (ret []CourseInfo) {
	for _, s := range ss {
		if test(s) {
			ret = append(ret, s)
		}
	}
	return
}

func MapModuleInfo(ss []ModuleInfo, fn func(ModuleInfo) ModuleInfo) (ret []ModuleInfo) {
	for _, s := range ss {
		ret = append(ret, fn(s))
	}
	return
}

func tprint(v interface{}) {
	fmt.Printf("%s\n", fmt.Sprintf("%T", v))
}

func cprint(v interface{}) {
	fmt.Printf("%#v\n", v)
}
