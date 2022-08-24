package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/andrewjjenkins/picsync/pkg/cache"
	"github.com/andrewjjenkins/picsync/pkg/googlephotos"
	"github.com/andrewjjenkins/picsync/pkg/nixplay"
	"github.com/andrewjjenkins/picsync/pkg/util"
	"github.com/robfig/cron"
	"github.com/spf13/cobra"
)

var (
	syncCmd = &cobra.Command{
		Use:   "sync [<picsync.yaml>]",
		Short: "Sync pictures to Nixplay as specified in config file",
		Run:   runSync,
	}
)

func init() {
	rootCmd.AddCommand(syncCmd)
}

type syncClients struct {
	googlephotos googlephotos.Client
	nixplay      *http.Client
	cache        cache.Cache
}

func runSync(cmd *cobra.Command, args []string) {
	if len(args) > 1 {
		panic(fmt.Errorf("pass the path to one picsync.yaml config file"))
	}
	configFile := "picsync.yaml"
	if len(args) == 1 {
		configFile = args[0]
	}
	config, err := util.LoadConfig(configFile)
	if err != nil {
		panic(err)
	}

	// Start prometheus
	if config.Every != "" {
		promInitOrDie(config.Prometheus.Listen)
	}

	clients := syncClients{}

	// Create the cache up here so we can pass it down, this avoids
	// re-creating the cache (opening/closing Sqlite db) every run
	// and simplifies prometheus (which doesn't want the metrics re-registered)
	clients.cache, err = cache.New(promReg)
	if err != nil {
		panic(err)
	}

	// Log in to services; exit early if there's an auth problem
	clients.googlephotos = getGooglephotoClientOrExit(clients.cache)
	clients.nixplay = getNixplayClientOrExit()

	if config.Every != "" {
		runSyncGooglephotosEvery(clients, config.Albums, config.Every)
	} else {
		runSyncGooglephotosOnce(clients, config.Albums)
	}
}

func runSyncGooglephotosOnce(clients syncClients, albums []*util.ConfigAlbum) {
	for _, album := range albums {
		err := doSyncGooglephotos(clients, album)
		if err != nil {
			fmt.Printf("%v\n", err)
			os.Exit(1)
		}
	}
	os.Exit(0)
}

func runSyncGooglephotosEvery(clients syncClients, albums []*util.ConfigAlbum, every string) {
	everyCronSpec := fmt.Sprintf("@every %s", every)
	job := func() {
		for _, album := range albums {
			err := doSyncGooglephotos(clients, album)
			if err != nil {
				fmt.Printf("Error syncing album %s: %v\n", album.Name, err)
			}
		}
		fmt.Printf("%s: Sync of %d albums complete\n\n",
			time.Now().String(), len(albums))
	}

	c := cron.New()
	err := c.AddFunc(everyCronSpec, job)
	if err != nil {
		fmt.Printf("Cannot run every %s: %v\n", every, err)
		os.Exit(1)
	}
	fmt.Printf("Syncing every %s\n", every)

	// Run it once first so that we don't sleep at the beginning
	job()

	c.Run()
}

func doSyncGooglephotos(clients syncClients, album *util.ConfigAlbum) error {
	sourceAlbums := album.Sources.Googlephotos

	if len(sourceAlbums) == 0 {
		fmt.Printf("No source album. Cowardly refusing to delete all destination photos.\n")
		return nil
	}

	var sourceCacheImages []*googlephotos.CachedMediaItem
	for i, sourceAlbumId := range sourceAlbums {
		var sourceCacheUpdateCount int
		sourceCacheUpdateCb := func(cached *googlephotos.CachedMediaItem) {
			sourceCacheUpdateCount++
			fmt.Fprintf(os.Stdout, "\033[2K\rRefreshing source image %d...", sourceCacheUpdateCount)
		}

		var nextPageToken string
		for ok := true; ok; ok = (nextPageToken != "") {
			res, err := clients.googlephotos.UpdateCacheForAlbumId(
				sourceAlbumId, nextPageToken, sourceCacheUpdateCb)
			if err != nil {
				return err
			}
			nextPageToken = res.NextPageToken
			sourceCacheImages = append(sourceCacheImages, res.CachedMediaItems...)
		}
		fmt.Fprintf(os.Stdout, "\033[2K\rRefreshed %d source images for album %d/%d\n",
			sourceCacheUpdateCount, i+1, len(sourceAlbums))
	}

	// Get the nixplay image metadata for the requested album
	npAlbum, err := nixplay.GetAlbumByName(clients.nixplay, album.Name)
	if err != nil {
		return err
	}
	npPhotos, err := nixplay.GetPhotos(clients.nixplay, npAlbum.ID)
	if err != nil {
		return err
	}

	work, err := calcSyncGooglephotosWork(sourceCacheImages, npPhotos)
	if err != nil {
		return err
	}
	fmt.Printf("Sync work:\n")
	fmt.Printf("  To upload: %d\n", len(work.ToUpload))
	fmt.Printf("  To delete: %d\n", len(work.ToDelete))

	if album.DryRun != nil && *album.DryRun {
		return nil
	}

	for i, up := range work.ToUpload {
		fmt.Fprintf(os.Stdout, "\033[2K\rUploading image %d/%d...", i+1, len(work.ToUpload))
		err := uploadGooglephotoToNixplay(up, npAlbum.ID, clients.nixplay)
		if err != nil {
			fmt.Printf("\nError uploading photo %s (skipping): %v\n", up.MediaItem.Filename, err)
		}
	}
	if len(work.ToUpload) > 0 {
		fmt.Printf("DONE.  Uploading complete.\n")
	}

	if len(work.ToUpload) > 0 {
		fmt.Printf("Sleeping for 5 seconds to let nixplay digest uploaded photos...\n")
		time.Sleep(5 * time.Second)
	}

	for i, del := range work.ToDelete {
		fmt.Fprintf(os.Stdout, "\033[2K\rDeleting image %d/%d...", i+1, len(work.ToDelete))
		err := deleteGooglephotoFromNixplay(del, clients.nixplay)
		if err != nil {
			fmt.Printf("\nError deleting photo %s (skipping): %v\n", del.Filename, err)
		}
	}
	if len(work.ToDelete) > 0 {
		fmt.Printf("DONE.  Deleting complete.\n")
	}

	// FIXME: This should be commonized
	// Now, get the photos again and put them in a playlist
	npPhotos, err = nixplay.GetPhotos(clients.nixplay, npAlbum.ID)
	if err != nil {
		return err
	}
	plName := fmt.Sprintf("ss_%s", album.Name)
	pl, err := nixplay.GetPlaylistByName(clients.nixplay, plName)
	var playlistId int
	neededCreate := false
	if err == nil {
		playlistId = pl.Id
	} else {
		fmt.Printf("Could not find playlist %s (%v), creating\n", plName, err)
		fmt.Printf(
			"If this works, you must then assign the playlist %s to frames - "+
				"this program will not do that (but it will update the playlist once "+
				"you've assigned it)\n",
			plName,
		)
		neededCreate = true
		playlistId, err = nixplay.CreatePlaylist(clients.nixplay, plName)
		if err != nil {
			return err
		}
	}

	forcePublish := false
	if album.ForcePublish != nil {
		forcePublish = *album.ForcePublish
	}
	if len(work.ToUpload) > 0 || len(work.ToDelete) > 0 || neededCreate || forcePublish {
		err = nixplay.PublishPlaylist(clients.nixplay, playlistId, npPhotos)
		if err != nil {
			return err
		}
		fmt.Printf("Published %d photos to playlist %s\n", len(npPhotos), plName)
	} else {
		fmt.Printf(
			"No changes required for slideshow %s (%d photos)\n",
			plName,
			len(npPhotos),
		)
	}
	return nil
}

type syncGooglephotosWork struct {
	ToUpload []*googlephotos.CachedMediaItem
	ToDelete []*nixplay.Photo
}
type nixplayAlbumImagesByMd5 map[string]*nixplay.Photo

func calcSyncGooglephotosWork(sourceImgs []*googlephotos.CachedMediaItem, destImgs []*nixplay.Photo) (*syncGooglephotosWork, error) {
	work := syncGooglephotosWork{}

	// Create a lookup-by-md5 for all the images already in the destination album
	targetMd5s := make(nixplayAlbumImagesByMd5)
	for _, img := range destImgs {
		alreadyThere, ok := targetMd5s[img.Md5]
		if ok {
			fmt.Printf(
				"Warning: duplicate images with MD5 %s (%s, %s)\n",
				img.Md5, alreadyThere.Filename, img.Filename,
			)
			continue
		}
		targetMd5s[img.Md5] = img
	}

	// For each source image, find if it is already in the destination.
	for _, img := range sourceImgs {
		_, ok := targetMd5s[img.Md5]
		if !ok {
			work.ToUpload = append(work.ToUpload, img)
			continue
		}

		// If it is present, delete it from targetMd5s so it won't count toward
		// toDelete.  This only works if there are no duplicates in source.
		delete(targetMd5s, img.Md5)
	}

	// Everything left in np isn't referenced by an entry in sourceImgs.
	for _, targetImg := range targetMd5s {
		work.ToDelete = append(work.ToDelete, targetImg)
	}

	return &work, nil
}

func uploadGooglephotoToNixplay(from *googlephotos.CachedMediaItem, toAlbum int, npClient *http.Client) error {
	imgResp, err := http.Get(from.MediaItem.BaseUrl)
	if err != nil {
		return err
	}
	if imgResp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed downloading Googlephoto to upload (%d)", imgResp.StatusCode)
	}
	defer imgResp.Body.Close()

	filename := from.MediaItem.Filename
	filetype := imgResp.Header.Get("content-type")
	filesizeStr := imgResp.Header.Get("content-length")
	filesize, err := strconv.ParseUint(filesizeStr, 10, 64)
	if err != nil {
		return err
	}

	return nixplay.UploadPhoto(npClient, toAlbum, filename, filetype, filesize, imgResp.Body)
}

func deleteGooglephotoFromNixplay(del *nixplay.Photo, npClient *http.Client) error {
	return nixplay.DeletePhoto(npClient, del.ID)
}
