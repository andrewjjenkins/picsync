package googlephotos

import (
	"fmt"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type promImpl struct {
	promFactory promauto.Factory

	listAlbumsSuccess           prometheus.Counter
	listAlbumsFailure           prometheus.Counter
	listMediaItemsSuccess       prometheus.Counter
	listMediaItemsFailure       prometheus.Counter
	listMediaItemsCount         prometheus.Counter
	mediaItemsDownloadedSuccess prometheus.Counter
	mediaItemsDownloadedFailure prometheus.Counter
	mediaItemsDownloadedBytes   prometheus.Counter
	tokenExpiry                 prometheus.GaugeFunc
}

func (c *clientImpl) promRegister(reg prometheus.Registerer) error {
	c.prom.promFactory = promauto.With(reg)

	c.prom.listAlbumsSuccess = c.prom.promFactory.NewCounter(
		prometheus.CounterOpts{
			Name: "googlephotos_list_albums_success",
			Help: "Successful calls to list the user's albums",
		})
	c.prom.listAlbumsFailure = c.prom.promFactory.NewCounter(
		prometheus.CounterOpts{
			Name: "googlephotos_list_albums_failure",
			Help: "Failed calls to list the user's albums",
		})
	c.prom.listMediaItemsSuccess = c.prom.promFactory.NewCounter(
		prometheus.CounterOpts{
			Name: "googlephotos_list_mediaitems_success",
			Help: "Successful calls to list all media items in an album",
		})
	c.prom.listMediaItemsFailure = c.prom.promFactory.NewCounter(
		prometheus.CounterOpts{
			Name: "googlephotos_list_mediaitems_failure",
			Help: "Failed calls to list all media items in an album",
		})
	c.prom.listMediaItemsCount = c.prom.promFactory.NewCounter(
		prometheus.CounterOpts{
			Name: "googlephotos_list_mediaitems_count",
			Help: "Total number of media items from all calls to list",
		})
	c.prom.mediaItemsDownloadedSuccess = c.prom.promFactory.NewCounter(
		prometheus.CounterOpts{
			Name: "googlephotos_mediaitems_downloaded_success",
			Help: "Number of media items that were downloaded",
		})
	c.prom.mediaItemsDownloadedFailure = c.prom.promFactory.NewCounter(
		prometheus.CounterOpts{
			Name: "googlephotos_mediaitems_downloaded_failure",
			Help: "Number of media items that were downloaded but encountered an error",
		})
	c.prom.mediaItemsDownloadedBytes = c.prom.promFactory.NewCounter(
		prometheus.CounterOpts{
			Name: "googlephotos_mediaitems_downloaded_bytes",
			Help: "Total bytes of all media item downloads",
		})

	expiryGetter, err := c.newTokenExpiryGetter()
	if err != nil {
		return err
	}
	c.prom.tokenExpiry = c.prom.promFactory.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "googlephotos_access_token_expiry",
		Help: "Time that the current access token expires (unix timestamp)",
	}, expiryGetter)
	return nil
}

func (c *clientImpl) newTokenExpiryGetter() (func() float64, error) {
	var mu sync.Mutex
	lastTokenFetchTime := time.Now()
	lastToken, err := c.tokenSource.Token()
	if err != nil {
		return nil, err
	}

	tokenFetcher := func() float64 {
		mu.Lock()
		defer mu.Unlock()
		since := time.Now().Sub(lastTokenFetchTime)
		if since < 5*time.Minute {
			// return cached expiry time to avoid triggering too many refreshes
			// just for stats.  Unlikely to happen in the case that refreshes
			// are actually successful, but if refreshes fail, this limits how
			// often we try to refresh.
			return float64(lastToken.Expiry.Unix())
		}
		t, err := c.tokenSource.Token()
		if err != nil {
			fmt.Printf("Warning: failed to get a token: %s\n", err.Error())
			return float64(lastToken.Expiry.Unix())
		}
		lastToken = t
		lastTokenFetchTime = time.Now()
		return float64(lastToken.Expiry.Unix())
	}
	return tokenFetcher, nil
}
