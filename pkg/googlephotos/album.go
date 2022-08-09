package googlephotos

import (
	"fmt"
	"net/http"
)

type Album struct {
	Id                    string           `json:"id"`
	Title                 string           `json:"title"`
	ProductUrl            string           `json:"productUrl"`
	IsWriteable           bool             `json:"isWriteable"`
	MediaItemsCount       MaybeQuotedInt64 `json:"mediaItemsCount"`
	CoverPhotoBaseUrl     string           `json:"coverPhotoBaseUrl"`
	CoverPhotoMediaItemId string           `json:"coverPhotoMediaItemId"`

	// ShareInfo ShareInfo `json:"shareInfo"`
}

type albumsResponse struct {
	Albums        []*Album `json:"albums"`
	NextPageToken string   `json:"nextPageToken"`
}

// FIXME: Handle pagination
func ListAlbums(c *http.Client) ([]*Album, error) {
	resp := albumsResponse{}
	url := "https://photoslibrary.googleapis.com/v1/albums"
	err := GetUnmarshalJSON(c, url, &resp)
	return resp.Albums, err
}

type searchMediaItemsResponse struct {
	MediaItems    []*MediaItem `json:"mediaItems"`
	NextPageToken string       `json:"nextPageToken"`
}

// FIXME: Handle pagination
func ListMediaItemsForAlbumId(c *http.Client, albumId string) ([]*MediaItem, error) {
	resp := searchMediaItemsResponse{}
	url := "https://photoslibrary.googleapis.com/v1/mediaItems:search"
	body := fmt.Sprintf("{\"albumId\":\"%s\"}", albumId)
	err := PostUnmarshalJSON(c, url, body, resp)
	return resp.MediaItems, err
}
