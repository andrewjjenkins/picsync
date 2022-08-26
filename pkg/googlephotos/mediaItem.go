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
	Id            string        `json:"id"`
	Description   string        `json:"description"`
	ProductUrl    string        `json:"productUrl"`
	BaseUrl       string        `json:"baseUrl"`
	MimeType      string        `json:"mimeType"`
	Filename      string        `json:"filename"`
	MediaMetadata MediaMetadata `json:"mediaMetadata"`

	// ContributorInfo ContributorInfo `json:"contributorInfo"`
}

type MediaMetadata struct {
	CreationTime string           `json:"creationTime"`
	Width        MaybeQuotedInt64 `json:"width"`
	Height       MaybeQuotedInt64 `json:"height"`

	// Either Photo or Video will be present
	Photo *PhotoMediaMetadata `json:"photo"`
	Video *VideoMediaMetadata `json:"video"`
}

type PhotoMediaMetadata struct {
	CameraMake      string           `json:"cameraMake"`
	CameraModel     string           `json:"cameraModel"`
	FocalLength     float64          `json:"focalLength"`
	ApertureFNumber float64          `json:"apertureFNumber"`
	IsoEquivalent   MaybeQuotedInt64 `json:"isoEquivalent"`
	ExposureTime    string           `json:"exposureTime"`
}

type VideoMediaMetadata struct {
	CameraMake  string  `json:"cameraMake"`
	CameraModel string  `json:"cameraModel"`
	Fps         float64 `json:"fps"`
	Status      string  `json:"status"`
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

func (c *clientImpl) UpdateCacheForAlbumId(albumId string, nextPageToken string, cb UpdateCacheCallback) (*UpdateCacheResult, error) {
	res, err := c.ListMediaItemsForAlbumId(albumId, nextPageToken)
	toRet := &UpdateCacheResult{}
	if err != nil {
		return nil, err
	}
	toRet.NextPageToken = res.NextPageToken
	for _, item := range res.MediaItems {
		// First, see if it is already in the cache.  Google never changes
		// the contents of a Google Photos ID, so if it is already present we don't
		// need to download it again.
		currentEntry, err := c.cache.GetGooglephoto(item.Id)
		if err != nil {
			return nil, err
		}
		if currentEntry != nil {
			// Update the timestamps and set back to the cache.
			currentEntry.LastUpdated = time.Now()
			currentEntry.LastUsed = currentEntry.LastUpdated
			err = c.cache.UpsertGooglephoto(currentEntry)
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
		fullResUrl := item.BaseUrl + "=d"
		resp, err := http.Get(fullResUrl)
		if err != nil {
			// FIXME: Maybe we want to skip updating cache for this item if we
			// just have a download error rather than failing the entire call?
			c.prom.mediaItemsDownloadedFailure.Inc()
			return nil, err
		}
		if resp.StatusCode != http.StatusOK {
			c.prom.mediaItemsDownloadedFailure.Inc()
			return nil, fmt.Errorf("received HTTP %d", resp.StatusCode)
		}
		sha256Hash := sha256.New()
		md5Hash := md5.New()
		allHashes := io.MultiWriter(sha256Hash, md5Hash)
		if _, err := io.Copy(allHashes, resp.Body); err != nil {
			c.prom.mediaItemsDownloadedFailure.Inc()
			return nil, err
		}
		c.prom.mediaItemsDownloadedSuccess.Inc()

		// FIXME: resp.ContentLength may in theory be unknown, but is known
		// for google photos.  A safer approach would be to make another
		// member of the io.Multiwriter() that just counted bytes and threw them
		// on the ground, and then ask it how many bytes we saw.
		if resp.ContentLength > 0 {
			c.prom.mediaItemsDownloadedBytes.Add(float64(resp.ContentLength))
		}

		entry := cache.GooglephotoData{
			BaseUrl:        item.BaseUrl,
			GooglephotosId: item.Id,
			Sha256:         hex.EncodeToString(sha256Hash.Sum(nil)),
			Md5:            hex.EncodeToString(md5Hash.Sum(nil)),
			Width:          int64(item.MediaMetadata.Width),
			Height:         int64(item.MediaMetadata.Height),
			LastUpdated:    time.Now(),
			LastUsed:       time.Now(),
		}
		err = c.cache.UpsertGooglephoto(&entry)
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
