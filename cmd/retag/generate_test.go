package retag

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_generateCmd_generateMatrix(t *testing.T) {
	dateAdded, _ := time.Parse("2006-01-02", "2018-01-01")

	type fields struct {
		retags []Retag
	}
	tests := []struct {
		name   string
		fields fields
		want   map[string]map[string]string
	}{
		{
			name: "empty retags",
			fields: fields{
				retags: []Retag{},
			},
			want: map[string]map[string]string{},
		},
		{
			name: "one retag",
			fields: fields{
				retags: []Retag{
					{
						Source:         "source/test_repo-hello.123",
						Destination:    "unlisted/destination/test_repo-hello.123",
						DateAdded:      dateAdded,
						Tags:           []string{"tag1", "tag2"},
						EnableTimeBomb: true,
					},
				},
			},
			want: map[string]map[string]string{
				"source_test_repo_hello_123_unlisted": {
					"source":          "source/test_repo-hello.123",
					"destination":     "unlisted/destination/test_repo-hello.123",
					"date_added":      "2018-01-01",
					"tags":            "tag1,tag2",
					"enable_timebomb": "true",
					"tool":            "az",
				},
			},
		},
		{
			name: "multiple retags",
			fields: fields{
				retags: []Retag{
					{
						Source:         "source/test_repo-hello.123",
						Destination:    "unlisted/destination/test_repo-hello.123",
						DateAdded:      dateAdded,
						Tags:           []string{"tag1", "tag2"},
						EnableTimeBomb: true,
					},
					{
						Source:         "source/test_repo-hello.456",
						Destination:    "public/destination/test_repo-hello.456",
						DateAdded:      dateAdded,
						Tags:           []string{"tag3", "tag4"},
						EnableTimeBomb: false,
					},
				},
			},
			want: map[string]map[string]string{
				"source_test_repo_hello_123_unlisted": {
					"source":          "source/test_repo-hello.123",
					"destination":     "unlisted/destination/test_repo-hello.123",
					"date_added":      "2018-01-01",
					"tags":            "tag1,tag2",
					"enable_timebomb": "true",
					"tool":            "az",
				},
				"source_test_repo_hello_456_public": {
					"source":          "source/test_repo-hello.456",
					"destination":     "public/destination/test_repo-hello.456",
					"date_added":      "2018-01-01",
					"tags":            "tag3,tag4",
					"enable_timebomb": "false",
					"tool":            "az",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gc := &generateCmd{
				retags: tt.fields.retags,
			}
			if got := gc.generateMatrix(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("generateCmd.generateMatrix() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parse(t *testing.T) {
	tests := []struct {
		name    string
		file    string
		want    []Retag
		wantErr bool
	}{
		{
			name: "valid config",
			file: "mirror.yaml",
			want: []Retag{
				{
					Source:         "gcr.io/distroless/static",
					Destination:    "unlisted/mirror/gcr/distroless/static",
					Tags:           []string{"debug", "latest", "nonroot"},
					EnableTimeBomb: false,
				},
			},
		},
		{
			name: "mirror config with oras",
			file: "mirror-oras.yaml",
			want: []Retag{
				{
					Source:         "nvcr.io/nvidia/tritonserver",
					Destination:    "unlisted/mirror/nvcr/nvidia/tritonserver",
					Tags:           []string{"22.05-py3", "22.05-py3-min"},
					EnableTimeBomb: false,
					Tool:           "oras",
				},
			},
		},
		{
			name: "mirror config with timebomb",
			file: "mirror-timebomb.yaml",
			want: []Retag{
				{
					Source:         "docker.io/library/postgres",
					Destination:    "public/oss/mirror/docker.io/library/postgres",
					DateAdded:      time.Date(2019, 1, 19, 0, 0, 0, 0, time.UTC),
					Tags:           []string{"12.9-bullseye"},
					EnableTimeBomb: true,
				},
			},
		},
		{
			name:    "config with no tags is an error",
			file:    "mirror-no-tags.yaml",
			wantErr: true,
		},
		{
			name:    "config with no source is an error",
			file:    "mirror-no-source.yaml",
			wantErr: true,
		},
		{
			name:    "invalid config",
			file:    "invalid.yaml",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filebytes, err := os.ReadFile(filepath.Join("testdata", tt.file))
			require.NoError(t, err)
			got, err := parse(filebytes)
			if (err != nil) != tt.wantErr {
				t.Errorf("parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parse() = %v, want %v", got, tt.want)
			}
		})
	}
}
