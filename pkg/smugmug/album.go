package smugmug

import (
	"fmt"
	"net/http"
)

// AlbumURI is the data for a particular album URI
type AlbumURI struct {
	URI            string `json:"Uri"`
	Locator        string
	LocatorType    string
	URIDescription string `json:"UriDescription"`
	EndpointType   string
}

// AlbumURIs are the URIs associated with an album
type AlbumURIs struct {
	AlbumComments       AlbumURI
	AlbumGeoMedia       AlbumURI
	AlbumHighlightImage AlbumURI
	AlbumImages         AlbumURI
	AlbumPopularMedia   AlbumURI
	AlbumPrices         AlbumURI
	AlbumShareURIs      AlbumURI `json:"AlbumShareUris"`
	Folder              AlbumURI
	HighlightImage      AlbumURI
	Node                AlbumURI
	NodeCoverImage      AlbumURI
	ParentFolders       AlbumURI
	User                AlbumURI
}

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
	URLName             string `json:"UrlName"`
	URLPath             string `json:"UrlPath"`
	WebURI              string `json:"WebUri"`
	URIs                AlbumURIs
}

type albumResponse struct {
	Response struct {
		URI            string `json:"Uri"`
		URIDescription string `json:"UriDescription"`
		DocURI         string `json:"DocUri"`
		EndpointType   string
		Locator        string
		LocatorType    string
		Album          Album
	}
	Code    int
	Message string
}

func (a Album) String() string {
	return fmt.Sprintf(
		"Album %s (%s, %d images, updated %s): %s",
		a.Title, a.NiceName, a.ImageCount, a.LastUpdated, a.WebURI,
	)
}

// GetAlbum will get the metadata for an album from the API
func GetAlbum(c *http.Client, id string) (*Album, error) {
	resp := albumResponse{}
	url := fmt.Sprintf("https://api.smugmug.com/api/v2/album/%s", id)
	err := GetUnmarshalJSON(c, url, &resp)
	return &resp.Response.Album, err
}
