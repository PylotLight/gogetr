package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
)

func handler(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Path[len("/"):]
	sendClientMessage("Connected to server")
	files := GetDownloadItems()
	p := struct {
		Title  string
		Body   []byte
		Files  []NewDownloadFile
		Config Configuration
	}{
		Title:  title,
		Body:   []byte(""),
		Files:  files,
		Config: GetConfig(),
	}
	t, err := template.ParseFS(staticFS, "static/*.html")

	(*t).Execute(w, map[string]any{
		"Page": p,
		// "obj2": obj2,
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func SSEhandler(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	messageChan := make(chan []byte)
	register <- messageChan
	defer func() {
		unregister <- messageChan
		close(messageChan)
	}()

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case message := <-messageChan:
			fmt.Fprintf(w, "data: %s\n\n", message)
			flusher.Flush()
		}
	}
}

func sendClientMessage(message string) {
	data := map[string]interface{}{
		"message": message,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return
	}
	event := jsonData
	events <- []byte(event)
}

func eventLoop() {
	for {
		select {
		case client := <-register:
			clients[client] = true
		case client := <-unregister:
			delete(clients, client)
		case event := <-events:
			for client := range clients {
				client <- event
			}
		}
	}
}

func DownloadYTHandler(w http.ResponseWriter, r *http.Request) {

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form data", http.StatusBadRequest)
		return
	}
	link := r.FormValue("link")

	go GetMedia(link) // Run the download in the background

	// Set the response headers
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(struct {
		Message string `json:"message"`
		Success bool   `json:"success"`
	}{
		Message: "Audio download started successfully",
		Success: true,
	})
}

func DownloadRDHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the incoming data
	var ndf NewDownloadFile
	link := r.FormValue("rdlink")
	// Now we have a list of files, we return this to the user to select, then capture response again.
	// ManualFileSelection(ID)
	// HandleNewFile(ID, ndf) //Get available files to select  GET /torrents/info/{id} //Select the relevant files from the torrent POST /torrents/selectFiles/{id}
	// We've been able to add the magnet, then we need to present a selection window to user to select files to download. or we send a download all along with the request?
	// // Define the response data
	// responseData := struct {
	// 	// Message     string      `json:"message"`
	// 	Success     bool        `json:"success"`
	// 	TorrentInfo TorrentInfo `json:"TorrentInfo"`
	// }{}

	// Launch a goroutine to handle the long-running task
	body := "magnet=" + link
	resp, _ := RDAPI[MagnetCreated]("POST", "torrents/addMagnet", body)
	ID := resp.ID
	ndf.Magnet = link
	AutoHandleNewFile(ID, ndf)
	// files, _ := RDAPI[TorrentInfo]("GET", "torrents/info/"+ID, "")
	// responseData.TorrentInfo = files
	// responseData.Success = true

	// Set the response headers
	// w.Header().Set("Content-Type", "application/json")

	// Write the JSON response
	sendClientMessage("Submitted download for " + ID)
	// json.NewEncoder(w).Encode(responseData)
}

func SettingsHandler(w http.ResponseWriter, r *http.Request) {

	configuration := Configuration{}
	bytes, _ := io.ReadAll(r.Body)
	json.Unmarshal(bytes, &configuration)
	println(string(bytes))
	Page.SaveConfig(Page{}, configuration)
	http.Redirect(w, r, "/", http.StatusFound)
}

func TasksHandler(w http.ResponseWriter, r *http.Request) {

	bytes, _ := io.ReadAll(r.Body)
	println(string(bytes))

	if string(bytes) == "Yes" {
		ProcessNewFiles()
	}

	http.Redirect(w, r, "/", http.StatusFound)

}

func GetFoldersHandler(w http.ResponseWriter, r *http.Request) {
	// pass current path loaded from config which the browser is set to, / by default
	path := r.URL.RawQuery //strings.Split(r.URL.RawQuery, "=")[1]
	folders := GetFolders(path)

	// p, _ := loadPage("/", folders)
	// Load page data with new current path and available folders then send to render
	// Not sure if this will break the spa index page atm.
	v := &DirectoryData{CurrentPath: path, Dirs: folders}

	t, err := template.ParseFS(staticFS, "static/*.html")
	// (*t).Execute(w, map[string]any{
	// 	"DirectoryData": v,
	// 	// "obj2": obj2,
	// })
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	// renderTemplate(w, "static/index", v)
	err = t.Execute(w, v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// func handleFolderBrowser(w http.ResponseWriter, r *http.Request) {
// 	// Parse the template file
// 	tmpl := template.Must(template.ParseFiles("index.html"))

// 	// Get the current directory path from the URL query parameters
// 	currentPath := r.URL.Query().Get("path")
// 	if currentPath == "" {
// 		currentPath = rootPath
// 	}

// 	// Get the list of directories in the current directory
// 	dirs, err := getDirectories(currentPath)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	// Create the page data
// 	data := PageData{
// 		CurrentPath: currentPath,
// 		Dirs:        dirs,
// 	}

// 	// Execute the template with the page data
// 	err = tmpl.Execute(w, data)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}
// }

// func handleBrowse(w http.ResponseWriter, r *http.Request) {
// 	// Get the selected directory path from the form data
// 	selectedPath := r.FormValue("selectedPath")
// 	log.Println("Selected Path:", selectedPath)

// 	// Do something with the selected directory path, such as processing files within it
// 	// ...

// 	// Redirect back to the folder browser
// 	http.Redirect(w, r, "/", http.StatusFound)
// }
