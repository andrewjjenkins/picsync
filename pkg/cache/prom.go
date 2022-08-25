package cache

import (
	"fmt"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type cachePromImpl struct {
	promFactory promauto.Factory

	cacheGetHitsGooglephotos       prometheus.Counter
	cacheGetMissesGooglephotos     prometheus.Counter
	cacheUpsertsUpdateGooglephotos prometheus.Counter
	cacheUpsertsInsertGooglephotos prometheus.Counter
	cacheEntriesGooglephotos       prometheus.Gauge

	cacheGetHitsNixplay       prometheus.Counter
	cacheGetMissesNixplay     prometheus.Counter
	cacheUpsertsUpdateNixplay prometheus.Counter
	cacheUpsertsInsertNixplay prometheus.Counter
	cacheEntriesNixplay       prometheus.Gauge

	cacheFileSize prometheus.GaugeFunc
}

func (c *cacheImpl) promRegister(reg prometheus.Registerer) error {
	// Can take nil (counters are created but not registered)
	c.prom.promFactory = promauto.With(reg)

	c.prom.cacheGetHitsGooglephotos = c.prom.promFactory.NewCounter(
		prometheus.CounterOpts{
			Name: "cache_get_hits_googlephotos",
			Help: "Number of gets that were found in the cache",
		})
	c.prom.cacheGetMissesGooglephotos = c.prom.promFactory.NewCounter(
		prometheus.CounterOpts{
			Name: "cache_get_misses_googlephotos",
			Help: "Number of gets that were not found in the cache",
		})
	c.prom.cacheUpsertsUpdateGooglephotos = c.prom.promFactory.NewCounter(
		prometheus.CounterOpts{
			Name: "cache_upserts_update_googlephotos",
			Help: "Number of upserts that were updates (found in the cache)",
		})
	c.prom.cacheUpsertsInsertGooglephotos = c.prom.promFactory.NewCounter(
		prometheus.CounterOpts{
			Name: "cache_upserts_insert_googlephotos",
			Help: "Number of upserts that were inserts (not found in the cache)",
		})
	c.prom.cacheEntriesGooglephotos = c.prom.promFactory.NewGauge(
		prometheus.GaugeOpts{
			Name: "cache_entries_googlephotos",
			Help: "Number of entries in the googlephotos cache",
		})
	c.prom.cacheGetHitsNixplay = c.prom.promFactory.NewCounter(
		prometheus.CounterOpts{
			Name: "cache_get_hits_nixplay",
			Help: "Number of gets that were found in the cache",
		})
	c.prom.cacheGetMissesNixplay = c.prom.promFactory.NewCounter(
		prometheus.CounterOpts{
			Name: "cache_get_misses_nixplay",
			Help: "Number of gets that were not found in the cache",
		})
	c.prom.cacheUpsertsUpdateNixplay = c.prom.promFactory.NewCounter(
		prometheus.CounterOpts{
			Name: "cache_upserts_update_nixplay",
			Help: "Number of upserts that were updates (found in the cache)",
		})
	c.prom.cacheUpsertsInsertNixplay = c.prom.promFactory.NewCounter(
		prometheus.CounterOpts{
			Name: "cache_upserts_insert_nixplay",
			Help: "Number of upserts that were inserts (not found in the cache)",
		})
	c.prom.cacheEntriesNixplay = c.prom.promFactory.NewGauge(
		prometheus.GaugeOpts{
			Name: "cache_entries_nixplay",
			Help: "Number of entries in the nixplay cache",
		})

	// Rather than re-counting the cache every time, we will init the gauge
	// once and then keep it up-to-date by Inc/Dec.
	status, err := c.Status()
	if err != nil {
		return err
	}
	c.prom.cacheEntriesGooglephotos.Set(float64(status.GooglePhotosValidRows))
	c.prom.cacheEntriesNixplay.Set(float64(status.NixplayValidRows))

	c.prom.cacheFileSize = c.prom.promFactory.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "cache_file_size",
			Help: "Size of the cache database in bytes",
		}, c.newDbFilesizeGetter())

	return nil
}

func (c *cacheImpl) newDbFilesizeGetter() func() float64 {
	return func() float64 {
		fileinfo, err := os.Stat(c.dbFilename)
		if err != nil {
			fmt.Printf("Warning: error reading cache filesize (%s): %s\n",
				c.dbFilename, err.Error())
			return -1
		}
		return float64(fileinfo.Size())
	}
}
