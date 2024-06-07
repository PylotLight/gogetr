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
	data, err := yaml.Marshal(Configuration)
	if err != nil {
		log.Println("Failed to marshal yaml")
	}
	os.WriteFile(ConfigFile, data, 0644)
}

func GetConfig() *Configuration {

	configuration := &Configuration{}
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
		_, err := os.Stat("/Music")
		if os.IsNotExist(err) {
			// Output an error message indicating that the path does not exist
			log.Println("Error: the /Music path does not exist")
		}

		configuration.MediaTypes = append(configuration.MediaTypes, "mp4", "mkv")
		data, err := yaml.Marshal(configuration)
		if err != nil {
			log.Println("Failed to marshal yaml")
			return nil
		}
		err = os.WriteFile(ConfigFile, data, 0644)
		if err != nil {
			log.Printf("WriteFile:%s", err.Error())
			return nil
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
		NewFile.local = true
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

func GetDownloadItems() []NewDownloadFile {
	files := GetFiles()
	var fileitems []NewDownloadFile

	for _, file := range files {
		fileitems = append(fileitems, GetMagnetLink(file))
	}
	return fileitems
}

func RDAPI[T any](Method string, Endpoint string, Body string) (T, error) {
	var result T

	reqBody := strings.NewReader(Body)

	req, reqerror := http.NewRequest(Method, "https://api.real-debrid.com/rest/1.0/"+Endpoint, reqBody)
	if reqerror != nil {
		return result, fmt.Errorf("request error: %w", reqerror)
	}

	client := &http.Client{}

	APIKey := GetConfig().APIKey
	if APIKey == "" {
		return result, errors.New("API Key is not set")
	}

	req.Header.Add("Authorization", "Bearer "+APIKey)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		return result, fmt.Errorf("error in request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusNoContent {
		return result, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	if resp.StatusCode == http.StatusNoContent {
		return result, nil
	}

	data, reserror := io.ReadAll(resp.Body)
	if reserror != nil {
		return result, fmt.Errorf("response read error: %w", reserror)
	}

	jsonerr := json.Unmarshal(data, &result)
	if jsonerr != nil {
		return result, fmt.Errorf("unable to unmarshal JSON: %w", jsonerr)
	}

	return result, nil
}

func ProcessNewFiles() {
	for _, v := range GetDownloadItems() {
		body := "magnet=" + v.Magnet
		resp, err := RDAPI[MagnetCreated]("POST", "torrents/addMagnet", body)
		if err != nil {
			log.Println("Error adding magnet: " + err.Error())
			return
		}
		_, err = AutoHandleNewFile(resp.ID, v)
		if err != nil {
			fmt.Printf("Unable to process file %s, due to error %s", v.Filename, err)
		}
	}
}

func AutoHandleNewFile(TaskID string, ndf NewDownloadFile) (*UnrestrictedLink, error) {
	var fileselection []string
	var Downloaded bool
	config := GetConfig()

	for !Downloaded {
		files, err := RDAPI[TorrentInfo]("GET", "torrents/info/"+TaskID, "")
		if err != nil {
			return nil, err
		}
		switch files.Status {
		case "waiting_files_selection":
			{
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
				_, err := RDAPI[any]("POST", "torrents/selectFiles/"+TaskID, "files="+strings.Join(fileselection, ","))
				if err != nil {
					return nil, err
				}
			}
		case "Invalid Magnet":
			{
				fmt.Println("Skipping dud file - invalid magnet" + files.ID)
				continue
			}
		case "downloaded":
			{
				Downloaded = true
				for _, i := range files.Links {
					unrestrictedLink, err := RDAPI[UnrestrictedLink]("POST", "unrestrict/link/"+TaskID, "link="+i)
					if err != nil {
						return nil, err
					}
					if !ndf.local {
						sendClientMessage("Download ready for: " + unrestrictedLink.Filename)
						return &unrestrictedLink, nil
					}
					resp, err := http.Get(unrestrictedLink.Download)
					if err != nil {
						return nil, fmt.Errorf("unable to unmarshal JSON: %w", err)
					}
					defer resp.Body.Close()
					FullName := config.Export + unrestrictedLink.Filename
					println("Downloading:", FullName)
					fileHandle, err := os.OpenFile(FullName, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
					if err != nil {
						return nil, fmt.Errorf("unable to open file handle: %w", err)
					}
					defer fileHandle.Close()
					buf := make([]byte, 16384)
					n, err := io.CopyBuffer(fileHandle, resp.Body, buf)
					if err != nil {
						return nil, fmt.Errorf("error copying file: %w", err)
					}
					sendClientMessage("Downloaded " + unrestrictedLink.Filename)
					fmt.Printf("Downloaded a file %s with size %d and bytes copied %d\n", unrestrictedLink.Filename, unrestrictedLink.Filesize, n)
				}
				DeleteFile(ndf.Filename)
			}
		case "":
			{
				DeleteFile(ndf.Filename)
				return nil, fmt.Errorf("skipping dud file %s", files.ID)
			}
		default:
			{
				println("Waiting for 30 seconds.. Torrent Progress:", files.Progress)
				sendClientMessage("Waiting for 30 seconds.. Torrent Progress: " + fmt.Sprint(files.Progress))
				time.Sleep(30 * time.Second)
			}
		}
	}

	return &UnrestrictedLink{}, nil
}
func DeleteFile(path string) {
	println("Deleting..", path)
	os.Remove(path)
}
