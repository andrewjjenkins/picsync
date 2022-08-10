package googlephotos

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/andrewjjenkins/picsync/pkg/cache"
)

type MediaItem struct {
	Id          string `json:"id"`
	Description string `json:"description"`
	ProductUrl  string `json:"productUrl"`
	BaseUrl     string `json:"baseUrl"`
	MimeType    string `json:"mimeType"`
	Filename    string `json:"filename"`

	// MediaMetadata MediaMetadata `json:"mediaMetadata"`
	// ContributorInfo ContributorInfo `json:"contributorInfo"`
}

type CachedMediaItem struct {
	CacheId     int64
	Sha256      string
	Md5         string
	LastUpdated time.Time
	LastUsed    time.Time
	MediaItem   *MediaItem
}

type UpdateCacheCallback func(*CachedMediaItem)

func UpdateCacheForAlbumId(client *http.Client, c cache.Cache, albumId string, cb UpdateCacheCallback) ([]*CachedMediaItem, error) {
	mediaItems, err := ListMediaItemsForAlbumId(client, albumId)
	toRet := []*CachedMediaItem{}
	if err != nil {
		return []*CachedMediaItem{}, err
	}
	for _, item := range mediaItems {
		resp, err := http.Get(item.BaseUrl)
		if err != nil {
			// FIXME: Maybe we want to skip updating cache for this item if we
			// just have a download error rather than failing the entire call?
			return []*CachedMediaItem{}, err
		}
		if resp.StatusCode != http.StatusOK {
			return []*CachedMediaItem{}, fmt.Errorf("received HTTP %d", resp.StatusCode)
		}
		sha256Hash := sha256.New()
		md5Hash := md5.New()
		allHashes := io.MultiWriter(sha256Hash, md5Hash)
		if _, err := io.Copy(allHashes, resp.Body); err != nil {
			return []*CachedMediaItem{}, err
		}
		entry := cache.GooglephotoData{
			BaseUrl:     item.BaseUrl,
			Sha256:      hex.EncodeToString(sha256Hash.Sum(nil)),
			Md5:         hex.EncodeToString(md5Hash.Sum(nil)),
			LastUpdated: time.Now(),
			LastUsed:    time.Now(),
		}
		err = c.UpsertGooglephoto(&entry)
		if err != nil {
			// FIXME: Again, maybe just skip individual errors?
			return []*CachedMediaItem{}, err
		}
		cached := CachedMediaItem{
			CacheId:     entry.Id,
			Sha256:      entry.Sha256,
			Md5:         entry.Md5,
			LastUpdated: entry.LastUpdated,
			LastUsed:    entry.LastUsed,
			MediaItem:   item,
		}
		toRet = append(toRet, &cached)
		cb(&cached)
	}
	return toRet, nil
}
