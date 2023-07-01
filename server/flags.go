package server

import (
	"github.com/apex/log"
	"github.com/spf13/pflag"
)

type flags struct {
	ListenAddress string
	Taskfile      string

	EnableHistory  bool
	HistoryOutput  string
	PrintBuildInfo bool
}

func (f *flags) Bind() {
	pflag.StringVar(&f.ListenAddress, "l", ":3000", "Listen address for task-ui")
	pflag.StringVar(&f.Taskfile, "f", "Taskfile.yml", "Taskfile filename to load")
	pflag.BoolVar(&f.EnableHistory, "history-enable", false, "Enable history with ovh-ttyrecord")
	pflag.StringVar(&f.HistoryOutput, "history-output", "./history", "History output folder")
	pflag.BoolVarP(&f.PrintBuildInfo, "version", "v", false, "Print build info")
	pflag.Parse()
}

// Validate will evaluate *flags and modify them, return an error if any occurs.
// From then on, the *flags object is ready to use. This lets us have computed
// fields, and upgrade deprecated flags to newer fields...
func (f *flags) Validate() error {
	if f.EnableHistory {
		log.Infof("History is enabled, writing to %q", f.HistoryOutput)
	}
	return nil
}
