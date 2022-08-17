package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/andrewjjenkins/picsync/pkg/nixplay"
	"github.com/andrewjjenkins/picsync/pkg/smugmug"

	"github.com/robfig/cron"
)

func runSyncSmugmugOnce(smugmugAlbumName string, nixplayAlbumName string) {
	err := doSyncSmugmug(smugmugAlbumName, nixplayAlbumName)
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
	os.Exit(0)
}

func runSyncSmugmugEvery(
	smugmugAlbumName string,
	nixplayAlbumName string,
	every string,
) {
	everyCronSpec := fmt.Sprintf("@every %s", every)
	job := func() { doSyncSmugmug(smugmugAlbumName, nixplayAlbumName) }

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

func doSyncSmugmug(smugmugAlbumName string, nixplayAlbumName string) error {
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
				return fmt.Errorf("duplicate SmugMug albums named %s", smugmugAlbumName)
			}
			smAlbum = a
		}
	}
	if smAlbum == nil {
		return fmt.Errorf("could not find SmugMug album %s", smugmugAlbumName)
	}

	smImages, err := smugmug.GetAlbumImages(smClient, smAlbum.AlbumKey, maxPics)
	if err != nil {
		return err
	}
	fmt.Printf("Found %d SmugMug images:\n", len(smImages))
	for _, img := range smImages {
		fmt.Printf("  %v\n", img)
	}

	// Get the nixplay image metadata for the requested album
	npAlbum, err := nixplay.GetAlbumByName(npClient, nixplayAlbumName)
	if err != nil {
		return err
	}
	npPhotos, err := nixplay.GetPhotos(npClient, npAlbum.ID)
	if err != nil {
		return err
	}

	work, err := calcSyncSmugmugWork(smImages, npPhotos)
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
			fmt.Printf("Error uploading photo %s (skipping): %v\n", up.FileName, err)
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

type syncSmugmugWork struct {
	ToUpload []*smugmug.AlbumImage
	ToDelete []*nixplay.Photo
}

type smugmugAlbumImagesByMd5 map[string]*smugmug.AlbumImage
type nixplayAlbumImagesByMd5 map[string]*nixplay.Photo

func calcSyncSmugmugWork(smImages []*smugmug.AlbumImage, npPhotos []*nixplay.Photo) (*syncSmugmugWork, error) {
	work := syncSmugmugWork{}

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
	if imgResp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed downloading SmugMug photo to upload (%d)", imgResp.StatusCode)
	}
	defer imgResp.Body.Close()

	filename := from.FileName
	filetype := imgResp.Header.Get("content-type")
	filesizeStr := imgResp.Header.Get("content-length")
	filesize, err := strconv.ParseUint(filesizeStr, 10, 64)
	if err != nil {
		return err
	}

	return nixplay.UploadPhoto(npClient, toAlbum, filename, filetype, filesize, imgResp.Body)
}
