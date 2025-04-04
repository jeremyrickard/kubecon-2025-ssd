package retag

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_generateCmd_generateMatrix(t *testing.T) {

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
						Source:      "source/test_repo-hello.123",
						Destination: "unlisted/destination/test_repo-hello.123",
						Tags:        []string{"tag1", "tag2"},
					},
				},
			},
			want: map[string]map[string]string{
				"source_test_repo_hello_123_unlisted": {
					"source":      "source/test_repo-hello.123",
					"destination": "unlisted/destination/test_repo-hello.123",
					"tags":        "tag1,tag2",
				},
			},
		},
		{
			name: "multiple retags",
			fields: fields{
				retags: []Retag{
					{
						Source:      "source/test_repo-hello.123",
						Destination: "unlisted/destination/test_repo-hello.123",
						Tags:        []string{"tag1", "tag2"},
					},
					{
						Source:      "source/test_repo-hello.456",
						Destination: "public/destination/test_repo-hello.456",
						Tags:        []string{"tag3", "tag4"},
					},
				},
			},
			want: map[string]map[string]string{
				"source_test_repo_hello_123_unlisted": {
					"source":      "source/test_repo-hello.123",
					"destination": "unlisted/destination/test_repo-hello.123",
					"tags":        "tag1,tag2",
				},
				"source_test_repo_hello_456_public": {
					"source":      "source/test_repo-hello.456",
					"destination": "public/destination/test_repo-hello.456",
					"tags":        "tag3,tag4",
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
			file: "retag.yml",
			want: []Retag{
				{
					Source:      "docker.io/library/alpine",
					Destination: "mirror/docker.io/library/alpine",
					Tags:        []string{"3.13", "3.14", "3.15", "3.16"},
				},
				{
					Source:      "docker.io/library/rust",
					Destination: "mirror/docker.io/library/rust",
					Tags:        []string{"1.64", "1.64-slim"},
				},
			},
		},
		{
			name:    "config with no tags is an error",
			file:    "no-tags.yml",
			wantErr: true,
		},
		{
			name:    "config with no source is an error",
			file:    "no-source.yml",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filebytes, err := os.ReadFile(filepath.Join("testdata", tt.file))
			gc := generateCmd{
				prefix: "mirror",
			}
			require.NoError(t, err)
			got, err := gc.parse(filebytes)
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
