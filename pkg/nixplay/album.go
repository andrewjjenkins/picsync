package nixplay

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/andrewjjenkins/picsync/pkg/util"
)

// Img is a reference to an image
type Img struct {
	Src         string `json:"src"` // URL in cloudfront
	Rotation    int    `json:"rotation"`
	Orientation int    `json:"orientation"` // 1 means portrait?  Not rotated?
}

// Album is a set of Imgs; Albums can be added to Playlists
type Album struct {
	AlbumType  string `json:"album_type"` // "Web"?
	PhotoCount int    `json:"photo_count"`
	// PendingAlbumId
	AllowReorder        bool     `json:"allow_reorder"`
	Title               string   `json:"title"`
	IsUpdated           bool     `json:"is_updated"`
	CoverUrls           []string `json:"cover_urls"`  // []string?
	DateCreated         string   `json:"dateCreated"` // 31/Jul/2018
	AllowUpload         bool     `json:"allow_upload"`
	AllowDelete         bool     `json:"allow_delete"`
	Thumbs              []Img    `json:"thumbs"`
	Published           bool     `json:"published"`
	ID                  int      `json:"id"`
	AllowDeletePictures bool     `json:"allow_delete_pictures"`
	PictureStorageType  string   `json:"picture_storage_type"`
	Email               string   `json:"email"` // Empty if owned by you?
}

func (a Album) String() string {
	return fmt.Sprintf(
		"%s (%d, %s), %d photos, type \"%s\"",
		a.Title, a.ID, a.DateCreated, a.PhotoCount, a.AlbumType,
	)
}

// GetAlbums will get a list of Albums available to this user
func (c *clientImpl) GetAlbums() ([]*Album, error) {
	albums := []*Album{}
	err := util.GetUnmarshalJSON(c.httpClient, "https://api.nixplay.com/albums/web/json/", &albums)
	if err != nil {
		c.prom.getAlbumsFailure.Inc()
	} else {
		c.prom.getAlbumsSuccess.Inc()
	}
	return albums, err
}

func (c *clientImpl) GetAlbumByName(albumName string) (*Album, error) {
	npAlbums, err := c.GetAlbums()
	if err != nil {
		return nil, err
	}
	var npAlbum *Album
	for _, a := range npAlbums {
		if a.Title == albumName {
			if npAlbum != nil {
				c.prom.getAlbumByNameFailure.Inc()
				return nil, fmt.Errorf("duplicate Nixplay albums named %s", albumName)
			}
			npAlbum = a
		}
	}
	if npAlbum == nil {
		c.prom.getAlbumByNameFailure.Inc()
		return nil, fmt.Errorf("could not find Nixplay album %s", albumName)
	}
	c.prom.getAlbumByNameSuccess.Inc()
	return npAlbum, nil
}

func (c *clientImpl) CreateAlbum(name string) (*Album, error) {
	vals := url.Values{
		"name": []string{name},
	}
	res, err := doPost(c.httpClient, "https://api.nixplay.com/album/create/json/", &vals)
	if err != nil {
		c.prom.createAlbumFailure.Inc()
		return nil, err
	}
	defer res.Body.Close()
	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		c.prom.createAlbumFailure.Inc()
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		c.prom.createAlbumFailure.Inc()
		return nil, fmt.Errorf("Couldn't create album %s: http %d: %s", name,
			res.StatusCode, resBody)
	}

	var resData []*Album
	err = json.NewDecoder(bytes.NewBuffer(resBody)).Decode(&resData)
	if err != nil {
		c.prom.createAlbumFailure.Inc()
		return nil, err
	}
	if len(resData) != 1 {
		c.prom.createAlbumFailure.Inc()
		return nil, fmt.Errorf("expected create to return 1 album, got %d", len(resData))
	}
	c.prom.createAlbumSuccess.Inc()
	return resData[0], nil
}
