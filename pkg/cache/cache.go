package cache

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type GooglephotoData struct {
	Id             int64
	BaseUrl        string
	Sha256         string
	Md5            string
	GooglephotosId string
	LastUpdated    time.Time
	LastUsed       time.Time
}

type NixplayData struct {
	Id          int64
	URL         string
	Filename    string
	SortDate    string
	Md5         string
	NixplayId   int
	LastUpdated time.Time
	LastUsed    time.Time
}

type Cache interface {
	UpsertGooglephoto(p *GooglephotoData) error
	GetGooglephoto(baseUrl string) (*GooglephotoData, error)
	UpsertNixplay(n *NixplayData) error

	Status() (StatusResponse, error)
}

type cacheImpl struct {
	db *sql.DB

	prom cachePromImpl
}

func New(reg prometheus.Registerer) (Cache, error) {
	cache := cacheImpl{}
	var err error

	cache.db, err = Open()
	if err != nil {
		return nil, err
	}

	cache.promRegister(reg)
	return &cache, nil
}

// Updates/inserts a cache entry for a Google Photo.
// p will be modified with new last used time and, if insert, the new row id.
func (c *cacheImpl) UpsertGooglephoto(p *GooglephotoData) error {
	if p.GooglephotosId == "" || p.BaseUrl == "" || p.Sha256 == "" || p.Md5 == "" {
		return errors.New("must provide GooglephotosId, baseUrl, Sha256, Md5")
	}
	if p.LastUpdated.IsZero() {
		p.LastUpdated = time.Now()
	}
	p.LastUsed = time.Now()

	if p.Id == 0 {
		// Caller doesn't know an Id.  Maybe it's new, but let's try to find
		// it by GooglephotosId first.
		rows, err := c.db.Query("SELECT Id FROM googlephotos WHERE GooglephotosId=?;",
			p.GooglephotosId)
		if err != nil {
			return err
		}
		if rows.Next() {
			// This is an update.  Store the row Id we just found.
			err = rows.Scan(&p.Id)
			rows.Close()
			if err != nil {
				return err
			}
			c.prom.cacheUpsertsUpdateGooglephotos.Inc()
			return c.updateGooglephoto(p)
		}
		rows.Close()
		// This is an insert
		c.prom.cacheUpsertsInsertGooglephotos.Inc()
		return c.insertGooglephoto(p)
	}
	c.prom.cacheUpsertsUpdateGooglephotos.Inc()
	return c.updateGooglephoto(p)
}

func (c *cacheImpl) updateGooglephoto(p *GooglephotoData) error {
	res, err := c.db.Exec("UPDATE googlephotos "+
		"SET Sha256=?, Md5=?, BaseUrl=?, LastUsed=?, LastUpdated=? "+
		"WHERE Id=? AND GooglephotosId=? ;",
		p.Sha256, p.Md5, p.BaseUrl, p.LastUsed, p.LastUpdated, p.Id, p.GooglephotosId)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return fmt.Errorf("expected 1 row updated, got %d", rows)
	}
	return nil
}

func (c *cacheImpl) insertGooglephoto(p *GooglephotoData) error {
	res, err := c.db.Exec("INSERT INTO googlephotos "+
		"(Sha256, Md5, GooglephotosId, BaseUrl, LastUpdated, LastUsed)"+
		"VALUES(?,?,?,?,?,?);",
		p.Sha256, p.Md5, p.GooglephotosId, p.BaseUrl, p.LastUpdated, p.LastUsed)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return fmt.Errorf("expected 1 row updated, got %d", rows)
	}
	rowId, err := res.LastInsertId()
	if err != nil {
		return err
	}
	p.Id = rowId
	c.prom.cacheEntriesGooglephotos.Inc()
	return nil
}

func (c *cacheImpl) GetGooglephoto(googlephotosId string) (*GooglephotoData, error) {
	rows, err := c.db.Query(
		"SELECT Id, BaseUrl, Sha256, Md5, GooglephotosId, LastUpdated, LastUsed "+
			"FROM googlephotos WHERE GooglephotosId=? LIMIT 1;",
		googlephotosId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	found := rows.Next()
	if !found {
		c.prom.cacheGetMissesGooglephotos.Inc()
		return nil, nil
	}
	var toRet GooglephotoData
	rows.Scan(&toRet.Id, &toRet.BaseUrl, &toRet.Sha256, &toRet.Md5,
		&toRet.GooglephotosId, &toRet.LastUpdated, &toRet.LastUsed)
	c.prom.cacheGetHitsGooglephotos.Inc()
	return &toRet, nil
}

func (c *cacheImpl) UpsertNixplay(n *NixplayData) error {
	if n.Md5 == "" || n.NixplayId == 0 || n.Filename == "" ||
		n.URL == "" || n.SortDate == "" {
		return errors.New("must provide Md5, NixplayID, Filename, URL, SortDate")
	}
	if n.LastUpdated.IsZero() {
		n.LastUpdated = time.Now()
	}
	n.LastUsed = time.Now()

	if n.Id == 0 {
		// Caller doesn't know an Id.  Maybe it's new, but let's try to find
		// it by md5 first.
		// This assumes that there will be no md5 collisions.  It's unlikely
		// any one user's photo collection will have some, and I don't know
		// enough about what is persistent/unique in Nixplay's API to choose
		// something like Google's baseUrl.
		// We could use Sha256 someday if we implemented something like the
		// way we download all the images for googlephotos.
		rows, err := c.db.Query("SELECT Id FROM nixplay WHERE Md5=?;", n.Md5)
		if err != nil {
			return err
		}
		if rows.Next() {
			// This is an update.  Store the row ID we just found.
			err = rows.Scan(&n.Id)
			rows.Close()
			if err != nil {
				return err
			}
			c.prom.cacheUpsertsUpdateNixplay.Inc()
			return c.updateNixplay(n)
		}
		rows.Close()
		// This is an insert.
		c.prom.cacheUpsertsInsertNixplay.Inc()
		return c.insertNixplay(n)
	}
	c.prom.cacheUpsertsUpdateNixplay.Inc()
	return c.updateNixplay(n)
}

func (c *cacheImpl) updateNixplay(n *NixplayData) error {
	res, err := c.db.Exec("UPDATE nixplay "+
		"SET NixplayId=?, Filename=?, URL=?, SortDate=?, LastUsed=?, LastUpdated=? "+
		"WHERE Id=? AND Md5=? ;",
		n.NixplayId, n.Filename, n.URL, n.SortDate, n.LastUsed,
		n.LastUpdated, n.Id, n.Md5)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return fmt.Errorf("expected 1 row updated, got %d", rows)
	}
	return nil
}

func (c *cacheImpl) insertNixplay(n *NixplayData) error {
	res, err := c.db.Exec("INSERT INTO nixplay "+
		"(NixplayId, Filename, URL, SortDate, Md5, LastUpdated, LastUsed)"+
		"VALUES(?,?,?,?,?,?,?);",
		n.NixplayId, n.Filename, n.URL, n.SortDate, n.Md5,
		n.LastUpdated, n.LastUsed)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return fmt.Errorf("expected 1 row updated, got %d", rows)
	}
	rowId, err := res.LastInsertId()
	if err != nil {
		return err
	}
	n.Id = rowId
	c.prom.cacheEntriesNixplay.Inc()
	return nil
}

type StatusResponse struct {
	GooglePhotosValidRows int64
	NixplayValidRows      int64
}

func (c *cacheImpl) Status() (StatusResponse, error) {
	resp := StatusResponse{}

	// Because Google never modifies content at a particular Id
	// (instead creating a new Id), the mapping from Id to
	// md5 never becomes invalid.
	// FIXME: In the future, we could check consistency and also remove
	// least-recently-used entries.
	rows, err := c.db.Query("SELECT COUNT(Id) FROM googlephotos")
	if err != nil {
		return StatusResponse{}, err
	}
	rows.Next()
	rows.Scan(&resp.GooglePhotosValidRows)
	rows.Close()

	// FIXME: Implement expiring of Nixplay cache entries.
	rows, err = c.db.Query("SELECT COUNT(Id) FROM nixplay")
	if err != nil {
		return StatusResponse{}, err
	}
	defer rows.Close()
	rows.Next()
	rows.Scan(&resp.NixplayValidRows)
	rows.Close()

	return resp, nil
}
