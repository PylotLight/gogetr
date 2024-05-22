package main

import (
	"crypto/sha1"
	"time"
)

type Configuration struct {
	APIKey      string
	Import      string
	Export      string
	SelectAll   bool
	LastRunTime time.Time
	NextRunTime time.Time
	AppVersion  string
	MediaTypes  []string
}

type LocalFolders struct {
	Name string
	Path string
}

type DirectoryData struct {
	CurrentPath string   // Current directory path
	Dirs        []string // List of directories in the current directory
}

type Page struct {
	Title  string
	Body   []byte
	Files  []NewDownloadFile
	Config Configuration
}

type DirectoryInfo struct {
	Directory []struct {
		Path string
		Name string
	}
}

type NewDownloadFile struct {
	Description  string
	Filename     string
	Magnet       string
	FileCreated  time.Time
	LastModified time.Time
	local        bool
}

type TorrentMulti struct {
	Announce     string     `bencode:"announce"`
	AnnounceList [][]string `bencode:"announce-list"`
	Comment      string     `bencode:"comment"`
	CreatedBy    string     `bencode:"created by"`
	CreationDate int        `bencode:"creation date"`
	Encoding     string     `bencode:"encoding"`
	Info         struct {
		Name        string `bencode:"name"`
		PieceLength int    `bencode:"piece length"`
		Pieces      string `bencode:"pieces"`
		Files       []struct {
			Length int      `bencode:"length"`
			Path   []string `bencode:"path"`
		} `bencode:"files"`
	} `bencode:"info"`
}

type TorrentSingle struct {
	Announce     string     `bencode:"announce"`
	AnnounceList [][]string `bencode:"announce-list"`
	Comment      string     `bencode:"comment"`
	CreatedBy    string     `bencode:"created by"`
	CreationDate int        `bencode:"creation date"`
	Encoding     string     `bencode:"encoding"`
	Info         struct {
		Length      int    `bencode:"length"`
		Name        string `bencode:"name"`
		PieceLength int    `bencode:"piece length"`
		Pieces      string `bencode:"pieces"`
	} `bencode:"info"`
}

func (t TorrentSingle) InfoHash(encode []byte) []byte {
	h := sha1.New()
	h.Write(encode)
	bs := h.Sum(nil)
	return bs
}

func (t TorrentMulti) InfoHash(encode []byte) []byte {
	h := sha1.New()
	h.Write(encode)
	bs := h.Sum(nil)
	return bs
}

func (p Page) SaveConfig(c *Configuration) error {
	currentconfig := GetConfig()
	currentconfig.APIKey = c.APIKey
	currentconfig.Export = c.Export
	currentconfig.Import = c.Import
	currentconfig.AppVersion = c.AppVersion
	currentconfig.MediaTypes = c.MediaTypes
	SetConfig(*currentconfig)
	return nil
}

type Data struct {
	Link string `json:"link"`
}

type MagnetCreated struct {
	ID  string `json:"id"`
	Uri string `json:"uri"` // URL of the created ressource
}

type TorrentInfo struct {
	ID               string    `json:"id"`
	Filename         string    `json:"filename"`
	OriginalFilename string    `json:"original_filename"`
	Hash             string    `json:"hash"`
	Bytes            int       `json:"bytes"`
	OriginalBytes    int       `json:"original_bytes"`
	Host             string    `json:"host"`
	Split            int       `json:"split"`
	Progress         int       `json:"progress"`
	Status           string    `json:"status"`
	Added            time.Time `json:"added"`
	Files            []struct {
		ID       int    `json:"id"`
		Path     string `json:"path"`
		Bytes    int    `json:"bytes"`
		Selected int    `json:"selected"`
	} `json:"files"`
	Links []string  `json:"links"`
	Ended time.Time `json:"ended"`
}

type UnrestrictedLink struct {
	ID         string `json:"id"`
	Filename   string `json:"filename"`
	MimeType   string `json:"mimeType"`
	Filesize   int    `json:"filesize"`
	Link       string `json:"link"`
	Host       string `json:"host"`
	HostIcon   string `json:"host_icon"`
	Chunks     int    `json:"chunks"`
	Crc        int    `json:"crc"`
	Download   string `json:"download"`
	Streamable int    `json:"streamable"`
}
