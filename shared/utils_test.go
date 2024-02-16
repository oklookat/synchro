package shared

import (
	"net/url"
	"reflect"
	"testing"
)

func TestCompareImages(t *testing.T) {
	url1, _ := url.Parse("https://avatars.yandex.net/get-music-content/2806365/401f25f3.a.10432824-1/m300x300")
	url2, _ := url.Parse("https://i.scdn.co/image/ab67616d0000b273b492477206075438e0751176")
	result, err := CompareImages(*url1, *url2)
	if err != nil {
		t.Fatalf(err.Error())
	}
	t.Logf("same covers: %v", result)
}

func TestRemoteIDSlice(t *testing.T) {
	input := RemoteIDSlice[string]{"a", "b", "c", "d"}
	expected := RemoteIDSlice[string]{"d", "c", "b", "a"}

	reversed := make(RemoteIDSlice[string], len(input))
	copy(reversed, input)
	reversed.Reverse()

	if !reflect.DeepEqual(reversed, expected) {
		t.Errorf("RemoteIDSlice(%v) = %v; expected %v", input, reversed, expected)
	}
}

func TestSameNameSlices(t *testing.T) {
	testCases := []struct {
		s1       []string
		s2       []string
		expected float64
	}{
		{
			s1:       []string{"John", "Mary", "Anna"},
			s2:       []string{"Peter", "David"},
			expected: 0.0, // no matches
		},
		{
			s1:       []string{"John", "Mary", "Anna"},
			s2:       []string{"Mary", "John", "Anna"},
			expected: 1.0, // all names match
		},
		{
			s1:       []string{"John", "Mary", "Anna"},
			s2:       []string{"mary", "jOhn", "aNNa"},
			expected: 1.0, // case-insensitive matching
		},
		{
			s1:       []string{"John", "Mary", "Anna", "Ok"},
			s2:       []string{"Mary", "John"},
			expected: 0.5,
		},
	}

	for _, tc := range testCases {
		result := SameNameSlices(tc.s1, tc.s2)
		if result != tc.expected {
			t.Errorf("For %v and %v, expected %f but got %f", tc.s1, tc.s2, tc.expected, result)
		}
	}
}
