package nixplay

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

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
