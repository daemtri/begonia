package ringhash

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParseConfig(t *testing.T) {
	tests := []struct {
		name    string
		js      string
		want    *LBConfig
		wantErr bool
	}{
		{
			name: "OK",
			js:   `{"minRingSize": 1, "maxRingSize": 2}`,
			want: &LBConfig{MinRingSize: 1, MaxRingSize: 2},
		},
		{
			name: "OK with default min",
			js:   `{"maxRingSize": 2000}`,
			want: &LBConfig{MinRingSize: defaultMinSize, MaxRingSize: 2000},
		},
		{
			name: "OK with default max",
			js:   `{"minRingSize": 2000}`,
			want: &LBConfig{MinRingSize: 2000, MaxRingSize: defaultMaxSize},
		},
		{
			name:    "min greater than max",
			js:      `{"minRingSize": 10, "maxRingSize": 2}`,
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseConfig([]byte(tt.js))
			if (err != nil) != tt.wantErr {
				t.Errorf("parseConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Errorf("parseConfig() got unexpected output, diff (-got +want): %v", diff)
			}
		})
	}
}
