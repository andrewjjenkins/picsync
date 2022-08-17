package nixplay

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/andrewjjenkins/picsync/pkg/util"
	uuid "github.com/satori/go.uuid"
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

type Playlist struct {
	Id              int    `json:"id"`
	Name            string `json:"name"`
	PlaylistName    string `json:"playlist_name"`
	CoverUrls       string `json:"cover_urls"` // of a JSON, weird.
	CreatedDate     string `json:"created_date"`
	ConfigFile      string `json:"config_file"`
	Converted       bool   `json:"converted"`
	Duration        int    `json:"duration"` // often -1?
	LastUpdatedDate string `json:"last_updated_date"`
	PictureCount    int    `json:"picture_count"`
	Type            string `json:"type"`
	Version         int    `json:"version"`

	//OnFrames
	//OnScheduledFrames
	//Sharing
	//Transition
	//UploadKey
}

type createPlaylistData struct {
	Name string `json:"name"`
}

type createPlaylistResponseData struct {
	PlaylistId int `json:"playlistId"`
}

func CreatePlaylist(c *http.Client, name string) (int, error) {
	body, err := json.Marshal(createPlaylistData{
		Name: name,
	})
	if err != nil {
		return -1, err
	}
	u := "https://api.nixplay.com/v3/playlists"
	req, err := http.NewRequest("POST", u, bytes.NewBuffer(body))
	if err != nil {
		return -1, err
	}
	res, err := doNixplayCsrf(c, req)
	if err != nil {
		return -1, err
	}
	defer res.Body.Close()
	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return -1, err
	}
	if res.StatusCode != http.StatusOK {
		return -1, fmt.Errorf("couldn't create playlist %s: http %d: %s", name,
			res.StatusCode, resBody)
	}

	var resData createPlaylistResponseData
	err = json.NewDecoder(res.Body).Decode(&resData)
	if err != nil {
		return -1, err
	}
	return resData.PlaylistId, nil
}

// GetPlaylists gets all configured slideshows for this account
func GetPlaylists(c *http.Client) ([]*Playlist, error) {
	req, err := http.NewRequest("GET", "https://api.nixplay.com/v3/playlists", nil)
	req.Header.Set("accept", "application/json")
	if err != nil {
		return nil, err
	}
	res, err := doNixplayCsrf(c, req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		resBody, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("couldn't get playlists: http %d: %s", res.StatusCode, resBody)
	}
	var playlists []*Playlist
	err = json.NewDecoder(res.Body).Decode(&playlists)
	if err != nil {
		return nil, err
	}
	return playlists, err
}

// GetPlaylistByName gets a particular slideshow by the name
//
// Playlist names are not guaranteed unique - if you have defined multiple
// playlists with the same name, then the first one found will be returned.
func GetPlaylistByName(c *http.Client, name string) (*Playlist, error) {
	playlists, err := GetPlaylists(c)
	if err != nil {
		return nil, err
	}
	for _, playlist := range playlists {
		if playlist.Name == name {
			return playlist, nil
		}
	}
	return nil, fmt.Errorf("did not find playlist \"%s\" in %d playlists", name, len(playlists))
}

type publishPlaylistDataItem struct {
	PictureId int `json:"pictureId"`
}
type publishPlaylistData struct {
	Items []publishPlaylistDataItem `json:"items"`
}

func PublishPlaylist(c *http.Client, playlistId int, photos []*Photo) error {
	data := publishPlaylistData{}
	for _, p := range photos {
		data.Items = append(data.Items, publishPlaylistDataItem{
			PictureId: p.ID,
		})
	}
	body, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// First we have to delete all items (sigh)
	//
	// FIXME: I think there's an API for deleting individual items; we'd have
	// to repeat the computeWork approach.  As implemented, we're wastefully
	// clearing the playlist and then re-writing it; this is just metadata
	// (an array of photo IDs, just integers), so the waste isn't very big
	// (like deleting/uploading images).  As well, it only happens when an
	// image is actually changed.
	// If you don't delete all the items, then the POST just adds items again,
	// e.g. if your album has 384 pictures in it and you added one, then your
	// playlist ends up with 384+385 pictures in it (two copies of each old, and
	// one copy of the new), and this repeats until you hit the 2000-photo limit.
	u := fmt.Sprintf("https://api.nixplay.com/v3/playlists/%d/items", playlistId)
	req, err := http.NewRequest("DELETE", u, nil)
	if err != nil {
		return err
	}
	req.Header.Set("accept", "application/json")
	res, err := doNixplayCsrf(c, req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("couldn't delete items from playlist: http %d: %s", res.StatusCode, resBody)
	}

	// Now POST back the list of items for the playlist.
	req, err = http.NewRequest("POST", u, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("content-type", "application/json")
	req.Header.Set("accept", "application/json")
	res, err = doNixplayCsrf(c, req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	resBody, err = io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("couldn't publish playlist: http %d: %s", res.StatusCode, resBody)
	}

	// If we got 200 OK, we don't care about the body.
	return nil
}
