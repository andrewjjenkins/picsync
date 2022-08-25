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
	httpClient  *http.Client
	tokenSource oauth2.TokenSource
	cache       cache.Cache

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
	tokenSource := config.TokenSource(ctx, t)
	httpClient := oauth2.NewClient(ctx, tokenSource)

	gpClient := clientImpl{
		httpClient:  httpClient,
		tokenSource: tokenSource,
		cache:       c,
	}

	gpClient.promRegister(reg)

	return &gpClient
}
