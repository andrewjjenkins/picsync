package nixplay

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/andrewjjenkins/nixplay/pkg/util"
	"github.com/satori/go.uuid"
)

// CoverURL is media shown as the cover for a slideshow
type CoverURL struct {
	URL         string `json:"url"`
	Orientation int    `json:"orientation"`
	Rotation    int    `json:"rotation"`
	PhotoID     int    `json:"photoId"`
}

// Slide is one slide in a slideshow
type Slide struct {
	UniqueHTMLID string `json:"uniqueHtmlId"`
	Type         string `json:"type"`
	Orientation  int    `json:"orientation"`
	Rotation     int    `json:"rotation"`
	Lookup       string `json:"lookup"`
	Duration     int    `json:"duration"`
	Caption      string `json:"caption"`
	PhotoID      int    `json:"photoId"`

	// SlideID and PlaylistItemID only relevant for patches to playlists
	SlideID        string `json:"slideId,omitempty"`
	PlaylistItemID string `json:"playlistItemId,omitempty"`
}

// SlideshowFacebookPhoto is media from Facebook (or other 3rd parties?) that
// can be referenced by a slide
//
// This is not implemented yet.
type SlideshowFacebookPhoto struct {
}

// SlideshowNixPhoto is media that can be referenced by a slideshow slide
type SlideshowNixPhoto struct {
	Lookup      string `json:"lookup"`
	Orientation int    `json:"orientation"`
	Rotation    int    `json:"rotation"`
	S3Key       string `json:"s3key"`
	ID          int    `json:"id"`
}

type publishSlideshowPhotos struct {
	FacebookPhoto []*SlideshowFacebookPhoto `json:"facebookPhoto"`
	NixPhoto      []*SlideshowNixPhoto      `json:"nixPhoto"`
}

type publishSlideshowData struct {
	SsID       int                    `json:"ssId"`
	Name       string                 `json:"name"`
	LastUpdate int64                  `json:"lastUpdate"`
	Version    int                    `json:"version"`
	Slides     []*Slide               `json:"slides"`
	Photos     publishSlideshowPhotos `json:"photos"`
	Transition string                 `json:"transition"`
	Duration   int                    `json:"duration"`
}

// PublishSlideshowBody is the data sent to publish photos to a slideshow
type PublishSlideshowBody struct {
	SlideshowID  int                  `json:"slideshow_id"`
	PictureCount int                  `json:"picture_count"`
	Version      int                  `json:"version"`
	CoverUrls    []*CoverURL          `json:"cover_urls"`
	Data         publishSlideshowData `json:"data"`
}

func photosToSlides(photos []*Photo) ([]*SlideshowNixPhoto, []*Slide) {
	r1 := []*SlideshowNixPhoto{}
	r2 := []*Slide{}
	for _, p := range photos {
		r1 = append(r1, &SlideshowNixPhoto{
			// Lookup probably works with anything unique but the web app uses md5.
			Lookup:      p.Md5,
			Orientation: p.Orientation,
			Rotation:    p.Rotation,
			S3Key:       p.S3Filename,
			ID:          p.ID,
		})
		u2 := uuid.NewV4()
		r2 = append(r2, &Slide{
			UniqueHTMLID: u2.String(),
			Type:         "nixPhoto",
			Orientation:  p.Orientation,
			Rotation:     p.Rotation,
			Lookup:       p.Md5,
			Duration:     -1,
			Caption:      p.Caption,
			PhotoID:      p.ID,
		})
	}
	return r1, r2
}

type OnFrame struct {
	Pk           int    `json:"pk"`
	Name         string `json:"name"`
	SerialNumber string `json:"serial_number"` // a longish int as a string
}

// Slideshow is the metadata about a slideshow
type Slideshow struct {
	UploadKey    string    `json:"upload_key"`
	Version      int       `json:"version"`
	ConfigFile   string    `json:"config_file"`
	Name         string    `json:"name"`
	PictureCount int       `json:"picture_count"`
	OnFrames     []OnFrame `json:"on_frames"`
	CoverUrls    string    `json:"cover_urls"` // A string of empty array, like "[]"?
	ID           int       `json:"id"`
}

// CreateSlideshow will create a new empty slideshow in the nixplay service
func CreateSlideshow(c *http.Client, name string) (*Slideshow, error) {
	bodyVals := url.Values{
		"name": {name},
	}
	ss := Slideshow{}
	url := "https://api.nixplay.com/v2/slideshow/add/"
	resp, err := doPost(c, url, &bodyVals)
	if err != nil {
		return nil, err
	}
	err = util.UnmarshalJSON(resp, &ss)
	return &ss, err
}

// GetSlideshows gets all configured slideshows for this account
func GetSlideshows(c *http.Client) ([]*Slideshow, error) {
	shows := []*Slideshow{}
	err := util.GetUnmarshalJSON(c, "https://api.nixplay.com/slideshow/list/json/", &shows)
	return shows, err
}

// GetSlideshowByName gets a particular slideshow by the name
//
// Slideshow names are not guaranteed unique - if you have defined multiple
// slideshows with the same name, then the first one found will be returned.
func GetSlideshowByName(c *http.Client, name string) (*Slideshow, error) {
	shows, err := GetSlideshows(c)
	if err != nil {
		return nil, err
	}
	for _, show := range shows {
		if show.Name == name {
			return show, nil
		}
	}
	return nil, fmt.Errorf("Did not find slideshow \"%s\" in %d slideshows", name, len(shows))
}

// PublishSlideshow publishes an array of photos as a simple slideshow
func PublishSlideshow(c *http.Client, show *Slideshow, photos []*Photo) error {
	nixPhotos, slides := photosToSlides(photos)
	lastUpdate := time.Now().Unix()

	data := publishSlideshowData{
		SsID:       show.ID,
		Name:       show.Name,
		LastUpdate: lastUpdate,
		Version:    0,
		Slides:     slides,
		Photos: publishSlideshowPhotos{
			FacebookPhoto: []*SlideshowFacebookPhoto{},
			NixPhoto:      nixPhotos,
		},
		Transition: "",
		Duration:   -1,
	}
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	dataStr := string(dataBytes)

	bodyVals := url.Values{
		"slideshow_id":  {fmt.Sprintf("%d", show.ID)},
		"picture_count": {fmt.Sprintf("%d", len(slides))},
		"version":       {"0"},
		"cover_urls":    {"[]"},
		"data":          {dataStr},
	}

	resp, err := doPost(c, "https://api.nixplay.com/slideshow/publish/", &bodyVals)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("Couldn't publish slideshow (%d): %v", resp.StatusCode, err)
	}
	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return nil
}
