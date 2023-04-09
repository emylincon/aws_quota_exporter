package pkg

import (
	"reflect"
	"testing"
	"time"
)

func TestNewCache(t *testing.T) {
	type args struct {
		fileName string
		lifeTime time.Duration
	}
	tests := []struct {
		name string
		args args
		want *Cache
	}{
		{
			name: "test_cache",
			args: args{
				fileName: "test_cache.json",
				lifeTime: time.Duration(1),
			},
			want: &Cache{
				FileName: CacheFolder + "test_cache.json",
				LifeTime: time.Second * time.Duration(1),
				Expires:  time.Now(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewCache(tt.args.fileName, tt.args.lifeTime); !reflect.DeepEqual(got.FileName, tt.want.FileName) {
				t.Errorf("NewCache() = %v, want %v", got, tt.want)
			}
		})
	}
}
