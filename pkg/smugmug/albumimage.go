package smugmug

import (
	"fmt"
	"net/http"
)

// AlbumImage is the metadata for an AlbumImage.
type AlbumImage struct {
	URI            string `json:"Uri"`
	URIDescription string `json:"UriDescription"`
	Altitude       int
	ArchivedMD5    string
	ArchivedSize   int
	ArchivedURI    string `json:"ArchivedUri"`
	CanBuy         bool
	CanEdit        bool
	Caption        string
	Collectable    bool
	Date           string
	FileName       string
	Format         string
	Hidden         bool
	ImageKey       string
	IsArchive      bool
	IsVideo        bool
	KeywordArray   []string
	Keywords       string
	LastUpdated    string
	Latitude       string
	Longitude      string
	Movable        bool
	OriginalHeight int
	OriginalWidth  int
	OriginalSize   int
	Processing     bool
	Protected      bool
	ThumbnailURL   string `json:"ThumbnailUrl"`
	Title          string
	UploadKey      string
	Watermark      string
	Watermarked    bool
	WebURI         string            `json:"WebUri"`
	URIs           map[string]URIRef `json:"Uris"`

	// FormattedValues - skip
}

type albumImagesResponse struct {
	Response struct {
		responseCommon
		AlbumImage []*AlbumImage
	}
	Code    int
	Message string
}

func (i AlbumImage) String() string {
	return fmt.Sprintf(
		"Image %s (%dx%d, %s, \"%s\"): %s",
		i.FileName, i.OriginalWidth, i.OriginalHeight, i.ImageKey,
		i.Caption, i.WebURI,
	)
}

// GetAlbumImages will get the metadata for all images in an album
func GetAlbumImages(c *http.Client, albumKey string) ([]*AlbumImage, error) {
	resp := albumImagesResponse{}
	url := fmt.Sprintf("https://api.smugmug.com/api/v2/album/%s!images", albumKey)
	err := GetUnmarshalJSON(c, url, &resp)
	return resp.Response.AlbumImage, err
}
