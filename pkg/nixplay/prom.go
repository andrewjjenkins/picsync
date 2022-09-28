package nixplay

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type promImpl struct {
	promFactory promauto.Factory

	getAlbumsSuccess         prometheus.Counter
	getAlbumsFailure         prometheus.Counter
	getAlbumByNameSuccess    prometheus.Counter
	getAlbumByNameFailure    prometheus.Counter
	getPhotosSuccess         prometheus.Counter
	getPhotosFailure         prometheus.Counter
	getPhotosPhotoCount      prometheus.Counter
	uploadPhotoSuccess       prometheus.Counter
	uploadPhotoFailure       prometheus.Counter
	uploadPhotoTotalBytes    prometheus.Counter
	deletePhotoSuccess       prometheus.Counter
	deletePhotoFailure       prometheus.Counter
	createAlbumSuccess       prometheus.Counter
	createAlbumFailure       prometheus.Counter
	createPlaylistSuccess    prometheus.Counter
	createPlaylistFailure    prometheus.Counter
	getPlaylistsSuccess      prometheus.Counter
	getPlaylistsFailure      prometheus.Counter
	getPlaylistByNameSuccess prometheus.Counter
	getPlaylistByNameFailure prometheus.Counter
	publishPlaylistSuccess   prometheus.Counter
	publishPlaylistFailure   prometheus.Counter
}

func (c *clientImpl) promRegister(reg prometheus.Registerer) error {
	c.prom.promFactory = promauto.With(reg)

	c.prom.getAlbumsSuccess = c.prom.promFactory.NewCounter(
		prometheus.CounterOpts{
			Name: "nixplay_get_albums_success",
			Help: "Successful calls to list the user's albums",
		},
	)
	c.prom.getAlbumsFailure = c.prom.promFactory.NewCounter(
		prometheus.CounterOpts{
			Name: "nixplay_get_albums_failure",
			Help: "Failed calls to list the user's albums",
		},
	)
	c.prom.getAlbumByNameSuccess = c.prom.promFactory.NewCounter(
		prometheus.CounterOpts{
			Name: "nixplay_get_album_by_name_success",
			Help: "Successful calls to get an album by name",
		},
	)
	c.prom.getAlbumByNameFailure = c.prom.promFactory.NewCounter(
		prometheus.CounterOpts{
			Name: "nixplay_get_album_by_name_failure",
			Help: "Failed calls to get an album by name",
		},
	)
	c.prom.getPhotosSuccess = c.prom.promFactory.NewCounter(
		prometheus.CounterOpts{
			Name: "nixplay_get_photos_success",
			Help: "Successful calls to get photos in an album",
		},
	)
	c.prom.getPhotosFailure = c.prom.promFactory.NewCounter(
		prometheus.CounterOpts{
			Name: "nixplay_get_photos_failure",
			Help: "Failed calls to get photos in an album",
		},
	)
	c.prom.getPhotosPhotoCount = c.prom.promFactory.NewCounter(
		prometheus.CounterOpts{
			Name: "nixplay_get_photos_photo_count",
			Help: "Total count of photo metadata received",
		},
	)
	c.prom.deletePhotoSuccess = c.prom.promFactory.NewCounter(
		prometheus.CounterOpts{
			Name: "nixplay_delete_photo_success",
			Help: "Photos deleted successfully",
		},
	)
	c.prom.deletePhotoFailure = c.prom.promFactory.NewCounter(
		prometheus.CounterOpts{
			Name: "nixplay_delete_photo_failure",
			Help: "Failures when deleting photo",
		},
	)
	c.prom.uploadPhotoSuccess = c.prom.promFactory.NewCounter(
		prometheus.CounterOpts{
			Name: "nixplay_upload_photos_success",
			Help: "Successful uploads of photos",
		},
	)
	c.prom.uploadPhotoFailure = c.prom.promFactory.NewCounter(
		prometheus.CounterOpts{
			Name: "nixplay_upload_photos_failure",
			Help: "Failed uploads of photos",
		},
	)
	c.prom.uploadPhotoTotalBytes = c.prom.promFactory.NewCounter(
		prometheus.CounterOpts{
			Name: "nixplay_upload_photos_bytes",
			Help: "Total count of bytes of photos successfully uploaded",
		},
	)
	c.prom.createAlbumSuccess = c.prom.promFactory.NewCounter(
		prometheus.CounterOpts{
			Name: "nixplay_create_album_success",
			Help: "Successful creation of album",
		},
	)
	c.prom.createAlbumFailure = c.prom.promFactory.NewCounter(
		prometheus.CounterOpts{
			Name: "nixplay_create_album_failure",
			Help: "Failed creation of album",
		},
	)
	c.prom.createPlaylistSuccess = c.prom.promFactory.NewCounter(
		prometheus.CounterOpts{
			Name: "nixplay_create_playlist_success",
			Help: "Successful creation of playlist",
		},
	)
	c.prom.createPlaylistFailure = c.prom.promFactory.NewCounter(
		prometheus.CounterOpts{
			Name: "nixplay_create_playlist_failure",
			Help: "Failed creation of playlist",
		},
	)
	c.prom.getPlaylistsSuccess = c.prom.promFactory.NewCounter(
		prometheus.CounterOpts{
			Name: "nixplay_get_playlists_success",
			Help: "Successful calls to list the user's playlists",
		},
	)
	c.prom.getPlaylistsFailure = c.prom.promFactory.NewCounter(
		prometheus.CounterOpts{
			Name: "nixplay_get_playlists_failure",
			Help: "Failed calls to list the user's playlists",
		},
	)
	c.prom.getPlaylistByNameSuccess = c.prom.promFactory.NewCounter(
		prometheus.CounterOpts{
			Name: "nixplay_get_playlist_by_name_success",
			Help: "Successful calls to get an playlist by name",
		},
	)
	c.prom.getPlaylistByNameFailure = c.prom.promFactory.NewCounter(
		prometheus.CounterOpts{
			Name: "nixplay_get_playlist_by_name_failure",
			Help: "Failed calls to get an playlist by name",
		},
	)
	c.prom.publishPlaylistSuccess = c.prom.promFactory.NewCounter(
		prometheus.CounterOpts{
			Name: "nixplay_publish_playlist_success",
			Help: "Successful calls to publish a playlist",
		},
	)
	c.prom.publishPlaylistFailure = c.prom.promFactory.NewCounter(
		prometheus.CounterOpts{
			Name: "nixplay_publish_playlist_failure",
			Help: "Failed calls to publish a playlist",
		},
	)

	return nil
}
