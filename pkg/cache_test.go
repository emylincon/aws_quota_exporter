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

func FuzzNewCache(f *testing.F) {
	type args struct {
		fileName string
		lifeTime int
	}
	tests := []struct {
		name string
		args args
		want *Cache
	}{
		{
			name: "test_cache1",
			args: args{
				fileName: "test_cache1.json",
				lifeTime: 1,
			},
			want: &Cache{
				FileName: CacheFolder + "test_cache1.json",
				LifeTime: time.Second * time.Duration(1),
				Expires:  time.Now(),
			},
		},
		{
			name: "test_cache2",
			args: args{
				fileName: "test_cache2.json",
				lifeTime: 10,
			},
			want: &Cache{
				FileName: CacheFolder + "test_cache2.json",
				LifeTime: time.Second * time.Duration(10),
				Expires:  time.Now(),
			},
		},
	}
	for _, tt := range tests {
		f.Add(tt.args.fileName, tt.args.lifeTime)
	}
	f.Fuzz(
		func(t *testing.T, filename string, lifetime int) {
			c := NewCache(filename, time.Duration(lifetime))
			want := CacheFolder + filename
			if c.FileName != want {
				t.Errorf("NewCache().FileName = %v, want %v", c.FileName, want)
			}
		},
	)

}
