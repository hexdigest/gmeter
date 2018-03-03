package gmeter

import (
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

type errorReader struct{}

func (er errorReader) Read([]byte) (int, error) {
	return 0, errors.New("bad reader")
}

func Test_nopTripper_RoundTrip(t *testing.T) {
	type args struct {
		r *http.Request
	}
	tests := []struct {
		name    string
		init    func(t *testing.T) nopTripper
		inspect func(r nopTripper, t *testing.T) //inspects receiver after test run

		args func(t *testing.T) args

		want1      *http.Response
		wantErr    bool
		inspectErr func(err error, t *testing.T) //use for more precise error evaluation after test
	}{
		{
			name: "failed to dump request",
			init: func(*testing.T) nopTripper { return nopTripper{} },
			args: func(*testing.T) args {
				return args{r: httptest.NewRequest("POST", "http://github.com/hexdigest/gmeter", errorReader{})}
			},
			wantErr: true,
			inspectErr: func(err error, t *testing.T) {
				if !strings.Contains(err.Error(), "failed to dump request") {
					t.Errorf("unexpected error: %v", err)
				}
			},
		},
		{
			name: "track not found",
			init: func(*testing.T) nopTripper { return nopTripper{} },
			args: func(*testing.T) args {
				return args{r: httptest.NewRequest("POST", "http://github.com/hexdigest/gmeter", strings.NewReader(""))}
			},
			wantErr: true,
			inspectErr: func(err error, t *testing.T) {
				if !strings.Contains(err.Error(), "track not found") {
					t.Errorf("unexpected error: %v", err)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tArgs := tt.args(t)

			receiver := tt.init(t)
			got1, err := receiver.RoundTrip(tArgs.r)

			if tt.inspect != nil {
				tt.inspect(receiver, t)
			}

			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("nopTripper.RoundTrip got1 = %v, want1: %v", got1, tt.want1)
			}

			if (err != nil) != tt.wantErr {
				t.Fatalf("nopTripper.RoundTrip error = %v, wantErr: %t", err, tt.wantErr)
			}

			if tt.inspectErr != nil {
				tt.inspectErr(err, t)
			}
		})
	}
}

type checkStatusWriter struct {
	http.ResponseWriter

	expectedCode int
	t            *testing.T
}

func newCheckStatusWriter(t *testing.T, code int) checkStatusWriter {
	return checkStatusWriter{expectedCode: code, t: t}
}

func (c checkStatusWriter) WriteHeader(code int) {
	if code != c.expectedCode {
		c.t.Errorf("unexpected status code, got: %d, want: %d", code, c.expectedCode)
	}
}

func Test_RoundTripper_Record(t *testing.T) {
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name    string
		init    func(t *testing.T) *RoundTripper
		inspect func(r *RoundTripper, t *testing.T) //inspects receiver after test run

		args func(t *testing.T) args
	}{
		{
			name: "bad request",
			init: func(*testing.T) *RoundTripper {
				return &RoundTripper{logger: log.New(ioutil.Discard, "", 0)}
			},
			args: func(t *testing.T) args {
				return args{
					r: httptest.NewRequest("POST", "https://github.com/hexdigest/gmeter", strings.NewReader("{")),
					w: newCheckStatusWriter(t, 400),
				}
			},
		},
		{
			name: "success",
			init: func(*testing.T) *RoundTripper {
				return &RoundTripper{logger: log.New(ioutil.Discard, "", 0)}
			},
			args: func(t *testing.T) args {
				body := strings.NewReader(`{"cassette": "nice music"}`)
				return args{r: httptest.NewRequest("POST", "https://github.com/hexdigest/gmeter", body)}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tArgs := tt.args(t)

			receiver := tt.init(t)
			receiver.Record(tArgs.w, tArgs.r)

			if tt.inspect != nil {
				tt.inspect(receiver, t)
			}

		})
	}
}

func Test_decodeRequest(t *testing.T) {
	type args struct {
		r io.Reader
	}
	tests := []struct {
		name string
		args func(t *testing.T) args

		want1      *request
		wantErr    bool
		inspectErr func(err error, t *testing.T) //use for more precise error evaluation after test
	}{
		{
			name: "failed to decode request",
			args: func(t *testing.T) args {
				return args{r: strings.NewReader("{")}
			},
			wantErr: true,
		},
		{
			name: "empty cassette name",
			args: func(t *testing.T) args {
				return args{r: strings.NewReader("{}")}
			},
			wantErr: true,
		},
		{
			name: "nice music",
			args: func(t *testing.T) args {
				return args{r: strings.NewReader(`{"cassette": "nice music"}`)}
			},
			want1: &request{Cassette: "nice music"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tArgs := tt.args(t)

			got1, err := decodeRequest(tArgs.r)

			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("decodeRequest got1 = %v, want1: %v", got1, tt.want1)
			}

			if (err != nil) != tt.wantErr {
				t.Fatalf("decodeRequest error = %v, wantErr: %t", err, tt.wantErr)
			}

			if tt.inspectErr != nil {
				tt.inspectErr(err, t)
			}
		})
	}
}

func Test_RoundTripper_Play(t *testing.T) {
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name    string
		init    func(t *testing.T) *RoundTripper
		inspect func(r *RoundTripper, t *testing.T) //inspects receiver after test run

		args func(t *testing.T) args
	}{
		{
			name: "bad request",
			init: func(*testing.T) *RoundTripper {
				return &RoundTripper{logger: log.New(ioutil.Discard, "", 0)}
			},
			args: func(t *testing.T) args {
				r, _ := http.NewRequest("POST", "https://github.com/hexdigest/gmeter", strings.NewReader("{"))
				return args{
					r: r,
					w: newCheckStatusWriter(t, 400),
				}
			},
		},
		{
			name: "success",
			init: func(*testing.T) *RoundTripper {
				return &RoundTripper{logger: log.New(ioutil.Discard, "", 0)}
			},
			args: func(t *testing.T) args {
				body := strings.NewReader(`{"cassette": "nice music"}`)
				r, _ := http.NewRequest("POST", "https://github.com/hexdigest/gmeter", body)
				return args{r: r}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tArgs := tt.args(t)

			receiver := tt.init(t)
			receiver.Play(tArgs.w, tArgs.r)

			if tt.inspect != nil {
				tt.inspect(receiver, t)
			}
		})
	}
}

func TestNewRoundTripper(t *testing.T) {
	expected := RoundTripper{}
	rt := NewRoundTripper(Options{}, nil)
	if *rt != expected {
		t.Errorf("expected pointer to empty RoundTripper, got: %v", *rt)
	}
}

type roundTripperMock struct {
	resp *http.Response
	err  error
}

func (rt roundTripperMock) RoundTrip(*http.Request) (*http.Response, error) {
	return rt.resp, rt.err
}

func TestRoundTripper_RoundTrip(t *testing.T) {
	type args struct {
		r *http.Request
	}
	tests := []struct {
		name    string
		init    func(t *testing.T) *RoundTripper
		inspect func(r *RoundTripper, t *testing.T) //inspects receiver after test run

		args func(t *testing.T) args

		want1      *http.Response
		wantErr    bool
		inspectErr func(err error, t *testing.T) //use for more precise error evaluation after test
	}{
		{
			name: "not initialized",
			init: func(t *testing.T) *RoundTripper { return &RoundTripper{} },
			args: func(t *testing.T) args {
				return args{r: httptest.NewRequest("POST", "http://github.com/hexdigest/gmeter", strings.NewReader(""))}
			},
			wantErr: true,
			inspectErr: func(err error, t *testing.T) {
				if err != errNotInitialized {
					t.Errorf("unexpected error: %v", err)
				}
			},
		},
		{
			name: "success",
			init: func(t *testing.T) *RoundTripper {
				rtMock := roundTripperMock{resp: &http.Response{StatusCode: http.StatusTeapot}, err: nil}
				return &RoundTripper{
					RoundTripper: rtMock,
					logger:       log.New(ioutil.Discard, "", 0),
				}
			},
			args: func(t *testing.T) args {
				return args{r: httptest.NewRequest("POST", "http://github.com/hexdigest/gmeter", strings.NewReader(""))}
			},
			want1: &http.Response{StatusCode: http.StatusTeapot},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tArgs := tt.args(t)

			receiver := tt.init(t)
			got1, err := receiver.RoundTrip(tArgs.r)

			if tt.inspect != nil {
				tt.inspect(receiver, t)
			}

			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("RoundTripper.RoundTrip got1 = %v, want1: %v", got1, tt.want1)
			}

			if (err != nil) != tt.wantErr {
				t.Fatalf("RoundTripper.RoundTrip error = %v, wantErr: %t", err, tt.wantErr)
			}

			if tt.inspectErr != nil {
				tt.inspectErr(err, t)
			}
		})
	}
}
