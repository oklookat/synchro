package shared

import (
	"net/url"
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

func TestCompareNames(t *testing.T) {
	cases := []struct {
		n1       string
		n2       string
		expected float64
	}{
		{
			n1:       "Super track",
			n2:       "Super track",
			expected: 1.0,
		},
		{
			n1:       "tears in the club (feat. the weeknd)",
			n2:       "TEARS IN THE CLUB",
			expected: 0.8,
		},
		{
			n1:       "Автомобили (Video Mix 1989)",
			n2:       "АвТОмоБилИ",
			expected: 0.8,
		},
		{
			n1:       "PERMSKY KRAY",
			n2:       "пермскИЙ край",
			expected: 0.8,
		},
		{
			n1:       "MOYA LUBIMAYA MASHINA",
			n2:       "моя любимая машина",
			expected: 0.8,
		},
		{
			n1:       "Smack That",
			n2:       "Smack That (durak remix)",
			expected: 0.8,
		},
		{
			n1:       "LiiTE_13-37 w/ iglooghost",
			n2:       "LiiTE_13-37",
			expected: 0.8,
		},
		{
			n1:       "Miss Anthropocene (Deluxe Edition)",
			n2:       "Miss Anthropocene",
			expected: 0.8,
		},
		{
			n1:       "Bliding Lights",
			n2:       "Gliding Lights",
			expected: 0.0,
		},
		{
			n1:       "SH",
			n2:       "Ш",
			expected: 1.0,
		},
		{
			n1:       "200 KM/H In The Wrong Lane (10th Anniversary Edition)",
			n2:       "200 KM/H In The Wrong Lane",
			expected: 0.8,
		},
	}
	for _, cased := range cases {
		res := CompareNames(cased.n1, cased.n2)
		if res != cased.expected {
			t.Errorf("Compare '%s' with '%s'. Expected: %v, got: %v", cased.n1,
				cased.n2, cased.expected, res)
		}
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
			expected: 0.0,
		},
		{
			s1:       []string{"John", "Mary", "Anna"},
			s2:       []string{"Mary", "John", "Anna"},
			expected: 1.0,
		},
		{
			s1:       []string{"John", "Mary", "Anna"},
			s2:       []string{"mary", "jOhn", "aNNa"},
			expected: 1.0,
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
