package rss

import (
	"testing"
	"time"

	"github.com/mmcdole/gofeed"
	"github.com/write-blog-every-week/write-blog-every-week-remind/database"
	"github.com/write-blog-every-week/write-blog-every-week-remind/date"
)

var asiaTokyo, _ = time.LoadLocation("Asia/Tokyo")

func parse(published string) time.Time {
	// なぜかRFC3339でうまくパースできないのでRFC1123Z
	parsed, _ := time.ParseInLocation(time.RFC1123Z, published, asiaTokyo)
	return parsed
}

func item(published string) *gofeed.Item {
	parsed := parse(published)
	return &gofeed.Item {
		Published: published,
		PublishedParsed: &parsed,
	}
}

func TestGetLatestFeedPubDate(t *testing.T) {
	date.SetFakeTime(time.Date(2018, 12, 27, 0, 0, 0, 0, asiaTokyo))
	thisMonday := date.GetThisMonday()
	tests := []struct {
		name		 string
		feed		 *gofeed.Feed
		requireCount int
		want		 time.Time
	}{
		{
			name: "not enough feeds",
			feed: &gofeed.Feed{},
			requireCount: 0,
			want: thisMonday,
		},
		{
			name: "1 feed required and written",
			feed: &gofeed.Feed{
				Items: []*gofeed.Item{
					item("Wed, 26 Dec 2018 19:00:00 +0900"),
				},
			},
			requireCount: 0,
			want: parse("Wed, 26 Dec 2018 19:00:00 +0900"),
		},
		{
			name: "1 feed required and not written",
			feed: &gofeed.Feed{
				Items: []*gofeed.Item{
					item("Wed, 19 Dec 2018 19:00:00 +0900"),
				},
			},
			requireCount: 0,
			want: parse("Wed, 19 Dec 2018 19:00:00 +0900"),
		},
		{
			name: "2 feeds required and only 1 feed exists",
			feed: &gofeed.Feed{
				Items: []*gofeed.Item{
					item("Wed, 26 Dec 2018 19:00:00 +0900"),
				},
			},
			requireCount: 1,
			want: thisMonday,
		},
		{
			name: "2 feeds required and only 1 feed written this week",
			feed: &gofeed.Feed{
				Items: []*gofeed.Item{
					item("Wed, 26 Dec 2018 19:00:00 +0900"),
					item("Tue, 18 Dec 2018 19:00:00 +0900"),
				},
			},
			requireCount: 1,
			want: parse("Tue, 18 Dec 2018 19:00:00 +0900"),
		},
		{
			name: "2 feeds required and written",
			feed: &gofeed.Feed{
				Items: []*gofeed.Item{
					item("Wed, 26 Dec 2018 19:00:00 +0900"),
					item("Tue, 25 Dec 2018 19:00:00 +0900"),
				},
			},
			requireCount: 1,
			want: parse("Tue, 25 Dec 2018 19:00:00 +0900"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getLatestFeedPubDate(tt.feed, tt.requireCount, asiaTokyo); !got.Equal(tt.want) {
				t.Errorf("want \n%s\n, but got \n%s\n", tt.want, got)
			}
		})
	}
}

var parseMap = map[string]*gofeed.Feed{
	"noitem": &gofeed.Feed{
		Items: []*gofeed.Item{},
	},
	"1item": &gofeed.Feed{
		Items: []*gofeed.Item{
			item("Tue, 25 Dec 2018 19:00:00 +0900"),
		},
	},
	"2items": &gofeed.Feed{
		Items: []*gofeed.Item{
			item("Wed, 26 Dec 2018 19:00:00 +0900"),
			item("Tue, 25 Dec 2018 19:00:00 +0900"),
		},
	},
}

type mockParser struct {
}

func (mp *mockParser) ParseURL(url string) (feed *gofeed.Feed, err error) {
	return parseMap[url], nil
}

func TestFindTargetUserList(t *testing.T) {
	date.SetFakeTime(time.Date(2018, 12, 27, 0, 0, 0, 0, asiaTokyo))
	thisMonday := time.Date(2018, 12, 24, 0, 0, 0, 0, asiaTokyo)
	tests := []struct {
		name	string
		members	[]database.WriteBlogEveryWeek
		monday	time.Time
		want	map[string]int
	}{
		{
			name: "0 required returns 0",
			members: []database.WriteBlogEveryWeek{
				database.WriteBlogEveryWeek{
					UserID: "user1",
					FeedURL: "1item",
					RequireCount: 0,
				},
			},
			monday: thisMonday,
			want: map[string]int{
				"user1": 0,
			},
		},
		{
			name: "required 1 more returns 1",
			members: []database.WriteBlogEveryWeek{
				database.WriteBlogEveryWeek{
					UserID: "user1",
					FeedURL: "noitem",
					RequireCount: 1,
				},
				database.WriteBlogEveryWeek{
					UserID: "user2",
					FeedURL: "1item",
					RequireCount: 2,
				},
			},
			monday: thisMonday,
			want: map[string]int{
				"user1": 1,
				"user2": 1,
			},
		},
		{
			name: "2 required 2 written returns 0",
			members: []database.WriteBlogEveryWeek{
				database.WriteBlogEveryWeek{
					UserID: "user1",
					FeedURL: "2items",
					RequireCount: 2,
				},
			},
			monday: thisMonday,
			want: map[string]int{
				"user1": 0,
			},
		},
	}
	parser := &mockParser{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findTargetUserList(tt.members, tt.monday, parser)
			for k, v := range got {
				if v != tt.want[k] {
					t.Errorf("want \n%d for %s\n, but got \n%d\n", tt.want[k], k, v)
				}
			}
		})
	}
}