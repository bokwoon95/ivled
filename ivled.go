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
	ExcludedFilePaths map[string]bool
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

// Semantically identical to ModuleInfo (field-for-field), but we need this
// because the field names are different
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
	Title string

	FolderName string
	Folders    []HomoFolder
	Files      []HomoFolder

	FileName string
	FileType string
	FileSize int
	ID       string
}

// Global Variables
//-----------------
var ivleconfig IVLEConfig
var downloadedfiles []string

// OS specific variables
var fpdlm string
var configfolder string
var configfile string

func main() {
	// Setup the OS variables
	switch runtime.GOOS {
	case "windows":
		configfolder = os.ExpandEnv("%APPDATA%\\ivled\\")
		fpdlm = "\\"
	case "linux":
		configfolder = os.ExpandEnv("$HOME/.config/ivled/")
		fpdlm = "/"
	case "darwin":
		configfolder = os.ExpandEnv("$HOME/.config/ivled/")
		fpdlm = "/"
	default:
		log.Fatalln("unsupported platform")
	}
	CreateDirIfNotExist(configfolder)
	configfile = configfolder + "config.json"

	// Parse the CLI arguments
	if len(os.Args) >= 2 {
		cmd := os.Args[1]
		switch cmd {
		case "config":
			if _, err := os.Stat(configfile); os.IsNotExist(err) {
				SetupConfig()
			}
			openfile(configfile)
			fmt.Println("Opened your config file", configfile)
			fmt.Println("If your operating system cannot figure out what to open json files with, please manually open it with a text editor.")
			os.Exit(0)
		case "reset":
			deletefile(configfile)
			fmt.Println(configfile, "has been deleted")
			os.Exit(0)
		case "help":
			fmt.Println("Available Commands")
			fmt.Println("------------------")
			fmt.Println("ivled        : Downloads your latest IVLE files into a directory based on your config file")
			fmt.Println("               If your config file is absent, it will run you through the configuration process")
			fmt.Println("ivled config : Opens your config file with an external text editor")
			fmt.Println("               If your config file is absent, it will run you through the configuration process")
			fmt.Println("ivled reset  : Deletes your config file")
			fmt.Println("ivled help   : Displays this help")
			os.Exit(0)
		default:
			fmt.Println("Unknown command '" + cmd + "': ignoring")
			os.Exit(0)
		}
	}

	// Read in the user config into struct ivleconfig
	var config_error error
	ivleconfig, config_error = ReadConfig()
	if config_error != nil {
		ivleconfig = SetupConfig()
	}
	modules := ivleconfig.ModulesThisSem

	// This is the for loop that actually downloads all the files. It loops
	// through every module's workbin and calls the recursive IVLEWalk() on it
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
	fmt.Println("==================================")
	fmt.Println("Download Summary")
	fmt.Println("==================================")
	l := len(downloadedfiles)
	if l == 1 {
		fmt.Println("There was 1 downloaded file")
	} else {
		fmt.Println("There were", l, "downloaded files")
	}
	for i, s := range downloadedfiles {
		fmt.Print(i + 1)
		fmt.Println(".", s)
	}
}

// Main function to interact with the IVLE API
func IVLEGetRequest(url string) (ivleresponse IVLEResponse, err error) {
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalln(err)
	}
	json.NewDecoder(resp.Body).Decode(&ivleresponse)
	if ivleresponse.Comments == "Invalid login!" {
		err = errors.New("Invalid login!")
		log.Fatalln(err)
	}
	return ivleresponse, err
}

// Recursive, Depth First Search function that walks through the HomoFolder and
// downloads files that are not present in the computer
func IVLEWalk(modulecode string, filedir string, hf HomoFolder) {
	if hf.Title != "" { // Means this is a Workbin
		disdir := filedir + fpdlm + modulecode
		if !strings.Contains(strings.ToLower(hf.FolderName), "submission") && !ivleconfig.ExcludedFilePaths[disdir] {
			fmt.Println("Folder      :", disdir+"/")
			CreateDirIfNotExist(disdir)
			for _, hf1 := range hf.Folders {
				IVLEWalk(modulecode, disdir, hf1)
			}
		} else {
			fmt.Println("Ignored     :", disdir+"/")
		}
	} else if hf.FolderName != "" { // Means this is a Folder
		disdir := filedir + fpdlm + hf.FolderName
		if !strings.Contains(strings.ToLower(hf.FolderName), "submission") && !ivleconfig.ExcludedFilePaths[disdir] {
			fmt.Println("Folder      :", disdir+"/")
			CreateDirIfNotExist(disdir)
			for _, hf1 := range hf.Folders {
				IVLEWalk(modulecode, disdir, hf1)
			}
		} else {
			fmt.Println("Ignored     :", disdir+"/")
		}
		for _, hf1 := range hf.Files {
			IVLEWalk(modulecode, disdir, hf1)
		}
	} else if hf.FileName != "" { // Means this is a File
		disfile := filedir + fpdlm + hf.FileName
		if !ivleconfig.ExcludedFileTypes[strings.ToLower(hf.FileType)] && !ivleconfig.ExcludedFilePaths[disfile] {
			DownloadFileIfNotExist(disfile, hf.ID)
		} else {
			fmt.Println("Ignored     :", disfile)
		}
	}
}

func ReadConfig() (ivleconfig IVLEConfig, err error) {
	if _, err = os.Stat(configfile); err == nil {
		jsonbytes, err := ioutil.ReadFile(configfile)
		if err != nil {
			return ivleconfig, err
		}
		err = json.Unmarshal(jsonbytes, &ivleconfig)
		if err != nil {
			return ivleconfig, err
		}
		ivleconfig.DownloadLocation = strings.TrimSuffix(ivleconfig.DownloadLocation, fpdlm)
		if len(ivleconfig.ModulesThisSem) <= 0 {
			return ivleconfig, errors.New("config.json may be corrupted, no modules found")
		}
	}
	return ivleconfig, err
}

// Creates the user's config file
func SetupConfig() (ivleconfig IVLEConfig) {
	reader := bufio.NewReader(os.Stdin)

	// Get StudentID
	fmt.Print("What is your student ID? (e.g. e0031878): ")
	StudentID, _ := reader.ReadString('\n')
	StudentID = strings.Trim(StudentID, " \n\t")
	ivleconfig.StudentID = StudentID

	// Get LAPIkey
	fmt.Println("\n↓↓↓↓↓     PLEASE VISIT THE URL      ↓↓↓↓↓")
	fmt.Println("~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~")
	fmt.Println("https://ivle.nus.edu.sg/LAPI/default.aspx")
	fmt.Println("~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~")
	fmt.Println("Copy your LAPI key and paste it here. (If it fails use mine wRDGB8jT2IbKNRBrYnd6F)")
	openbrowser("https://ivle.nus.edu.sg/LAPI/default.aspx")
	fmt.Print("LAPI key: ")
	LAPIkey, _ := reader.ReadString('\n')
	LAPIkey = strings.Trim(LAPIkey, " \n\t")
	ivleconfig.LAPIkey = LAPIkey

	// Get AuthToken
	fmt.Println("\n↓↓↓↓↓↓↓↓↓↓           PLEASE VISIT THE URL            ↓↓↓↓↓↓↓↓↓↓")
	fmt.Println("~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~")
	fmt.Println("https://ivle.nus.edu.sg/api/login/?apikey=" + LAPIkey)
	fmt.Println("~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~")
	fmt.Println("Log in with your IVLE credentials, select all and paste the very long login token here")
	// openbrowser(authtoken_url) // IVLE disabled it or something, that's why all IVLEDownloaders have stopped 'working'. Not to worry we can tell the user to manually visit the URL.
	fmt.Print("Authorization Token: ")
	AuthToken, _ := reader.ReadString('\n')
	AuthToken = strings.Trim(AuthToken, " \n\t")
	ivleconfig.AuthToken = AuthToken

	//TODO obtain authtoken expiry with https://ivle.nus.edu.sg/api/Lapi.svc/Validate?APIKey={System.String}&Token={System.String}

	// Get AcadYear and Semester
	month, _ := strconv.Atoi(time.Now().Format("1"))
	var Semester, sem1year, sem2year string
	if month >= 8 && month <= 12 {
		Semester = "Semester 1"
		sem1year = time.Now().Format("2006")
		sem2year = time.Now().AddDate(1, 0, 0).Format("2006")
	} else if month >= 1 && month <= 7 {
		Semester = "Semester 2"
		sem1year = time.Now().AddDate(-1, 0, 0).Format("2006")
		sem2year = time.Now().Format("2006")
	} else {
		log.Fatalf("Wtf? Your month falls outside 1-12")
	}
	AcadYear := sem1year + "/" + sem2year
	fmt.Println("\nThe current Academic Year is:", AcadYear, Semester)
	ivleconfig.AcadYear = strings.Trim(AcadYear, " \n\t")
	ivleconfig.Semester = strings.Trim(Semester, " \n\t")

	// Get DownloadLocation
	currdir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("\nWHERE WOULD YOU LIKE TO DOWNLOAD YOUR IVLE FOLDERS TO?")
	fmt.Println("e.g. ~/Downloads/NUS. You can edit this later")
	fmt.Println("↓↓↓   Leave blank to download files into the current folder   ↓↓↓")
	fmt.Println("~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~")
	fmt.Println(currdir)
	fmt.Println("~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~")
	fmt.Print("ivled Download Location: ")
	DownloadLocation, _ := reader.ReadString('\n')
	DownloadLocation = strings.Trim(DownloadLocation, " \n\t")
	if DownloadLocation == "" {
		DownloadLocation = currdir
	}
	switch runtime.GOOS {
	case "linux":
		DownloadLocation = strings.Replace(DownloadLocation, "~", "$HOME", 1)
	case "darwin":
		DownloadLocation = strings.Replace(DownloadLocation, "~", "$HOME", 1)
	case "windows":
		DownloadLocation = strings.Replace(DownloadLocation, "~", "%userprofile%", 1)
	default:
		log.Fatalln("unsupported platform")
	}
	DownloadLocation = os.ExpandEnv(DownloadLocation)
	DownloadLocation = strings.TrimSuffix(DownloadLocation, "/")
	DownloadLocation = strings.TrimSuffix(DownloadLocation, "\\")
	fmt.Println("You have chosen to download files into:", DownloadLocation + "\n")
	ivleconfig.DownloadLocation = DownloadLocation

	// Get ModulesThisSem
	fmt.Println("=====================================")
	fmt.Println("GETting your modules this semester..")
	fmt.Println("      (This may take some time)      ")
	fmt.Println("=====================================")
	ivleresponse, _ := IVLEGetRequest("https://ivle.nus.edu.sg/api/Lapi.svc/Modules_Taken?APIKey=" + ivleconfig.LAPIkey + "&AuthToken=" + ivleconfig.AuthToken + "&StudentID=" + ivleconfig.StudentID)
	var moduleinfos []ModuleInfo
	json.Unmarshal(ivleresponse.Results, &moduleinfos)
	cprint(moduleinfos)
	moduleinfos = FilterModuleInfo(moduleinfos, func(mi ModuleInfo) bool {
		return mi.AcadYear == ivleconfig.AcadYear && mi.SemesterDisplay == ivleconfig.Semester
	})
	moduleinfos = MapModuleInfo(moduleinfos, func(mi ModuleInfo) ModuleInfo {
		fmt.Println("GETting module info for :", mi.ModuleCode)
		ivleresponse, _ := IVLEGetRequest("https://ivle.nus.edu.sg/api/Lapi.svc/Modules_Search?APIKey=" + ivleconfig.LAPIkey + "&AuthToken=" + ivleconfig.AuthToken + "&IncludeAllInfo=false&ModuleCode=" + mi.ModuleCode)
		var courseinfos []CourseInfo
		json.Unmarshal(ivleresponse.Results, &courseinfos)
		courseinfos = FilterCourseInfo(courseinfos, func(ci CourseInfo) bool {
			return ci.CourseAcadYear == mi.AcadYear && ci.CourseSemester == mi.SemesterDisplay
		})
		if len(courseinfos) <= 0 {
			log.Fatalln("The module ID for", mi.ModuleCode, "was not found")
		}
		mi.ID = courseinfos[0].ID
		return mi
	})
	ivleconfig.ModulesThisSem = moduleinfos

	// Set Up the initial ExcludedFileTypes and ExcludedFilePaths
	excludedfiletypes := map[string]bool{
		"mp4": true,
		"mp3": true,
		"mov": true,
		"avi": true,
	}
	excludedfilepaths := map[string]bool{}
	ivleconfig.ExcludedFileTypes = excludedfiletypes
	excludedfilepaths[DownloadLocation+fpdlm+"Folder1"] = true
	excludedfilepaths[DownloadLocation+fpdlm+"Folder2"] = true
	ivleconfig.ExcludedFilePaths = excludedfilepaths

	// Write ivleconfig struct into config file
	fh, _ := os.OpenFile(configfile, os.O_WRONLY|os.O_CREATE, 0666)
	json, _ := JSONMarshalIndent(ivleconfig, true)
	fh.Truncate(0)
	fh.Seek(0, 0)
	fh.Write(json)
	defer fh.Close()

	return ivleconfig
}

//===============================//
// HERE LIE THE HELPER FUNCTIONS //
//===============================//

func DownloadFileIfNotExist(filepath string, fileid string) {
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		// Get the data
		fmt.Println("Downloading :", filepath)
		resp, err := http.Get("https://ivle.nus.edu.sg/api/downloadfile.ashx?APIKey=" + ivleconfig.LAPIkey + "&AuthToken=" + ivleconfig.AuthToken + "&ID=" + fileid + "&target=workbin")
		if err != nil {
			fmt.Println("GET request for", filepath, "("+fileid+")", "failed with error:", err)
		}
		defer resp.Body.Close()

		// Create the file
		out, err := os.Create(filepath)
		if err != nil {
			fmt.Println("os.Create(filepath) failed:", err)
		}
		defer out.Close()

		// Write the data to file
		_, err = io.Copy(out, resp.Body)
		if err != nil {
			fmt.Println("io.Copy(out, resp.Body) failed:", err)
		}

		// Append to list of downloaded files
		downloadedfiles = append(downloadedfiles, filepath)
	}
}

func CreateDirIfNotExist(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			panic(err)
		}
	}
}

// Converts a struct to JSON byte array with proper indentation
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

func openfile(filepath string) {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", filepath).Start()
	case "windows":
		err = exec.Command("notepad", filepath).Start()
	case "darwin":
		err = exec.Command("open", filepath).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Fatal(err)
	}
}

func deletefile(filepath string) {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("rm", filepath).Start()
	case "windows":
		err = exec.Command("del", filepath).Start()
	case "darwin":
		err = exec.Command("rm", filepath).Start()
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
