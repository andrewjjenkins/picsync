package smugmug

import (
	"fmt"
	"net/http"
)

// Album is the metadata for a SmugMug album
type Album struct {
	URI                 string `json:"Uri"`
	URIDescription      string `json:"UriDescription"`
	ResponseLevel       string
	AlbumKey            string
	AllowDownloads      bool
	CanShare            bool
	Description         string
	External            bool
	HasDownloadPassword bool
	ImageCount          int
	ImagesLastUpdated   string
	Keywords            string
	LastUpdated         string
	Name                string
	NiceName            string
	NodeID              string
	PasswordHint        string
	Protected           bool
	SecurityType        string
	SortDirection       string
	SortMethod          string
	Title               string
	URLName             string            `json:"UrlName"`
	URLPath             string            `json:"UrlPath"`
	WebURI              string            `json:"WebUri"`
	URIs                map[string]URIRef `json:"Uris"`
}

type albumResponse struct {
	Response struct {
		responseCommon
		Album Album
	}
	Code    int
	Message string
}

func (a Album) String() string {
	return fmt.Sprintf(
		"Album %s (%s, %s, %d images, updated %s): %s",
		a.Title, a.NodeID, a.NiceName, a.ImageCount, a.LastUpdated, a.WebURI,
	)
}

// GetAlbum will get the metadata for an album from the API
func GetAlbum(c *http.Client, id string) (*Album, error) {
	resp := albumResponse{}
	url := fmt.Sprintf("https://api.smugmug.com/api/v2/album/%s", id)
	err := GetUnmarshalJSON(c, url, &resp)
	return &resp.Response.Album, err
}
