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
		Config: *GetConfig(),
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
	// link := r.FormValue("link")
	// link := r.PathValue("link")
	link := r.URL.Query().Get("link")
	if link == "" {
		http.Error(w, "Missing 'link' parameter", http.StatusBadRequest)
		return
	}
	println("Got link:", link)
	if err := GetMedia(link, w, r); err != nil {
		// In case of error during GetMedia, respond with a failure message.
		// Set headers for an HTML response before initiating the download
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, "<pre>Download failed for some reason: ", err, "</pre>")
		return
	}
}

func DownloadRDHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the incoming data
	var ndf NewDownloadFile
	link := r.FormValue("rdlink")

	// Launch a goroutine to handle the long-running task
	body := "magnet=" + link
	resp, err := RDAPI[MagnetCreated]("POST", "torrents/addMagnet", body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		sendClientMessage(err.Error())
		return
	}
	ID := resp.ID
	ndf.Magnet = link
	ndf.local = false
	sendClientMessage("Submitted download for " + resp.ID)
	unrestrictedLink, err := AutoHandleNewFile(ID, ndf)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	w.Header().Set("Content-Type", "text/html")
	// Your HTML content goes here
	fmt.Fprint(w, "<pre>Link is downloaded, grab from here:\n<a href='"+unrestrictedLink.Download+"'>"+unrestrictedLink.Filename+"</a></pre>")

	// Set the response headers
	// w.Header().Set("Content-Type", "application/json")

	// json.NewEncoder(w).Encode(responseData)
}

func SettingsHandler(w http.ResponseWriter, r *http.Request) {

	configuration := Configuration{}
	bytes, _ := io.ReadAll(r.Body)
	json.Unmarshal(bytes, &configuration)
	println(string(bytes))
	Page.SaveConfig(Page{}, &configuration)
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
