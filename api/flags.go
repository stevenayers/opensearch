package api

import "flag"

type (
	// Flags holds the app Flags
	Flags struct {
		ConfigFile *string
		Port       *int
		Verbose    *bool
	}
)

var (
	// AppFlags makes a global Flag struct
	AppFlags Flags
)

// InitFlags loads flags into global var AppFlags
func InitFlags(appFlags *Flags) {
	appFlags.ConfigFile = flag.String("config", "../cmd/Config.toml", "Config file path")
	appFlags.Port = flag.Int("port", 0, "Port to listen on")
	appFlags.Verbose = flag.Bool("verbose", false, "Verbosity")
	flag.Parse()
}
