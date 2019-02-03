package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// Used to store the user config
type IVLEConfig struct {
	LAPIkey           string
	AuthToken         string
	AuthTokenExpiry   string
	StudentID         string
	AcadYear          string
	Semester          string
	DownloadLocation  string
	ExcludedFileTypes map[string]bool
	ModulesThisSem    []ModuleInfo
}

// All IVLE API responses look like this
type IVLEResponse struct {
	Results  json.RawMessage
	Comments string
}

// Stores all the info on a module. The important thing to note is that the ID
// field must be queried for in a separate API call, because the IVLE API
// doesn't include it when you ask for all the modules that a user is taking.
type ModuleInfo struct {
	ModuleCode      string
	ModuleTitle     string
	AcadYear        string
	SemesterDisplay string
	ID              string
}

// Used to unmarshal the json responses when searching for ModuleID by
// ModuleCode.  Although it greatly resembles ModuleInfo struct, we cannot use
// ModuleInfo because the json fields are named differently (although their
// contents are semantically the same)
type CourseInfo struct {
	CourseCode     string
	CourseName     string
	CourseAcadYear string
	CourseSemester string
	ID             string
}

// Homogenous Folders (not what you were thinking). Homofolders can either be
// folders or files (or workbins). Although folders and files and workbins are
// all different, Go's anal typing means that it will refuse to unmarshal json
// into non-homogenous structs. It must be a homogenous struct. Thus I have
// combined all the json properties of files, folders and workbins into one
// giant struct. It's a hack, but it works
type HomoFolder struct {
	// Results []HomoFolder

	Title string

	FolderName string
	Folders    []HomoFolder
	Files      []HomoFolder

	FileName string
	FileType string
	FileSize int
	ID       string
}

var filetype_exclusionlist = map[string]bool{
	"mp4": true,
	"mp3": true,
	"mov": true,
	"avi": true,
}

var ivleconfig IVLEConfig

func main() {

	// Read in the user config into struct ivleconfig
	// If it doesn't exist we'll have to set it up the first time
	doSetupConfig := true
	if _, err := os.Stat(os.ExpandEnv("$HOME/.config/ivled.json")); err == nil {
		jsonbytes, _ := ioutil.ReadFile(os.ExpandEnv("$HOME/.config/ivled.json"))
		err := json.Unmarshal(jsonbytes, &ivleconfig)
		if err != nil {
			panic(err)
		}
		if len(ivleconfig.ModulesThisSem) >= 0 {
			doSetupConfig = false
		}
	}
	if doSetupConfig {
		ivleconfig = SetupConfig()
	}
	modules := ivleconfig.ModulesThisSem

	// For each module in modules, download the workbin directory structure (a
	// json file) with an API call and parse the json into a recursive
	// []HomoFolder data struct. Then, for each homofolder in homofolders, call IVLEWalk() on it
	for _, module := range modules {

		// Download module's workbin directory structure
		fmt.Println("==================================")
		fmt.Println("Downloading", module.ModuleCode, "Workbin")
		fmt.Println("==================================")
		ivleresponse, _ := IVLEGetRequest("https://ivle.nus.edu.sg/api/Lapi.svc/Workbins?APIKey=" + ivleconfig.LAPIkey + "&AuthToken=" + ivleconfig.AuthToken + "&CourseID=" + module.ID)

		// Parse workbin directory structure
		var homofolders []HomoFolder
		json.Unmarshal(ivleresponse.Results, &homofolders)

		// Recursively iterate over workbin directory structure
		CreateDirIfNotExist(os.ExpandEnv(ivleconfig.DownloadLocation))
		for _, hf := range homofolders {
			IVLEWalk(module.ModuleCode, os.ExpandEnv(ivleconfig.DownloadLocation), hf)
		}
	}
}

// Main function to interact with the IVLE API
func IVLEGetRequest(url string) (ivleresponse IVLEResponse, err error) {
	resp, _ := http.Get(url)
	body, _ := ioutil.ReadAll(resp.Body)
	json.Unmarshal(body, &ivleresponse)
	if ivleresponse.Comments == "Invalid login!" {
		err = errors.New("Invalid login!")
	}
	return ivleresponse, err
}

// Recursive, Depth First Search function that walks through the HomoFolder and
// downloads files that are missing
func IVLEWalk(modulecode string, filedir string, hf HomoFolder) {
	// if len(hf.Results) > 0 {
	// 	for _, hf1 := range hf.Results {
	// 		IVLEWalk(modulecode, filedir, hf1)
	// 	}
	// } else if hf.Title != "" {
	if hf.Title != "" {
		disdir := filedir + "/" + modulecode
		if !strings.Contains(strings.ToLower(hf.FolderName), "submission") {
			fmt.Println("Folder      :", disdir)
			CreateDirIfNotExist(disdir)
			for _, hf1 := range hf.Folders {
				IVLEWalk(modulecode, disdir, hf1)
			}
		}
	} else if hf.FolderName != "" {
		disdir := filedir + "/" + hf.FolderName
		if !strings.Contains(strings.ToLower(hf.FolderName), "submission") {
			fmt.Println("Folder      :", disdir)
			CreateDirIfNotExist(disdir)
			for _, hf1 := range hf.Folders {
				IVLEWalk(modulecode, disdir, hf1)
			}
		}
		for _, hf1 := range hf.Files {
			IVLEWalk(modulecode, disdir, hf1)
		}
	} else if hf.FileName != "" {
		disfile := filedir + "/" + hf.FileName
		if err := DownloadFileIfNotExist(disfile, hf.ID, hf.FileType); err != nil {
		} else {
		}
	}
}

// If ~/.config/ivled.json is missing (e.g. the user is running ivled for the first time), this function will create it
func SetupConfig() IVLEConfig {
	// create config struct
	var ivleconfig IVLEConfig

	// Get StudentID
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("What is your student ID? (e.g. e0031878): ")
	StudentID, _ := reader.ReadString('\n')
	fmt.Println(StudentID)
	ivleconfig.StudentID = strings.Trim(StudentID, " \n")

	// Get LAPIkey
	fmt.Println("A browser should have opened the URL https://ivle.nus.edu.sg/LAPI/default.aspx (if not, open it manually). Copy your LAPI key, paste it back here (Cmd+V for macOS) then press Enter. (If that doesn't work just use my LAPI key wRDGB8jT2IbKNRBrYnd6F)")
	openbrowser("https://ivle.nus.edu.sg/LAPI/default.aspx")
	fmt.Print("LAPI key: ")
	LAPIkey, _ := reader.ReadString('\n')
	ivleconfig.LAPIkey = strings.Trim(LAPIkey, " \n")

	// Get AuthToken
	authtoken_url := "https://ivle.nus.edu.sg/api/login/?apikey=" + LAPIkey
	fmt.Println("\nA browser should have opened the URL " + authtoken_url + " (you will likely see a blank page, if so just visit the url manually). Enter your IVLE credentials, copy the long authorization token (Cmd+A to select all), paste it back here (Cmd+V for macOS) then press Enter")
	openbrowser(authtoken_url) // IVLE disabled it or something, that's why all IVLEDownloaders have stopped 'working'. Not to worry we can tell the user to manually visit the URL.
	fmt.Print("Authorization Token: ")
	AuthToken, _ := reader.ReadString('\n')
	fmt.Println(AuthToken)
	ivleconfig.AuthToken = strings.Trim(AuthToken, " \n")

	//TODO obtain authtoken expiry with https://ivle.nus.edu.sg/api/Lapi.svc/Validate?APIKey={System.String}&Token={System.String}

	// Get AcadYear and Semester
	month, _ := strconv.Atoi(time.Now().Format("1"))
	var Semester, sem1year, sem2year string
	if month >= 8 && month <= 12 {
		Semester = "1"
		sem1year = time.Now().Format("2006")
		sem2year = time.Now().AddDate(1, 0, 0).Format("2006")
	} else if month >= 1 && month <= 7 {
		Semester = "2"
		sem1year = time.Now().AddDate(-1, 0, 0).Format("2006")
		sem2year = time.Now().Format("2006")
	} else {
		log.Fatalf("Error 01: Wtf? Your month falls outside 1-12")
	}
	AcadYear := sem1year + "/" + sem2year
	fmt.Println(AcadYear, Semester)
	ivleconfig.AcadYear = strings.Trim(AcadYear, " \n")
	ivleconfig.Semester = strings.Trim(Semester, " \n")

	// Get DownloadLocation
	fmt.Println("Where would you like to download your IVLE folders to? Leave blank to download them into the current folder, otherwise provide a path like \"~/Documents/ivled_nus\"")
	fmt.Print("ivled Download Location: ")
	DownloadLocation, _ := reader.ReadString('\n')
	if DownloadLocation == "\n" {
		currdir, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
		DownloadLocation = currdir
	}
	DownloadLocation = strings.Replace(DownloadLocation, "~", "$HOME", 1)
	fmt.Println(DownloadLocation)
	ivleconfig.DownloadLocation = strings.Trim(DownloadLocation, " \n")

	// Get ModulesThisSem
	fmt.Println("=====================================")
	fmt.Println("GETting your modules this semester..")
	fmt.Println("=====================================")
	// resp, _ := http.Get(os.ExpandEnv("https://ivle.nus.edu.sg/api/Lapi.svc/Modules_Taken?APIKey=" + ivleconfig.LAPIkey + "&AuthToken=" + ivleconfig.AuthToken + "&StudentID=" + ivleconfig.StudentID))
	// body, _ := ioutil.ReadAll(resp.Body)
	// var ivleresponse IVLEResponse
	// json.Unmarshal(body, &ivleresponse)
	ivleresponse, _ := IVLEGetRequest("https://ivle.nus.edu.sg/api/Lapi.svc/Modules_Taken?APIKey=" + ivleconfig.LAPIkey + "&AuthToken=" + ivleconfig.AuthToken + "&StudentID=" + ivleconfig.StudentID)
	tprint(ivleresponse.Results)
	cprint(string(ivleresponse.Results))
	var moduleinfos []ModuleInfo
	json.Unmarshal(ivleresponse.Results, &moduleinfos)
	cprint(moduleinfos)
	moduleinfos = FilterModuleInfo(moduleinfos, func(mi ModuleInfo) bool {
		return mi.AcadYear == "2018/2019" && mi.SemesterDisplay == "Semester 2"
	})
	moduleinfos = MapModuleInfo(moduleinfos, func(mi ModuleInfo) ModuleInfo {
		fmt.Println("GETting module info for :", mi.ModuleCode)
		ivleresponse, _ := IVLEGetRequest("https://ivle.nus.edu.sg/api/Lapi.svc/Modules_Search?APIKey=" + ivleconfig.LAPIkey + "&AuthToken=" + ivleconfig.AuthToken +"&IncludeAllInfo=false&ModuleCode=" + mi.ModuleCode)
		// resp, _ := http.Get(os.ExpandEnv("https://ivle.nus.edu.sg/api/Lapi.svc/Modules_Search?APIKey=" + ivleconfig.LAPIkey + "&AuthToken=" + ivleconfig.AuthToken +"&IncludeAllInfo=false&ModuleCode=" + mi.ModuleCode))
		// body, _ := ioutil.ReadAll(resp.Body)
		// var ivleresponse IVLEResponse
		// json.Unmarshal(body, &ivleresponse)
		var courseinfos []CourseInfo
		json.Unmarshal(ivleresponse.Results, &courseinfos)
		courseinfos = FilterCourseInfo(courseinfos, func(ci CourseInfo) bool {
			return ci.CourseAcadYear == mi.AcadYear && ci.CourseSemester == mi.SemesterDisplay
		})
		//TODO check that courseinfos[] has at least one element else next line will fail
		mi.ID = courseinfos[0].ID
		return mi
	})
	ivleconfig.ModulesThisSem = moduleinfos

	// Write Data config file for writing to
	CreateDirIfNotExist(os.ExpandEnv("$HOME/.config"))
	configfile, _ := os.OpenFile(os.ExpandEnv("$HOME/.config/ivled.json"), os.O_WRONLY|os.O_CREATE, 0666)
	json, _ := JSONMarshalIndent(ivleconfig, true)
	configfile.Truncate(0)
	configfile.Seek(0, 0)
	configfile.Write(json)
	defer configfile.Close()

	return ivleconfig
}

//===============================//
// HERE LIE THE HELPER FUNCTIONS //
//===============================//

func DownloadFileIfNotExist(filepath string, fileid string, filetype string) error {
	if filetype_exclusionlist[strings.ToLower(filetype)] {
		return nil
	}
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		// Get the data
		fmt.Println("Downloading :", filepath)
		// url := "https://ivle.nus.edu.sg/api/downloadfile.ashx?APIKey=" + ivleconfig.LAPIkey + "&AuthToken=" + ivleconfig.AuthToken + "&ID=" + fileid + "&target=workbin"
		resp, err := http.Get("https://ivle.nus.edu.sg/api/downloadfile.ashx?APIKey=" + ivleconfig.LAPIkey + "&AuthToken=" + ivleconfig.AuthToken + "&ID=" + fileid + "&target=workbin")
		if err != nil {
			return err
		}
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

func openbrowser(url string) {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Fatal(err)
	}
}

func tprint(v interface{}) {
	fmt.Printf("%s\n", fmt.Sprintf("%T", v))
}

func cprint(v interface{}) {
	fmt.Printf("%#v\n", v)
}
