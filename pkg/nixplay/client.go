package nixplay

import (
	"io"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type Client interface {
	GetAlbums() ([]*Album, error)
	GetAlbumsByName(albumName string) ([]*Album, error)
	CreateAlbum(albumName string) (*Album, error)
	DeleteAlbumsByName(albumName string, allowMultiple bool) (int, error)
	DeleteAlbumByID(albumID int) error
	GetPhotos(albumID int, page int, limit int) ([]*Photo, error)
	UploadPhoto(albumID int, filename string, filetype string, filesize uint64, body io.ReadCloser) error
	DeletePhoto(id int) error
	CreatePlaylist(name string) (int, error)
	GetPlaylists() ([]*Playlist, error)
	GetPlaylistByName(name string) (*Playlist, error)
	PublishPlaylist(playlistId int, photos []*Photo) error
}

type clientImpl struct {
	httpClient *http.Client

	prom promImpl
}

// NewClient logs in to Nixplay and returns a Client for future requests
func NewClient(username, password string, reg prometheus.Registerer) (Client, error) {
	auth, err := doLogin(username, password)
	if err != nil {
		return nil, err
	}
	tr := &http.Transport{
		ResponseHeaderTimeout: time.Duration(600 * time.Second),
	}
	httpClient := &http.Client{
		Timeout:   time.Duration(600 * time.Second),
		Transport: tr,
		Jar:       auth.Jar,
	}
	client := clientImpl{
		httpClient: httpClient,
	}
	err = client.promRegister(reg)
	if err != nil {
		return nil, err
	}
	return &client, nil
}
