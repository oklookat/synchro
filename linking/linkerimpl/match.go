package linkerimpl

import (
	"strings"

	"github.com/oklookat/synchro/shared"
)

// Get the most similar album from the array, based on origin.
//
// If there are no similar albums, returns nil.
func matchAlbum(origin shared.RemoteAlbum, albums []shared.RemoteAlbum) shared.RemoteAlbum {
	lastWeight := 0.0
	bestIndex := 0

	for i := range albums {
		if shared.IsNil(albums[i]) {
			continue
		}

		weight, exact := compareAlbums(origin, albums[i])
		if exact {
			return albums[i]
		}

		// Skip the unlikely.
		if weight < 0.6 {
			continue
		}

		if weight > lastWeight {
			lastWeight = weight
			bestIndex = i
		}
	}

	if lastWeight == 0 || len(albums) == 0 {
		return nil
	}

	return albums[bestIndex]
}

// Compare albums.
//
// Exact - albums equals by UPC or EAN.
func compareAlbums(first, second shared.RemoteAlbum) (totalWeight float64, exact bool) {
	if shared.IsNil(first, second) {
		return 0, false
	}

	// If UPC or EAN, compare by them.
	if first.UPC() != nil && second.UPC() != nil {
		if strings.EqualFold(*first.UPC(), *second.UPC()) {
			return 1, true
		}
	}
	if first.EAN() != nil && second.EAN() != nil {
		if strings.EqualFold(*first.EAN(), *second.EAN()) {
			return 1, true
		}
	}

	// Remotes can have different tracks for the same album.
	trackCountWeightMap := map[uint64]float64{
		// Max 4 tracks tolerance.
		4: 0.06,
		3: 0.07,
		2: 0.08,
		1: 0.09,
		0: 0.1,
	}
	trackCountWeight := shared.NumDiffWeight(uint64(first.TrackCount()), uint64(second.TrackCount()), trackCountWeightMap)
	if trackCountWeight == 0 {
		return 0, false
	}

	// Compare covers.
	var coversWeight float64
	if first.CoverURL() != nil && second.CoverURL() != nil {
		same, err := shared.CompareImages(*first.CoverURL(), *second.CoverURL())
		if err == nil && same {
			coversWeight = 0.3
		}
	}

	// Compare album names.
	albumNamesWeight := shared.CompareNames(first.Name(), second.Name()) * 0.3

	// Check the inclusion of artists on both albums.
	var firstArtists, secondArtists []string
	for _, artist := range first.Artists() {
		firstArtists = append(firstArtists, artist.Name())
	}
	for _, artist := range second.Artists() {
		secondArtists = append(secondArtists, artist.Name())
	}
	artistNamesWeight := shared.SameNameSlices(firstArtists, secondArtists) * 0.2

	// We could compare more album artist names,
	// but different remotes may have different order of album artists.

	// Remotes can have different release dates for the same album.
	yearWeightMap := map[uint64]float64{
		6: 0.04,
		5: 0.05,
		4: 0.06,
		3: 0.07,
		2: 0.08,
		1: 0.09,
		0: 0.1,
	}
	yearWeight := shared.NumDiffWeight(uint64(first.Year()), uint64(second.Year()), yearWeightMap)

	total := coversWeight + albumNamesWeight + artistNamesWeight + yearWeight + trackCountWeight

	if total >= 0.99 {
		total = 1
		return total, true
	}
	return total, false
}

// Get the most similar track from the array, based on origin.
//
// If there are no similar tracks, returns nil.
func matchTrack(origin shared.RemoteTrack, tracks []shared.RemoteTrack) shared.RemoteTrack {
	lastWeight := 0.0
	bestIndex := 0

	for i := range tracks {
		if shared.IsNil(tracks[i]) {
			continue
		}

		weight, exact := compareTracks(origin, tracks[i])
		if exact {
			return tracks[i]
		}

		// Skip the unlikely.
		if weight < 0.6 {
			continue
		}

		if weight > lastWeight {
			lastWeight = weight
			bestIndex = i
		}
	}

	if lastWeight == 0 || len(tracks) == 0 {
		return nil
	}

	return tracks[bestIndex]
}

// Compare tracks.
//
// Exact - tracks equals by ISRC.
func compareTracks(first, second shared.RemoteTrack) (totalWeight float64, exact bool) {
	if shared.IsNil(first, second) {
		return 0, false
	}

	// If ISRC, compare by them.
	if first.ISRC() != nil && second.ISRC() != nil {
		if strings.EqualFold(*first.ISRC(), *second.ISRC()) {
			return 1, true
		}
	}

	// Length.
	// Remotes can have different track length for same track.
	lengthWeight := 0.0
	lengthDiff := shared.NumDiff(uint64(first.LengthMs()), uint64(second.LengthMs()))

	// 1.5 seconds tolerance.
	if lengthDiff <= 1500 {
		lengthWeight = 0.2
	} else {
		return 0, false
	}

	// Covers.
	coverWeight := 0.0
	if first.CoverURL() != nil && second.CoverURL() != nil {
		same, err := shared.CompareImages(*first.CoverURL(), *second.CoverURL())
		if err == nil && same {
			coverWeight = 0.2
		}
	}

	// Track names.
	trackNamesWeight := shared.CompareNames(first.Name(), second.Name()) * 0.2

	var artistsNamesWeight = 0.2
	// Else - bypass.
	if len(first.Artists()) > 0 && len(second.Artists()) > 0 {
		// Inclusion of artists on both tracks.
		var firstArtists, secondArtists []string
		for _, artist := range first.Artists() {
			firstArtists = append(firstArtists, artist.Name())
		}
		for _, artist := range second.Artists() {
			secondArtists = append(secondArtists, artist.Name())
		}
		artistsNamesWeight = shared.SameNameSlices(firstArtists, secondArtists) * 0.2
	}

	// Albums.
	albumsWeight := 0.2
	fistAlbum, err1 := first.Album()
	secondAlbum, err2 := second.Album()
	// Else - bypass.
	if !shared.IsNil(fistAlbum) && !shared.IsNil(secondAlbum) && err1 == nil && err2 == nil {
		result, _ := compareAlbums(fistAlbum, secondAlbum)
		if result > 1 {
			result = 1
		}
		albumsWeight = result * 0.2
	}

	total := coverWeight + trackNamesWeight + artistsNamesWeight + albumsWeight + lengthWeight

	if total >= 0.99 {
		total = 1
		return total, true
	}
	return total, false
}
