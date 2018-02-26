package gmeter

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"sync"

	"github.com/seborama/govcr"
)

var (
	errNotInitialized = errors.New("gmeter is not initialized, please call /gmeter/record or /gmeter/play first")
)

type (
	//RoundTripper implements http.RoundTripper instrumented with recording and playing capabilities
	RoundTripper struct {
		http.RoundTripper

		lock    sync.RWMutex
		logger  *log.Logger
		options Options
	}

	request struct {
		Cassette string `json:"cassette"`
	}

	nopTripper struct{}
)

//RoundTrip implements http.RoundTripper
func (nt nopTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	b, err := httputil.DumpRequest(r, true)
	if err != nil {
		return nil, fmt.Errorf("failed to dump request: %v", err)
	}
	return nil, fmt.Errorf("track not found for request: %s", string(b))
}

//NewRoundTripper returns a pointer to RoundTripper struct
func NewRoundTripper(options Options, logger *log.Logger) *RoundTripper {
	return &RoundTripper{options: options, logger: logger}
}

//RoundTrip implements http.RoundTripper
func (rt *RoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	rt.lock.RLock()
	defer rt.lock.RUnlock()

	if rt.RoundTripper == nil {
		return nil, errNotInitialized
	}

	resp, err := rt.RoundTripper.RoundTrip(r)
	if resp != nil {
		rt.logger.Printf("%s %s %d", r.Method, r.URL, resp.StatusCode)
	}

	return resp, err
}

//Record starts recording of a cassette
func (rt *RoundTripper) Record(w http.ResponseWriter, r *http.Request) {
	rt.lock.Lock()
	defer rt.lock.Unlock()

	req := request{}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&req); err != nil {
		rt.logger.Printf("failed to decode record request: %v", err)
		return
	}

	config := govcr.VCRConfig{
		DisableRecording: false,
		CassettePath:     rt.options.CassettePath,
		Client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: rt.options.Insecure,
				},
			},
		},
	}

	rt.RoundTripper = govcr.NewVCR(req.Cassette, &config).Client.Transport
	rt.logger.Printf("started recording of the cassette: %s", req.Cassette)
}

//Play stops recording and starts playing a cassette
func (rt *RoundTripper) Play(w http.ResponseWriter, r *http.Request) {
	rt.lock.Lock()
	defer rt.lock.Unlock()

	req := request{}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&req); err != nil {
		rt.logger.Printf("failed to decode play request: %v", err)
		return
	}

	config := govcr.VCRConfig{
		DisableRecording: true,
		CassettePath:     rt.options.CassettePath,
		Client: &http.Client{
			Transport: nopTripper{},
		},
	}

	rt.RoundTripper = govcr.NewVCR(req.Cassette, &config).Client.Transport
	rt.logger.Printf("started playing the cassette: %s", req.Cassette)
}
