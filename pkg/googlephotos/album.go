package googlephotos

import (
	"fmt"
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
func (c clientImpl) ListAlbums() ([]*Album, error) {
	resp := albumsResponse{}
	url := "https://photoslibrary.googleapis.com/v1/albums"
	err := GetUnmarshalJSON(c.httpClient, url, &resp)
	return resp.Albums, err
}

type sharedAlbumsResponse struct {
	SharedAlbums  []*Album `json:"sharedAlbums"`
	NextPageToken string   `json:"nextPageToken"`
}

func (c clientImpl) ListSharedAlbums() ([]*Album, error) {
	resp := sharedAlbumsResponse{}
	url := "https://photoslibrary.googleapis.com/v1/sharedAlbums"
	err := GetUnmarshalJSON(c.httpClient, url, &resp)
	return resp.SharedAlbums, err
}

type SearchMediaItemsResponse struct {
	MediaItems    []*MediaItem `json:"mediaItems"`
	NextPageToken string       `json:"nextPageToken"`
}

func (c clientImpl) ListMediaItemsForAlbumId(albumId string, nextPageToken string) (*SearchMediaItemsResponse, error) {
	resp := SearchMediaItemsResponse{}
	url := "https://photoslibrary.googleapis.com/v1/mediaItems:search"
	var body string
	if nextPageToken == "" {
		body = fmt.Sprintf("{\"albumId\":\"%s\"}", albumId)
	} else {
		body = fmt.Sprintf("{\"albumId\":\"%s\",\"pageToken\":\"%s\"}", albumId, nextPageToken)
	}
	err := PostUnmarshalJSON(c.httpClient, url, body, &resp)
	return &resp, err
}
