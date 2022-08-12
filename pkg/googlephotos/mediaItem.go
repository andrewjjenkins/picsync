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

type UpdateCacheResult struct {
	CachedMediaItems []*CachedMediaItem
	NextPageToken    string
}

type UpdateCacheCallback func(*CachedMediaItem)

func UpdateCacheForAlbumId(client *http.Client, c cache.Cache, albumId string, nextPageToken string, cb UpdateCacheCallback) (*UpdateCacheResult, error) {
	res, err := ListMediaItemsForAlbumId(client, albumId, nextPageToken)
	toRet := &UpdateCacheResult{}
	if err != nil {
		return nil, err
	}
	toRet.NextPageToken = res.NextPageToken
	for _, item := range res.MediaItems {
		// First, see if it is already in the cache.  Google never changes
		// the contents of a baseUrl, so if it is already present we don't
		// need to download it again.
		currentEntry, err := c.GetGooglephoto(item.BaseUrl)
		if err != nil {
			return nil, err
		}
		if currentEntry != nil {
			// Update the timestamps and set back to the cache.
			currentEntry.LastUpdated = time.Now()
			currentEntry.LastUsed = currentEntry.LastUpdated
			err = c.UpsertGooglephoto(currentEntry)
			if err != nil {
				return nil, err
			}
			cached := CachedMediaItem{
				CacheId:     currentEntry.Id,
				Sha256:      currentEntry.Sha256,
				Md5:         currentEntry.Md5,
				LastUpdated: currentEntry.LastUpdated,
				LastUsed:    currentEntry.LastUsed,
				MediaItem:   item,
			}
			toRet.CachedMediaItems = append(toRet.CachedMediaItems, &cached)
			cb(&cached)
			continue
		}

		// Item not in the cache.  We must download it and calculate hashes.
		resp, err := http.Get(item.BaseUrl)
		if err != nil {
			// FIXME: Maybe we want to skip updating cache for this item if we
			// just have a download error rather than failing the entire call?
			return nil, err
		}
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("received HTTP %d", resp.StatusCode)
		}
		sha256Hash := sha256.New()
		md5Hash := md5.New()
		allHashes := io.MultiWriter(sha256Hash, md5Hash)
		if _, err := io.Copy(allHashes, resp.Body); err != nil {
			return nil, err
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
			return nil, err
		}
		cached := CachedMediaItem{
			CacheId:     entry.Id,
			Sha256:      entry.Sha256,
			Md5:         entry.Md5,
			LastUpdated: entry.LastUpdated,
			LastUsed:    entry.LastUsed,
			MediaItem:   item,
		}
		toRet.CachedMediaItems = append(toRet.CachedMediaItems, &cached)
		cb(&cached)
	}
	return toRet, nil
}
