package database

import (
	"reflect"
	"testing"

	"github.com/write-blog-every-week/write-blog-every-week-remind/config"
)

func TestFindByPK(t *testing.T) {
	dummyConfig := config.ConfigData{}
	tests := []struct {
		name string
		pk   string
		want WriteBlogEveryWeek
	}{
		{
			name: "existing user",
			pk:   "user1",
			want: WriteBlogEveryWeek{
				UserID:       "id_user1",
				UserName:     "user1",
				FeedURL:      "https://blog.example.com",
				RequireCount: 1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FindByPK(dummyConfig, tt.pk); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("want %#v, but got %v", tt.want, got)
			}
		})
	}
}
