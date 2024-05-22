package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"

	"github.com/kkdai/youtube/v2"
)

func GetMedia(link string, w http.ResponseWriter, r *http.Request) error {
	// Create a youtube client
	client := youtube.Client{}
	format := &youtube.Format{}
	videoIDs := []string{}
	// Check if link is a playlist, if its a playlist loop through all videos and append to videoIDs
	playlist, _ := client.GetPlaylist(link)
	// https://www.youtube.com/watch?v=vXYXHHYdwTo&list=PLVjTe37QSG1ecSZefjVDI6P-To5FWv2Nr
	// println(playlist.Title, playlist.ID)
	// if err != nil {
	// 	// Return the error if it occurs
	// 	log.Printf("error getting playlist: %v", err)
	// }

	if playlist != nil && len(playlist.Videos) > 0 {
		// Get the first video in the playlist
		println("Playlist found, number of videos:", len(playlist.Videos))
		for _, video := range playlist.Videos {
			videoIDs = append(videoIDs, video.ID)
		}
	}

	if playlist == nil || len(playlist.Videos) == 0 {
		// Extract the video ID from the link
		videoID, err := youtube.ExtractVideoID(link)
		if err != nil {
			// Return the error if it occurs
			log.Printf("error extracting video ID: %v", err)
		}

		videoIDs = append(videoIDs, videoID)
	}

	for _, videoID := range videoIDs {
		// Get the video from youtube
		video, err := client.GetVideo(videoID)
		if err != nil {
			// Return the error if it occurs
			log.Printf("error getting video: %v", err)
			continue
		}

		preferences := []int{338, 251, 250, 249, 140, 327}

		// Loop through the preferences and find the first matching format
		for _, pref := range preferences {
			for i := range video.Formats {
				if video.Formats[i].ItagNo == pref {
					format = &video.Formats[i]
					break
				}
			}
			if format.ItagNo != 0 {
				break
			}
		}

		// Download the video using the selected format
		stream, size, FileName, err := getDownloadStream(video, format)
		if err != nil {
			// Return the error if it occurs
			log.Printf("error getting stream: %v", err)
			continue
		}
		// if local {
		downloadLocal(stream, size, FileName)
		// }
		// if !local {
		// downloadPublic(stream, w, r)
		// }
		fmt.Fprint(w, "<pre>Download completed successfully:\n", FileName, "</pre>")

	}
	return nil
}

func getDownloadStream(video *youtube.Video, format *youtube.Format) (io.ReadCloser, int64, string, error) {
	client := youtube.Client{}
	re := regexp.MustCompile(`[\\/:*?"<>|]`)
	videoTitle := re.ReplaceAllString(video.Title, "-")
	var FileName string
	switch GetConfig().AppVersion {
	case "":
		FileName = "Music/" + videoTitle + "." + "opus"
	default:
		FileName = "/Music/" + videoTitle + "." + "opus"
	}
	stream, size, err := client.GetStream(video, format)
	if err != nil {
		return nil, 0, "", fmt.Errorf("error getting video stream: %v", err)
	}

	return stream, size, FileName, nil
}

func downloadLocal(stream io.ReadCloser, size int64, FileName string) error {
	log.Printf("Downloading video: %s", FileName)
	sendClientMessage("Downloading " + strconv.FormatInt(size, 10) + " bytes to path: " + FileName)
	file, err := os.Create(FileName)
	if err != nil {
		return fmt.Errorf("error creating file: %v", err)
	}
	defer file.Close()
	// Create a buffer to read the response body into.
	buf := make([]byte, 4096)
	// Read the response body into the buffer, and write it to the file.
	written, err := io.CopyBuffer(file, stream, buf)
	if err != nil {
		return err
	}
	mb := float64(written) / 1024 / 1024
	sendClientMessage("Finished downloading video: " + FileName)
	log.Printf("Copied %.1fMB\n", mb)

	return nil
}

func downloadPublic(stream io.ReadCloser, w http.ResponseWriter, r *http.Request) {

	defer stream.Close()

	// Create a temporary file
	tempFile, err := os.CreateTemp("", "downloaded_file_*.opus")
	if err != nil {
		http.Error(w, "Failed to create temporary file", http.StatusInternalServerError)
		return
	}
	defer os.Remove(tempFile.Name()) // Clean up the temporary file

	// Write the stream content to the temporary file
	_, err = io.Copy(tempFile, stream)
	if err != nil {
		http.Error(w, "Failed to write data to temporary file", http.StatusInternalServerError)
		return
	}

	// Set the headers for the download
	w.Header().Set("Content-Disposition", "attachment; filename=\"downloaded_file.opus\"")
	w.Header().Set("Content-Type", "audio/opus")

	// Serve the temporary file
	http.ServeFile(w, r, tempFile.Name())
}
