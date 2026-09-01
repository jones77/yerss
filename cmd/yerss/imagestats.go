package main

import (
	"fmt"
	"io"
	"math"
	"os"
	"sort"

	"yerss/internal/config"
	"yerss/internal/store"
)

// runImageStats gathers image statistics from the local database and prints a
// read-only report to stdout, returning the process exit code. It runs before
// the startup gate and performs no network fetches and no database writes.
func runImageStats(cfg *config.Config, st *store.Store) int {
	stats, err := st.ImageStats()
	if err != nil {
		fmt.Fprintf(os.Stderr, "yerss: cannot read image statistics: %v\n", err)
		return 1
	}
	dbSize, err := st.DBSize()
	if err != nil {
		fmt.Fprintf(os.Stderr, "yerss: cannot stat database: %v\n", err)
		return 1
	}
	printImageStats(os.Stdout, dbSize, stats)
	return 0
}

// printImageStats writes the image footprint report: the aggregated sizes,
// the photo-byte min/median/p90/max distribution, and one line per stored
// image. The decoded-size estimate sums width*height*4 over decodable photos;
// unparseable photos are reported separately and excluded from the estimate.
func printImageStats(w io.Writer, dbSize int64, stats store.ImageStats) {
	fmt.Fprintf(w, "database size (incl. WAL): %d bytes\n", dbSize)
	fmt.Fprintf(w, "articles: %d\n", stats.ArticleCount)
	fmt.Fprintf(w, "stored images: %d\n", stats.ImageCount)
	if stats.ImageCount == 0 {
		return
	}

	fmt.Fprintf(w, "photo bytes total: %d\n", stats.TotalPhotoBytes)
	fmt.Fprintf(w, "block bytes total: %d\n", stats.TotalBlockBytes)

	sizes := make([]int, len(stats.Photos))
	for i, p := range stats.Photos {
		sizes[i] = p.PhotoBytes
	}
	sorted := append([]int(nil), sizes...)
	sort.Ints(sorted)
	fmt.Fprintf(w, "photo per-image bytes: min %d / median %d / p90 %d / max %d\n",
		sorted[0], percentile(sorted, 0.5), percentile(sorted, 0.9), sorted[len(sorted)-1])

	decodedSize, unparseable := 0, 0
	for _, p := range stats.Photos {
		if p.Decodable {
			decodedSize += p.Width * p.Height * 4
		} else {
			unparseable++
		}
	}
	if unparseable > 0 {
		fmt.Fprintf(w, "unparseable photos: %d\n", unparseable)
	}

	fmt.Fprintln(w, "per-photo:")
	for _, p := range stats.Photos {
		if p.Decodable {
			fmt.Fprintf(w, "  %s photo %dB block %dB %dx%d decoded %dB\n",
				p.URL, p.PhotoBytes, p.BlockBytes, p.Width, p.Height, p.Width*p.Height*4)
		} else {
			fmt.Fprintf(w, "  %s photo %dB block %dB unparseable\n",
				p.URL, p.PhotoBytes, p.BlockBytes)
		}
	}
	fmt.Fprintf(w, "decoded-size estimate total: %d bytes\n", decodedSize)
}

// percentile returns the nearest-rank percentile of sorted samples: the value
// at the smallest rank ceil(p*len) (1-based). It returns 0 for an empty slice.
func percentile(sorted []int, p float64) int {
	if len(sorted) == 0 {
		return 0
	}
	idx := int(math.Ceil(p*float64(len(sorted)))) - 1
	if idx < 0 {
		idx = 0
	}
	if idx > len(sorted)-1 {
		idx = len(sorted) - 1
	}
	return sorted[idx]
}