package googlephotos

import (
	"context"
	"net/http"

	"github.com/andrewjjenkins/picsync/pkg/cache"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/oauth2"
)

type Client interface {
	ListAlbums() ([]*Album, error)
	ListSharedAlbums() ([]*Album, error)
	ListMediaItemsForAlbumId(albumId string, nextPageToken string) (*SearchMediaItemsResponse, error)
	UpdateCacheForAlbumId(albumId string, nextPageToken string, cb UpdateCacheCallback) (*UpdateCacheResult, error)
}

type clientImpl struct {
	httpClient *http.Client
	cache      cache.Cache

	prom promImpl
}

func NewClient(
	consumerKey string,
	consumerSecret string,
	ctx context.Context,
	t *oauth2.Token,
	c cache.Cache,
	reg prometheus.Registerer,
) Client {
	config := newOauth2Config(consumerKey, consumerSecret, "")

	gpClient := clientImpl{
		httpClient: config.Client(ctx, t),
		cache:      c,
	}

	gpClient.promRegister(reg)

	return &gpClient
}
