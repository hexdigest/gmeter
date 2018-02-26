package gmeter

import (
	"flag"
	"fmt"
	"io"
	"net/url"
)

//Options contains parsed command line options
type Options struct {
	CassettePath  string
	ListenAddress string
	TargetURL     *url.URL
	Insecure      bool
}

type exitFunc func(int)

//GetOptions parses arguments and returns Options struct on success, otherwise
//writes error message to the stderr writer and calls exit function
func GetOptions(arguments []string, stdout, stderr io.Writer, exit exitFunc) Options {
	var (
		flagset  = flag.NewFlagSet("gmeter", flag.ExitOnError)
		listen   = flagset.String("l", "localhost:8080", "listen address")
		target   = flagset.String("t", "", "target base URL")
		dir      = flagset.String("d", ".", "cassettes dir")
		help     = flagset.Bool("h", false, "display this help text and exit")
		insecure = flagset.Bool("insecure", false, "skip HTTPs checks")
	)

	flagset.Parse(arguments)

	if *help {
		flagset.SetOutput(stdout)
		flagset.Usage()
		exit(0)
	}

	var errors []string

	if *target == "" {
		errors = append(errors, "missing target base URL: -t")
	}

	targetURL, err := url.Parse(*target)
	if err != nil {
		errors = append(errors, fmt.Sprintf("failed to parse target URL: %v", err))
	}

	if targetURL.Scheme != "http" && targetURL.Scheme != "https" {
		errors = append(errors, fmt.Sprintf("unsupported scheme: %q", targetURL.Scheme))
	}

	if len(errors) > 0 {
		for _, e := range errors {
			fmt.Fprintf(stderr, "%s\n", e)
		}
		flagset.Usage()
		exit(2)
	}

	return Options{
		CassettePath:  *dir,
		Insecure:      *insecure,
		ListenAddress: *listen,
		TargetURL:     targetURL,
	}
}
