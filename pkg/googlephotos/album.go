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

type sharedAlbumsResponse struct {
	SharedAlbums  []*Album `json:"sharedAlbums"`
	NextPageToken string   `json:"nextPageToken"`
}

func ListSharedAlbums(c *http.Client) ([]*Album, error) {
	resp := sharedAlbumsResponse{}
	url := "https://photoslibrary.googleapis.com/v1/sharedAlbums"
	err := GetUnmarshalJSON(c, url, &resp)
	return resp.SharedAlbums, err
}

type SearchMediaItemsResponse struct {
	MediaItems    []*MediaItem `json:"mediaItems"`
	NextPageToken string       `json:"nextPageToken"`
}

func ListMediaItemsForAlbumId(c *http.Client, albumId string, nextPageToken string) (*SearchMediaItemsResponse, error) {
	resp := SearchMediaItemsResponse{}
	url := "https://photoslibrary.googleapis.com/v1/mediaItems:search"
	var body string
	if nextPageToken == "" {
		body = fmt.Sprintf("{\"albumId\":\"%s\"}", albumId)
	} else {
		body = fmt.Sprintf("{\"albumId\":\"%s\",\"pageToken\":\"%s\"}", albumId, nextPageToken)
	}
	err := PostUnmarshalJSON(c, url, body, &resp)
	return &resp, err
}
