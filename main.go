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
- Clean up home page with nicer dashboard of buttons - done
- music download notification - done
- remove exessive logging (debug flag) - done
New:
- better file type selector on the settings page - na
- replace html with better templating? - na
- better error handling for RD file selection and API stuff - maby
- fix select all files setting to apply on selection - maby
- fix import/export folder to have directory searcher to select paths - maby
- ability to add custom torrents like games or movies etc outside of sonarr. magnet selector with display for items to select.
- show download progress so that we dont have to check folder
-  add web torrent add.
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
	SetConfig(*config)
}

func main() {
	go watcher()
	// Build
	staticSubFS, _ := fs.Sub(staticFS, "static")
	mux := http.NewServeMux()

	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticSubFS))))
	// rootPath = "/"
	// Debug
	// http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	// http.Handle("/static/", http.StripPrefix(strings.TrimRight("static", "/"), http.FileServer(http.Dir("/static/"))))
	mux.HandleFunc("/", handler)
	mux.HandleFunc("/api/handshake", SSEhandler)
	mux.HandleFunc("/api/settings", SettingsHandler)
	mux.HandleFunc("/api/tasks", TasksHandler)
	mux.HandleFunc("/api/getfolders", GetFoldersHandler)
	mux.HandleFunc("/api/ytdownload", DownloadYTHandler)
	mux.HandleFunc("/api/rddownload", DownloadRDHandler)
	// http.HandleFunc("/", handleFolderBrowser)
	// http.HandleFunc("/browse", handleBrowse)
	go eventLoop()
	fmt.Println("Listening on http://localhost:9000")
	log.Fatal("HTTP server error: ", http.ListenAndServe("0.0.0.0:9000", mux))

}
