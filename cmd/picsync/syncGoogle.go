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
	"github.com/spf13/cobra"
)

var (
	syncCmd = &cobra.Command{
		Use:   "sync",
		Short: "Sync pictures to Nixplay",
		Run:   runSync,
	}

	syncEvery string
	dryRun    bool
)

func init() {
	syncCmd.PersistentFlags().StringVarP(
		&syncEvery,
		"every",
		"d",
		"",
		"Sync every interval (like \"30s\" or \"1h\")",
	)
	syncCmd.PersistentFlags().BoolVarP(
		&dryRun,
		"dry-run",
		"n",
		false,
		"Do not actually update anything, just print what you would do",
	)

	rootCmd.AddCommand(syncCmd)
}

func runSync(cmd *cobra.Command, args []string) {
	if syncEvery != "" {
		panic(fmt.Errorf("sync-every unimplemented here"))
	}
	// FIXME: Configurable target
	runSyncGooglephotosOnce(args, "test2022")
}

func runSyncGooglephotosOnce(sourceAlbums []string, nixplayAlbumName string) {
	err := doSyncGooglephotos(sourceAlbums, nixplayAlbumName)
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
	os.Exit(0)
}

func doSyncGooglephotos(sourceAlbums []string, nixplayAlbumName string) error {
	// Log in to services; exit early if there's an auth problem
	gpClient := getGooglephotoClientOrExit()
	npClient := getNixplayClientOrExit()
	c, err := cache.New()
	if err != nil {
		return err
	}

	if len(sourceAlbums) == 0 {
		fmt.Printf("No source album. Cowardly refusing to delete all destination photos.\n")
		return nil
	}
	if len(sourceAlbums) > 1 {
		return fmt.Errorf("multiple source albums (%d) unimplemented", len(sourceAlbums))
	}
	sourceAlbumId := sourceAlbums[0]

	var sourceCacheUpdateCount int
	sourceCacheUpdateCb := func(cached *googlephotos.CachedMediaItem) {
		sourceCacheUpdateCount++
		fmt.Fprintf(os.Stdout, "\033[2K\rUpdating source image %d...", sourceCacheUpdateCount)
	}

	sourceCacheImages, err := googlephotos.UpdateCacheForAlbumId(
		gpClient, c, sourceAlbumId, sourceCacheUpdateCb)
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stdout, "\033[2K\rUpdated %d source images\n",
		sourceCacheUpdateCount)

	// Get the nixplay image metadata for the requested album
	npAlbum, err := nixplay.GetAlbumByName(npClient, nixplayAlbumName)
	if err != nil {
		return err
	}
	npPhotos, err := nixplay.GetPhotos(npClient, npAlbum.ID)
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

	if dryRun {
		return nil
	}

	for i, up := range work.ToUpload {
		fmt.Fprintf(os.Stdout, "\033[2K\rUploading image %d/%d...", i+1, len(work.ToUpload))
		err := uploadGooglephotoToNixplay(up, npAlbum.ID, npClient)
		if err != nil {
			fmt.Printf("\nError uploading photo %s (skipping): %v\n", up.MediaItem.Filename, err)
		}
	}
	fmt.Printf("DONE\nUploading complete.\n")

	if len(work.ToUpload) > 0 {
		fmt.Printf("Sleeping for 5 seconds to let nixplay digest uploaded photos...\n")
		time.Sleep(5 * time.Second)
	}

	// FIXME: This should be commonized
	// Now, get the photos again and put them in a playlist
	npPhotos, err = nixplay.GetPhotos(npClient, npAlbum.ID)
	if err != nil {
		return err
	}
	ssName := fmt.Sprintf("ss_%s", nixplayAlbumName)
	ss, err := nixplay.GetSlideshowByName(npClient, ssName)
	neededCreate := false
	if err != nil {
		fmt.Printf("Could not find slideshow %s (%v), creating\n", ssName, err)
		fmt.Printf(
			"If this works, you must then assign the slideshow %s to frames - "+
				"this program will not do that (but it will update the slideshow once "+
				"you've assigned it)\n",
			ssName,
		)
		neededCreate = true
		ss, err = nixplay.CreateSlideshow(npClient, ssName)
		if err != nil {
			return err
		}
	}

	if len(work.ToUpload) > 0 || neededCreate {
		err = nixplay.PublishSlideshow(npClient, ss, npPhotos)
		if err != nil {
			return err
		}
		fmt.Printf("Published %d photos to slideshow %s\n", len(npPhotos), ssName)
	} else {
		fmt.Printf(
			"No changes required for slideshow %s (%d photos)\n",
			ssName,
			len(npPhotos),
		)
	}
	return nil
}

type syncGooglephotosWork struct {
	ToUpload []*googlephotos.CachedMediaItem
	ToDelete []*nixplay.Photo
}

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
