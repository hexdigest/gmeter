package gmeter

import (
	"io"
	"io/ioutil"
	"net/url"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetOptions(t *testing.T) {
	type args struct {
		arguments []string
		stdout    io.Writer
		stderr    io.Writer
		exit      exitFunc
	}
	tests := []struct {
		name string
		args func(t *testing.T) args

		want1 Options
	}{
		{
			name: "show help",
			args: func(t *testing.T) args {
				return args{
					arguments: []string{"-h"},
					exit: func(code int) {
						assert.Equal(t, 0, code)
						t.Skip()
					},
				}
			},
		},
		{
			name: "invalid target",
			args: func(t *testing.T) args {
				return args{
					arguments: []string{""},
					stderr:    ioutil.Discard,
					exit: func(code int) {
						assert.Equal(t, 2, code)
						t.Skip()
					},
				}
			},
		},
		{
			name: "success",
			args: func(t *testing.T) args {
				return args{
					arguments: []string{"-t", "http://github.com"},
				}
			},
			want1: Options{
				CassettePath:  ".",
				Insecure:      false,
				ListenAddress: "localhost:8080",
				TargetURL:     &url.URL{Scheme: "http", Host: "github.com"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tArgs := tt.args(t)

			got1 := GetOptions(tArgs.arguments, tArgs.stdout, tArgs.stderr, tArgs.exit)

			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("GetOptions got1 = %v, want1: %v", got1, tt.want1)
			}
		})
	}
}
