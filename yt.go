package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strconv"

	"github.com/kkdai/youtube/v2"
)

func GetMedia(link string) error {
	// Create a youtube client
	client := youtube.Client{}
	format := &youtube.Format{}
	videoIDs := []string{}

	// Check if link is a playlist, if its a playlist loop through all videos and append to videoIDs
	playlist, err := client.GetPlaylist(link)
	if err != nil {
		// Return the error if it occurs
		return fmt.Errorf("error getting playlist: %v", err)
	}

	if playlist != nil {
		// Get the first video in the playlist
		println("Playlist found, number of videos:", len(playlist.Videos))
		for _, video := range playlist.Videos {
			// videoID, err := youtube.ExtractVideoID(video.ID)
			// if err != nil {
			// 	// Return the error if it occurs
			// 	return fmt.Errorf("error extracting video ID: %v", err)
			// }
			videoIDs = append(videoIDs, video.ID)
		}
	}
	if playlist == nil {
		// Extract the video ID from the link
		videoID, err := youtube.ExtractVideoID(link)
		if err != nil {
			// Return the error if it occurs
			return fmt.Errorf("error extracting video ID: %v", err)
		}

		videoIDs = append(videoIDs, videoID)
	}

	go func() {
		for _, videoID := range videoIDs {
			// Get the video from youtube
			video, err := client.GetVideo(videoID)
			if err != nil {
				// Return the error if it occurs
				fmt.Errorf("error getting video: %v", err)
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
				fmt.Errorf("error downloading video: %v", err)
			}

		}
	}()

	return nil

}

func DownloadVideo(video *youtube.Video, format *youtube.Format) error {
	client := youtube.Client{}
	re := regexp.MustCompile(`[\\/:*?"<>|]`)
	videoTitle := re.ReplaceAllString(video.Title, "-")
	title := "Music/" + videoTitle + "." + "opus"

	stream, size, err := client.GetStream(video, format)
	if err != nil {
		return fmt.Errorf("error getting video stream: %v", err)
	}
	log.Printf("Downloading video: %s", title)
	sendClientMessage("Downloading " + strconv.FormatInt(size, 10) + " bytes to path: " + title)
	file, err := os.Create(title)
	if err != nil {
		return fmt.Errorf("error creating file: %v", err)
	}
	defer file.Close()
	// Create a buffer to read the response body into.
	buf := make([]byte, 4096)
	// Read the response body into the buffer, and write it to the file.
	n, err := io.CopyBuffer(file, stream, buf)
	if err != nil {
		return err
	}
	sendClientMessage("Finished downloading video: " + videoTitle)
	log.Printf("Copied %d bytes\n", n)

	return nil
}
