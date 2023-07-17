package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
)

/* To Do
1. Setup app startup set envvars and file config done?
2. Setup functions to update mapped mounts to upload/dl dirs done
3. Get tasks from mounted folders done
4. Setup RD API Rest funcs starting with the gets, then the posts done
5. Setup web interface done
6. dockrise done
New:
1. better file type selector on the settings page - na
2. replace html with better templating? - na
3. better error handling for RD file selection and API stuff - maby
4. remove exessive logging (debug flag) - done
5. fix select all files setting to apply on selection - maby
6. fix import/export folder to have directory searcher to select paths - maby
7. ability to add custom torrents like games or movies etc outside of sonarr. magnet selector with display for items to select.
8. show download progress so that we dont have to check folder
9. Clean up home page with nicer dashboard of buttons - done
10. music download notification - done
*/

//go:embed static/*.html
// var templateFS embed.FS

//go:embed static/*.html
//go:embed static/*.css
//go:embed static/*.js
var staticFS embed.FS
var AppVersion string

// var messageChan chan []byte
var clients = make(map[chan []byte]bool)
var register = make(chan chan []byte)
var unregister = make(chan chan []byte)
var events = make(chan []byte)

// var (
// 	rootPath string // Root directory path
// )

func init() {
	// Check that the config file exists that stores mounted storage locations and other loggin/config data
	fmt.Println("Initialising app config..")

	config := GetConfig()
	config.AppVersion = AppVersion
	SetConfig(config)
}

func main() {
	go scheduler()
	// Build
	staticSubFS, _ := fs.Sub(staticFS, "static")
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticSubFS))))
	// rootPath = "/"
	// Debug
	// http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	// http.Handle("/static/", http.StripPrefix(strings.TrimRight("static", "/"), http.FileServer(http.Dir("/static/"))))
	http.HandleFunc("/", handler)
	http.HandleFunc("/api/handshake", SSEhandler)
	http.HandleFunc("/api/settings", SettingsHandler)
	http.HandleFunc("/api/tasks", TasksHandler)
	http.HandleFunc("/api/getfolders", GetFoldersHandler)
	http.HandleFunc("/api/youtube-download", DownloadYTHandler)
	http.HandleFunc("/api/rd-download", DownloadRDHandler)
	// http.HandleFunc("/", handleFolderBrowser)
	// http.HandleFunc("/browse", handleBrowse)
	go eventLoop()
	fmt.Println("Listening on http://localhost:9000")
	log.Fatal("HTTP server error: ", http.ListenAndServe("localhost:9000", nil))

}
