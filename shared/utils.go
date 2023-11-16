package shared

import (
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"time"

	"github.com/adrg/strutil/metrics"
	"github.com/gosimple/slug"
	"github.com/vitali-fedulov/images4"
	"golang.org/x/oauth2"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Compare albums, tracks, artists names.
//
// Max: 1.0 (same).
func CompareNames(name1, name2 string) float64 {
	if strings.EqualFold(name1, name2) {
		return 1
	}

	name1 = Normalize(name1)
	name2 = Normalize(name2)

	if name1 == name2 {
		return 1.0
	}

	if len(name1) > 0 && len(name2) > 0 {
		if name1[0] != name2[0] {
			return 0.0
		}
	}

	// Bullshit check.
	// Split by slug like "HELLO-WORLD" => ["HELLO", "WORLD"].
	splitted1 := strings.Split(name1, "-")
	splitted2 := strings.Split(name2, "-")
	// Get bigger slice.
	var splittedBigger, splittedSmaller []string
	if len(splitted1) > len(splitted2) {
		splittedBigger, splittedSmaller = splitted1, splitted2
	} else {
		splittedBigger, splittedSmaller = splitted2, splitted1
	}
	partsSame := 0
	for i, part := range splittedBigger {
		if i > len(splittedSmaller)-1 {
			break
		}
		// Same parts.
		if part == splittedSmaller[i] {
			partsSame++
			continue
		}

		// Not same parts.
		if i < 1 {
			continue
		}

		// If part contains version.
		if part == "FEAT" || part == "FT" || part == "FEAT." || part == "FT." ||
			part == "DELUXE" || part == "EDITION" ||
			part == "MIXTAPE" ||
			// Example: 10TH (ANNIVERSARY EDITION).
			strings.HasSuffix(part, "TH") {
			// Parts not same, but bigger contains version.
			// In the next iteration, parts will not be same, so break.
			partsSame++
			break
		}
	}
	// Example:
	//
	// bigger = ["HELLO", "WORLD", "DELUXE", "COOL", "VERSION"]
	// smaller = ["HELLO", "WORLD"].
	// Mark as same, because
	// the version may be different depending on the service.
	if len(splittedSmaller) == partsSame {
		return 0.8
	}

	// Levenshtein distance.
	leven := metrics.NewLevenshtein().Distance(name1, name2)
	levenNormalized := float64(leven) / math.Max(float64(len(name1)), float64(len(name2)))
	if levenNormalized < 0.3 {
		return 0.8
	}

	// Jaccard Index.
	jaccard := metrics.NewJaccard().Compare(name1, name2)
	if jaccard >= 0.75 {
		return 0.7
	}

	return 0
}

// Check for nil by comparing or reflect.
// Returns true if one of values is nil.
func IsNil(values ...any) bool {
	for i := range values {
		if values[i] == nil {
			return true
		}
		value := reflect.ValueOf(values[i])
		kind := value.Kind()
		if kind != reflect.Interface &&
			kind != reflect.Pointer &&
			kind != reflect.Map &&
			kind != reflect.Slice &&
			kind != reflect.Func {
			continue
		}
		if !value.IsValid() || value.IsNil() {
			return true
		}
	}
	return false
}

// Slice to chunks.
func ChunkSlice[T any](slice []T, chunkSize int) [][]T {
	if chunkSize <= 0 {
		return nil
	}
	var chunks [][]T
	for i := 0; i < len(slice); i += chunkSize {
		end := i + chunkSize
		if end > len(slice) {
			end = len(slice)
		}
		chunks = append(chunks, slice[i:end])
	}
	return chunks
}

// Trim space -> to ASCII -> to upper.
func Normalize(str string) string {
	return strings.TrimSpace(strings.ToUpper(slug.Make(str)))
}

// Get first word of str. It might be more convenient to find something that way.
//
// If splitted len == 0, returns str.
func SearchablePart(str string) string {
	res := strings.Split(str, " ")
	if len(res) == 0 {
		return str
	}
	return strings.TrimSpace(res[0])
}

// Get first two words of str. It might be more convenient to find something that way.
//
// Returns str, one or two words.
func SearchablePart2(str string) string {
	res := strings.Split(str, " ")
	if len(res) == 0 {
		return str
	}
	if len(res) == 1 {
		return strings.TrimSpace(res[0])
	}
	return strings.TrimSpace(res[0] + " " + res[1])
}

// Who + " " + SearchablePart(what) - all in Normalize().
func SearchableNormalized(who, what string) string {
	whoTr := Normalize(who)
	whatTr := Normalize(what)
	return whoTr + " " + SearchablePart(whatTr)
}

// Downloads img to memory and compare.
//
// Supports jpeg and png.
func CompareImages(url1, url2 url.URL) (bool, error) {
	img1, err := LoadImageFromUrl(url1)
	if err != nil {
		return false, err
	}

	img2, err := LoadImageFromUrl(url2)
	if err != nil {
		return false, err
	}

	icon1 := images4.Icon(img1)
	icon2 := images4.Icon(img2)

	return images4.Similar(icon1, icon2), err
}

// Supports jpeg and png.
func LoadImageFromUrl(url url.URL) (image.Image, error) {
	response, err := http.Get(url.String())
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	contentType := response.Header.Get("Content-Type")
	contentType = strings.ToUpper(contentType)

	// If JPEG.
	if strings.Contains(contentType, "JPEG") || strings.Contains(contentType, "JPG") {
		img, err := jpeg.Decode(response.Body)
		if err != nil {
			return nil, err
		}
		return img, nil
	}

	// If PNG.
	if strings.Contains(contentType, "PNG") {
		img, err := png.Decode(response.Body)
		if err != nil {
			return nil, err
		}
		return img, nil
	}

	return nil, fmt.Errorf("unsupported image type: %s", contentType)
}

// Number difference regardless of the position of the args.
func NumDiff(v1, v2 uint64) uint64 {
	if v2 > v1 {
		v1, v2 = v2, v1
	}
	return v1 - v2
}

// Number difference depend on weightMap.
//
// Example:
//
// v1 = 8, v2 = 10, weightMap = {0: 1.0, 1: 0.8, 2: 0.1}.
//
// Result: 0.1.
func NumDiffWeight(v1, v2 uint64, weightMap map[uint64]float64) float64 {
	diff := NumDiff(v1, v2)
	weight, ok := weightMap[diff]
	if !ok {
		return 0
	}
	return weight
}

// Normalize and get the similarity of the slice.
//
// Max: 1.0 if same names.
func SameNameSlices(s1, s2 []string) float64 {
	if len(s1) < len(s2) {
		s1, s2 = s2, s1
	}

	s1map := make(map[string]int)
	for _, elem := range s1 {
		s1map[Normalize(elem)]++
	}

	totalCount := 0
	for _, elem := range s2 {
		conv := Normalize(elem)
		if count, ok := s1map[conv]; ok && count > 0 {
			s1map[Normalize(conv)]--
			totalCount++
		}
	}

	result := float64(totalCount) / float64(len(s1)+len(s2)-totalCount)
	if result > 1 {
		result = 1
	} else if result < 0 {
		result = 0
	}

	return result
}

// Normalized vals in slice. Skips empty strings, cuts strings with SearchablePart.
func NormalizeStringSliceSearchablePart(slice []string) []string {
	normalized := make([]string, 0, len(slice))
	for _, name := range slice {
		if len(name) == 0 {
			continue
		}
		norm := SearchablePart(Normalize(name))
		normalized = append(normalized, norm)
	}
	return normalized
}

// Generate random word.
//
// Example: "Wobuxahe".
func GenerateWord() string {
	theRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	vowels := []rune{'a', 'e', 'i', 'o', 'u'}
	consonants := []rune{'b', 'c', 'd', 'f', 'g', 'h', 'j', 'k', 'l', 'm', 'n', 'p', 'q', 'r', 's', 't', 'v', 'w', 'x', 'y', 'z'}
	word := make([]rune, theRand.Intn(7)+2)
	for j := range word {
		if j%2 == 0 {
			word[j] = consonants[theRand.Intn(len(consonants))]
		} else {
			word[j] = vowels[theRand.Intn(len(vowels))]
		}
	}
	// First char to upper.
	caser := cases.Title(language.English)
	titleStr := caser.String(string(word))
	return titleStr
}

// Unix nano.
func TimestampNanoNow() int64 {
	return time.Now().UnixNano()
}

// Unix nano.
func TimestampNano(time time.Time) int64 {
	return time.UnixNano()
}

// Unix nano.
func TimeNano(timestamp int64) time.Time {
	return time.Unix(0, timestamp)
}

// Unix ms.
func TimestampNow() int64 {
	return time.Now().Unix()
}

// Unix ms.
func Timestamp(time time.Time) int64 {
	return time.Unix()
}

// Unix ms.
func Time(timestamp int64) time.Time {
	return time.Unix(timestamp, 0)
}

func TokenToAuth(tok *oauth2.Token) (string, error) {
	tokenBytes, err := json.Marshal(tok)
	if err != nil {
		return "", err
	}
	return string(tokenBytes), err
}

func AuthToToken(auth string) (*oauth2.Token, error) {
	tok := &oauth2.Token{}
	err := json.Unmarshal([]byte(auth), tok)
	if err != nil {
		return nil, err
	}
	return tok, err
}
