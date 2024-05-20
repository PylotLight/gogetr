package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/zeebo/bencode"
	"gopkg.in/yaml.v3"
)

var ConfigFile = "config/config.yaml"

// func scheduler() {
// 	for {
// 		config := GetConfig()
// 		// CurrentTime := time.Now().In(config.NextRunTime.Location())
// 		// if config.NextRunTime.Before(CurrentTime) {
// 		ProcessNewFiles()
// 		// }
// 		config.LastRunTime = time.Now().In(config.NextRunTime.Location())
// 		config.NextRunTime = config.LastRunTime.Add(time.Minute + 2)
// 		SetConfig(config)
// 		nextrunTime := "Next scan time: " + config.NextRunTime.Format(time.ANSIC)
// 		println(nextrunTime)
// 		sendClientMessage("Next scan at: " + config.NextRunTime.Local().Format(time.Kitchen))
// 		time.Sleep(time.Minute * 1)
// 	}
// }

func watcher() {
	// Create new watcher.
	watcher, err := fsnotify.NewWatcher()
	config := GetConfig()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	// Start listening for events.
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				// log.Println("event:", event)
				if event.Has(fsnotify.Create) {
					time.Sleep(time.Second * 1)
					log.Println("New file detected:", event.Name)
					sendClientMessage("New file " + event.Name + " detected at: " + time.Now().Local().Format(time.Kitchen))
					ProcessNewFiles()
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	// Add a path.
	err = watcher.Add(config.Import)
	if err != nil {
		log.Fatal(err)
	}

	// Block main goroutine forever.
	<-make(chan struct{})
}

func SetConfig(Configuration Configuration) {
	//get current config
	//get update options for setting var
	//combine?
	//set

	//config := GetConfig()

	data, _ := yaml.Marshal(Configuration)
	os.WriteFile(ConfigFile, data, 0644)
}

func GetConfig() Configuration {

	configuration := Configuration{}
	//check if file exists and create if it doesnt, otherwise proceeed to read the file and return the config data
	if _, err := os.Stat(ConfigFile); errors.Is(err, os.ErrNotExist) {
		os.MkdirAll("config/", os.ModePerm)

		configuration.APIKey = os.Getenv("APIKey")
		configuration.Import = os.Getenv("ImportFolder")
		configuration.Export = os.Getenv("ExportFolder")

		loc, locerr := time.LoadLocation("Australia/Melbourne")
		if locerr != nil {
			println(locerr.Error())
		}
		configuration.NextRunTime = time.Now().In(loc).Add(time.Hour + 1)
		configuration.LastRunTime = time.Now().In(loc)
		if os.Getenv("APIKey") == "" {
			configuration.APIKey = "You must set an API Key"
		}
		_, err := os.Stat("/Music")

		if os.IsNotExist(err) {
			// Output an error message indicating that the path does not exist
			log.Println("Error: the /Music path does not exist")
		}

		configuration.MediaTypes = append(configuration.MediaTypes, "mp4", "mkv")
		data, _ := yaml.Marshal(configuration)
		fileerr := os.WriteFile(ConfigFile, data, 0644)
		if fileerr != nil {
			println("WriteFile:", err.Error())
		}
	}

	configyml, err := os.ReadFile(ConfigFile)
	if err != nil {
		log.Fatal(err)
	}

	err2 := yaml.Unmarshal(configyml, &configuration)
	if err2 != nil {

		log.Fatal(err2)
	}
	return configuration
}

// Get torrent files and send them process files in the required format
func GetFiles() []NewDownloadFile {
	config := GetConfig()
	var files []string
	var filesDetails []NewDownloadFile

	mags, _ := filepath.Glob(config.Import + "*.magnet")
	tors, _ := filepath.Glob(config.Import + "*.torrent")
	files = append(files, mags...)
	files = append(files, tors...)

	for _, file := range files {
		NewFile := NewDownloadFile{}
		fileinfo, _ := os.Stat(file)
		NewFile.Filename = file
		NewFile.Description = fileinfo.Name()
		NewFile.LastModified = fileinfo.ModTime()
		NewFile.FileCreated = fileinfo.ModTime()
		filesDetails = append(filesDetails, NewFile)
	}
	return filesDetails
}

func GetFolders(path string) []string {
	// config := GetConfig()
	// where item == config.import or export already exists, get folders from that location otherwise get from /
	// var root = "/"
	// if item == "import" {
	// 	root = config.Import
	// }
	// if item == "export" {
	// 	root = config.Export
	// }

	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("failed to get wd: %v", err)
	}
	items, err := os.ReadDir(path)
	if err != nil {
		log.Fatalf("failed to read wd: %v", err)
	}
	var dirs []string
	for _, item := range items {
		if item.IsDir() {
			dirs = append(dirs, filepath.Join(wd, item.Name()))
		}
	}
	return dirs
}

// func getDirectories(dirPath string) ([]string, error) {
// 	var dirs []string

// 	// Open the directory
// 	dir, err := os.Open(dirPath)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer dir.Close()

// 	// Read the directory entries
// 	fileInfos, err := dir.Readdir(0)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// Filter and collect the directories
// 	for _, fileInfo := range fileInfos {
// 		if fileInfo.IsDir() {
// 			dirs = append(dirs, fileInfo.Name())
// 		}
// 	}

// 	return dirs, nil
// }

func GetMagnetLink(NewFile NewDownloadFile) NewDownloadFile {

	var decodeSingle TorrentSingle
	var decodeMulti TorrentMulti
	f, _ := os.Open(NewFile.Filename)

	filedata, _ := io.ReadAll(f)
	var encode []byte
	// altdecode, _ := altbencode.Unmarshal(filedata)
	if strings.Contains(NewFile.Filename, ".torrent") {

		bencode.DecodeBytes(filedata, &decodeSingle)
		bencode.DecodeBytes(filedata, &decodeMulti)

		if decodeSingle.Info.Length == 0 {
			encode, _ = bencode.EncodeBytes(decodeMulti.Info)
			NewFile.Description = decodeMulti.Info.Name
			NewFile.Magnet = "magnet:?xt=urn:btih:" + fmt.Sprintf("%x", decodeMulti.InfoHash(encode))
			return NewFile
		}
		if decodeSingle.Info.Length != 0 {
			encode, _ = bencode.EncodeBytes(decodeSingle.Info)
			NewFile.Description = decodeSingle.Info.Name
			NewFile.Magnet = "magnet:?xt=urn:btih:" + fmt.Sprintf("%x", decodeSingle.InfoHash(encode))
			return NewFile
		}
	}
	if strings.Contains(strings.ToLower(NewFile.Filename), ".magnet") {
		NewFile.Magnet = string(filedata)
		return NewFile
	}
	NewFile.local = true
	return NewFile
}

// convert files to submit
func GetDownloadItems() []NewDownloadFile {
	//Get file data, filename, path, dates
	//send to get link where we add the magnetlink for the torrent
	files := GetFiles()
	var fileitems []NewDownloadFile
	// print("Grabbing maglink for files")
	// config := GetConfig()
	for _, file := range files {
		fileitems = append(fileitems, GetMagnetLink(file))
	}
	// print("Got maglink for files")
	return fileitems
}

// Send task to RD(scheduler?) return status of submission and files info
func RDAPI[T any](Method string, Endpoint string, Body string) (T, error) {
	ContentType := "application/x-www-form-urlencoded"

	var result T

	reqBody := strings.NewReader(Body)

	req, reqerror := http.NewRequest(Method, "https://api.real-debrid.com/rest/1.0/"+Endpoint, reqBody)
	if reqerror != nil {
		println("reqerror:", reqerror.Error())
	}
	// println("252")
	client := &http.Client{}
	req.Header.Add("Authorization", "Bearer "+GetConfig().APIKey)
	req.Header.Add("Content-Type", ContentType)
	resp, err := client.Do(req)
	// respjson, _ := json.Marshal(resp)
	// print(string(respjson))
	if err != nil {
		fmt.Println("Error in request: ", err.Error())
		return result, err
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusNoContent {
		println("Status: ", resp.StatusCode)
		return result, err
	}
	// println("263")
	defer resp.Body.Close()
	// println("264")
	data, reserror := io.ReadAll(resp.Body)
	if reserror != nil {
		println("reserror:", reserror.Error())
	}
	// println("269")
	jsonerr := json.Unmarshal(data, &result)
	// println("271")
	if jsonerr != nil {
		println(jsonerr.Error())
	}
	// fmt.Println("275")
	//try return generic interface that is typed outside func and can return any type

	return result, err
}

// select the files
func AutoHandleNewFile(TaskID string, ndf NewDownloadFile) error {
	var fileselection []string
	var Downloaded bool
	config := GetConfig()

	for !Downloaded {
		files, _ := RDAPI[TorrentInfo]("GET", "torrents/info/"+TaskID, "")
		switch files.Status {
		case "waiting_files_selection":
			{
				// if getconfig().GrabALl == true, set files=all

				for _, v := range files.Files {
					// skip sample files
					if strings.Contains(strings.ToLower(v.Path), "sample") {
						continue
					}

					extension := v.Path[len(v.Path)-3:]
					for _, i := range config.MediaTypes {
						if extension == i {
							fileselection = append(fileselection, strconv.Itoa(v.ID))
						}
					}

				}
				RDAPI[any]("POST", "torrents/selectFiles/"+TaskID, "files="+strings.Join(fileselection, ","))
				// RDAPI[any]("POST", "torrents/selectFiles/"+TaskID, "files=all")
			}
		// case "magnet_conversion":
		// 	{
		// 	}
		case "Invalid Magnet":
			{
				// os.Remove(files.Filename)
				// delete the file cause it doesn't work.
				// DeleteFile(files.Filename)
				fmt.Println("Skipping dud file " + files.ID)
				continue
			}
		case "downloaded":
			{

				Downloaded = true
				for _, i := range files.Links {
					UnrestrictedLink, _ := RDAPI[UnrestrictedLink]("POST", "unrestrict/link/"+TaskID, "link="+i)
					if !ndf.local {
						sendClientMessage("Download here: " + UnrestrictedLink.Download)
						// DeleteFile(ndf.Filename)
						return nil
					}
					resp, err := http.Get(UnrestrictedLink.Download)
					if err != nil {
						log.Fatal(err)
					}
					defer resp.Body.Close()
					FullName := config.Export + UnrestrictedLink.Filename
					println("Downloading:", FullName)
					fileHandle, err := os.OpenFile(FullName, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
					if err != nil {
						panic(err)
					}
					defer fileHandle.Close()
					buf := make([]byte, 16384)
					n, err := io.CopyBuffer(fileHandle, resp.Body, buf)
					if err != nil {
						return err
					}
					sendClientMessage("Downloaded " + UnrestrictedLink.Filename)
					fmt.Printf("Downloaded a file %s with size %d and bytes copied %d", UnrestrictedLink.Filename, UnrestrictedLink.Filesize, n)

				}
				DeleteFile(ndf.Filename)
			}
		case "":
			{
				fmt.Println("Skipping dud file " + files.ID)
				DeleteFile(ndf.Filename)
				return nil
			}
		default:
			{
				// j, _ := json.Marshal(files)
				println("Waiting for 30 seconds.. Torrent Progress:", files.Progress)
				sendClientMessage("Waiting for 30 seconds.. Torrent Progress: " + fmt.Sprint(files.Progress))
				// println(string(j))
				time.Sleep(30 * time.Second)
			}
		}
	}

	return nil
}

// func ManualFileSelection(link string) {
// 	// We need to run a get to collect available files and return that in array to be presented.
// }

func DeleteFile(path string) {
	println("Deleting..", path)
	os.Remove(path)
}

func ProcessNewFiles() {
	for _, v := range GetDownloadItems() {
		body := "magnet=" + v.Magnet
		resp, _ := RDAPI[MagnetCreated]("POST", "torrents/addMagnet", body)
		ID := resp.ID
		AutoHandleNewFile(ID, v) //Get available files to select  GET /torrents/info/{id} //Select the relevant files from the torrent POST /torrents/selectFiles/{id}

		//check if ready GET /torrents/info/{id}
		//if ready get dl links POST /unrestrict/link
		//if not ready, loop till ready GET /torrents/info/{id}
		//dl GET
	}
}

// After download, delete torrent files and log data
func CleanDownloads() {}
