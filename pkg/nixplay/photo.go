package nixplay

import (
	"fmt"
	"net/http"

	"github.com/andrewjjenkins/nixplay/pkg/util"
)

// Photo is a reference to a photo in an album
type Photo struct {
	OriginalURL     string `json:"original_url"`
	Orientation     int    `json:"orientation"`
	Caption         string `json:"caption"`
	URL             string `json:"url"`
	RotationAllowed bool   `json:"rotation_allowed"`
	Filename        string `json:"filename"` // Looks like the user's filename when uploaded
	SortDate        string `json:"sortDate"` // 20180731232531
	AlbumID         int    `json:"albumId"`
	S3Filename      string `json:"s3filename"`
	PreviewURL      string `json:"previewUrl"`
	Published       bool   `json:"published"`
	SourceID        string `json:"source_id"` // Literally "unused" for my examples
	Rotation        int    `json:"rotation"`
	ThumbnailURL    string `json:"thumbnailUrl"`
	ID              int    `json:"id"`
	Md5             string `json:"md5"`
}

// GetPhotos returns the photos in an album
func GetPhotos(c *http.Client, albumID int) ([]*Photo, error) {
	type getPhotosResponse struct {
		Photos []*Photo `json:"photos"`
	}
	photos := getPhotosResponse{}
	u := fmt.Sprintf("https://api.nixplay.com/album/%d/pictures/json", albumID)
	err := util.GetUnmarshalJSON(c, u, &photos)
	return photos.Photos, err
}
func (p Photo) String() string {
	return fmt.Sprintf(
		"%s (%d/%d): %s [%s]",
		p.Filename, p.AlbumID, p.ID, p.S3Filename, p.Caption,
	)
}
