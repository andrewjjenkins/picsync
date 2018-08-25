package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/andrewjjenkins/picsync/pkg/nixplay"
	"github.com/andrewjjenkins/picsync/pkg/smugmug"

	"github.com/robfig/cron"
	"github.com/spf13/cobra"
)

var (
	sync = &cobra.Command{
		Use:   "sync <smugmug-album> <nixplay-album>",
		Short: "Sync a smugmug album to a nixplay album",
		Run:   runSync,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 2 || args[0] == "" || args[1] == "" {
				return errors.New("Specify the name of the source SmugMug album and destination NixPlay album")
			}
			return nil
		},
	}

	syncEvery string
)

func init() {
	sync.PersistentFlags().StringVarP(
		&syncEvery,
		"every",
		"d",
		"",
		"Sync every interval (like \"30s\" or \"1h\")",
	)
	rootCmd.AddCommand(sync)
}

func runSync(cmd *cobra.Command, args []string) {
	smugmugAlbumName := args[0]
	nixplayAlbumName := args[1]

	if syncEvery != "" {
		runSyncEvery(smugmugAlbumName, nixplayAlbumName, syncEvery)
	} else {
		runSyncOnce(smugmugAlbumName, nixplayAlbumName)
	}
}

func runSyncOnce(smugmugAlbumName string, nixplayAlbumName string) {
	err := doSync(smugmugAlbumName, nixplayAlbumName)
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
	os.Exit(0)
}

func runSyncEvery(
	smugmugAlbumName string,
	nixplayAlbumName string,
	every string,
) {
	everyCronSpec := fmt.Sprintf("@every %s", every)

	c := cron.New()
	err := c.AddFunc(
		everyCronSpec,
		func() { doSync(smugmugAlbumName, nixplayAlbumName) },
	)
	if err != nil {
		fmt.Printf("Cannot run every %s: %v\n", every, err)
		os.Exit(1)
	}
	fmt.Printf("Syncing every %s\n", every)
	c.Run()
}

func doSync(smugmugAlbumName string, nixplayAlbumName string) error {
	fmt.Printf(
		"Syncing images from SmugMug album %s to Nixplay album %s\n",
		smugmugAlbumName, nixplayAlbumName,
	)

	// Log in to both services; exit early if there's an auth problem
	smClient := getSmugmugClientOrExit()
	npClient := getNixplayClientOrExit()

	// Get the smugmug image metadata for the requested album
	user, err := smugmug.GetThisUser(smClient)
	if err != nil {
		return err
	}
	smAlbums, err := smugmug.GetAlbumsForUser(smClient, user.NickName)
	if err != nil {
		return err
	}
	var smAlbum *smugmug.Album
	for _, a := range smAlbums {
		if a.Name == smugmugAlbumName {
			if smAlbum != nil {
				return fmt.Errorf("Duplicate SmugMug albums named %s", smugmugAlbumName)
			}
			smAlbum = a
		}
	}
	if smAlbum == nil {
		return fmt.Errorf("Could not find SmugMug album %s", smugmugAlbumName)
	}

	smImages, err := smugmug.GetAlbumImages(smClient, smAlbum.AlbumKey)
	if err != nil {
		return err
	}
	fmt.Printf("Found SmugMug images: %v\n", smImages)

	// Get the nixplay image metadata for the requested album
	npAlbums, err := nixplay.GetAlbums(npClient)
	if err != nil {
		return err
	}
	var npAlbum *nixplay.Album
	for _, a := range npAlbums {
		if a.Title == nixplayAlbumName {
			if npAlbum != nil {
				return fmt.Errorf("Duplicate Nixplay albums named %s", nixplayAlbumName)
			}
			npAlbum = a
		}
	}
	if npAlbum == nil {
		return fmt.Errorf("Could not find Nixplay album %s", nixplayAlbumName)
	}
	npPhotos, err := nixplay.GetPhotos(npClient, npAlbum.ID)
	if err != nil {
		return err
	}

	work, err := calcSyncWork(smImages, npPhotos)
	if err != nil {
		return err
	}
	fmt.Printf(
		"Sync Work:\n  ToUpload: %v\n  ToDelete: %v\n",
		work.ToUpload, work.ToDelete,
	)

	for _, up := range work.ToUpload {
		fmt.Printf("Uploading %s to Nixplay...", up.FileName)
		err := uploadSmugmugToNixplay(up, npAlbum.ID, smClient, npClient)
		if err != nil {
			fmt.Printf("\n")
			return err
		}
		fmt.Printf("DONE\n")
	}

	if len(work.ToUpload) > 0 {
		fmt.Printf("Sleeping for 5 seconds to let nixplay digest uploaded photos...\n")
		time.Sleep(5 * time.Second)
	}

	// Now, get the photos again and put them in a playlist
	npPhotos, err = nixplay.GetPhotos(npClient, npAlbum.ID)
	if err != nil {
		return err
	}
	ssName := fmt.Sprintf("ss_%s", nixplayAlbumName)
	ss, err := nixplay.GetSlideshowByName(npClient, ssName)
	if err != nil {
		fmt.Printf("Could not find slideshow %s (%v), creating\n", ssName, err)
		fmt.Printf(
			"If this works, you must then assign the slideshow %s to frames - "+
				"this program will not do that (but it will update the slideshow once "+
				"you've assigned it)\n",
			ssName,
		)
		ss, err = nixplay.CreateSlideshow(npClient, ssName)
		if err != nil {
			return err
		}
	}
	err = nixplay.PublishSlideshow(npClient, ss, npPhotos)
	if err != nil {
		return err
	}
	fmt.Printf("Published %d photos to slideshow %s\n", len(npPhotos), ssName)

	return nil
}

type syncWork struct {
	ToUpload []*smugmug.AlbumImage
	ToDelete []*nixplay.Photo
}

type smugmugAlbumImagesByMd5 map[string]*smugmug.AlbumImage
type nixplayAlbumImagesByMd5 map[string]*nixplay.Photo

func calcSyncWork(smImages []*smugmug.AlbumImage, npPhotos []*nixplay.Photo) (*syncWork, error) {
	work := syncWork{}

	// Store a lookup-by-MD5 for SmugMug
	sm := make(smugmugAlbumImagesByMd5)
	for _, img := range smImages {
		alreadyThere, ok := sm[img.ArchivedMD5]
		if ok {
			fmt.Printf(
				"Warning: duplicate images with MD5 %s (%s, %s)\n",
				img.ArchivedMD5, alreadyThere.FileName, img.FileName,
			)
			continue
		}
		sm[img.ArchivedMD5] = img
	}

	// Store a lookup-by-MD5 for Nixplay
	np := make(nixplayAlbumImagesByMd5)
	for _, img := range npPhotos {
		alreadyThere, ok := np[img.Md5]
		if ok {
			fmt.Printf(
				"Warning: duplicate images with MD5 %s (%s, %s)\n",
				img.Md5, alreadyThere.Filename, img.Filename,
			)
			continue
		}
		np[img.Md5] = img
	}

	// Calculate the difference
	for md5, smImg := range sm {
		_, ok := np[md5]
		if !ok {
			work.ToUpload = append(work.ToUpload, smImg)
			continue
		}
		// If it is present, delete it from np so it won't count toward ToDelete
		delete(np, md5)
	}

	// Everything left in np isn't referenced by an entry in sm
	for _, npImg := range np {
		work.ToDelete = append(work.ToDelete, npImg)
	}

	return &work, nil
}

func uploadSmugmugToNixplay(from *smugmug.AlbumImage, toAlbum int, smClient *http.Client, npClient *http.Client) error {
	imgResp, err := smClient.Get(from.ArchivedURI)
	if err != nil {
		return err
	}
	defer imgResp.Body.Close()

	filename := from.FileName
	filetype := imgResp.Header.Get("content-type")
	filesizeStr := imgResp.Header.Get("content-length")
	filesize, err := strconv.ParseUint(filesizeStr, 10, 64)

	return nixplay.UploadPhoto(npClient, toAlbum, filename, filetype, filesize, imgResp.Body)
}
