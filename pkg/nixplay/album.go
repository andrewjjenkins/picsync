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

func (c *clientImpl) GetAlbumsByName(albumName string) ([]*Album, error) {
	allAlbums, err := c.GetAlbums()
	if err != nil {
		return nil, err
	}
	var npAlbums []*Album
	for _, a := range allAlbums {
		if a.Title == albumName {
			npAlbums = append(npAlbums, a)
		}
	}
	if len(npAlbums) == 0 {
		c.prom.getAlbumByNameEmpty.Inc()
	} else {
		c.prom.getAlbumByNameSuccess.Inc()
	}
	return npAlbums, nil
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

func (c *clientImpl) DeleteAlbumByID(id int) error {
	vals := url.Values{}
	url := fmt.Sprintf("https://api.nixplay.com/album/%d/delete/json/", id)
	fmt.Printf("POST to %s\n", url)
	res, err := doPost(c.httpClient, url, &vals)
	if err != nil {
		c.prom.deleteAlbumFailure.Inc()
		return err
	}
	defer res.Body.Close()
	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		c.prom.deleteAlbumFailure.Inc()
		return err
	}
	if res.StatusCode != http.StatusOK {
		c.prom.createAlbumFailure.Inc()
		return fmt.Errorf("Couldn't delete album %d: http %d: %s", id,
			res.StatusCode, resBody)
	}

	// We don't care about the body
	return nil
}

func (c *clientImpl) DeleteAlbumsByName(albumName string, allowMultiple bool) (int, error) {
	allAlbums, err := c.GetAlbums()
	if err != nil {
		return 0, err
	}
	var matchingAlbums []*Album
	for _, a := range allAlbums {
		if a.Title == albumName {
			matchingAlbums = append(matchingAlbums, a)
		}
	}
	if len(matchingAlbums) == 0 {
		return 0, nil
	}
	if len(matchingAlbums) > 1 && allowMultiple == false {
		return 0, fmt.Errorf(
			"%d albums named %s, but only allowed to delete one (see \"--delete-multiple\")",
			len(matchingAlbums),
			albumName,
		)
	}
	var deletedCount int
	for _, a := range matchingAlbums {
		err := c.DeleteAlbumByID(a.ID)
		if err != nil {
			return deletedCount, err
		}
		deletedCount++
	}
	return deletedCount, nil
}
