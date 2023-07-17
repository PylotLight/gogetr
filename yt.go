package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"regexp"

	"github.com/kkdai/youtube/v2"
)

func GetMedia(link string) error {
	// Create a youtube client
	client := youtube.Client{}
	format := &youtube.Format{}

	// Extract the video ID from the link
	videoID, err := youtube.ExtractVideoID(link)
	if err != nil {
		// Return the error if it occurs
		return fmt.Errorf("error extracting video ID: %v", err)
	}

	// Get the video from youtube
	video, err := client.GetVideo(videoID)
	if err != nil {
		// Return the error if it occurs
		return fmt.Errorf("error getting video: %v", err)
	}

	// Find a suitable format for the video
	for i := range video.Formats {
		switch video.Formats[i].ItagNo {
		case 251, 250:
			format = &video.Formats[i]
			// break
		}
	}

	// Download the video using the selected format
	err = DownloadVideo(video, format)
	if err != nil {
		// Return the error if it occurs
		return fmt.Errorf("error downloading video: %v", err)
	}

	return nil
}

func DownloadVideo(video *youtube.Video, format *youtube.Format) error {
	// Create a youtube client
	client := youtube.Client{}
	re := regexp.MustCompile(`[\\/:*?"<>|]`)
	videoTitle := re.ReplaceAllString(video.Title, "-")
	// Get the video file name
	title := GetConfig().Export + videoTitle + "." + "opus"

	// Log the video being downloaded
	log.Printf("Downloading video: %s", title)
	sendClientMessage("Downloading video to path: " + title)
	// Get the video stream from youtube
	stream, _, err := client.GetStream(video, format)
	if err != nil {
		// Return the error if it occurs
		return fmt.Errorf("error getting video stream: %v", err)
	}

	// Open the file for writing
	file, err := os.Create(title)
	if err != nil {
		// Return the error if it occurs
		return fmt.Errorf("error creating file: %v", err)
	}
	defer file.Close()

	// Copy the video stream to the file
	_, err = io.Copy(file, stream)
	if err != nil {
		// Return the error if it occurs
		return fmt.Errorf("error copying video stream to file: %v", err)
	}
	sendClientMessage("Finished downloading video: " + videoTitle)
	// Log that the download is finished
	log.Println("Finished downloading video")

	return nil
}
