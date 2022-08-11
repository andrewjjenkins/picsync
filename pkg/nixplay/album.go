package nixplay

import (
	"fmt"
	"net/http"

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
func GetAlbums(c *http.Client) ([]*Album, error) {
	albums := []*Album{}
	err := util.GetUnmarshalJSON(c, "https://api.nixplay.com/albums/web/json/", &albums)
	return albums, err
}

func GetAlbumByName(c *http.Client, albumName string) (*Album, error) {
	npAlbums, err := GetAlbums(c)
	if err != nil {
		return nil, err
	}
	var npAlbum *Album
	for _, a := range npAlbums {
		if a.Title == albumName {
			if npAlbum != nil {
				return nil, fmt.Errorf("duplicate Nixplay albums named %s", albumName)
			}
			npAlbum = a
		}
	}
	if npAlbum == nil {
		return nil, fmt.Errorf("could not find Nixplay album %s", albumName)
	}
	return npAlbum, nil
}
