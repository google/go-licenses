package gocli

import "time"

// Module provides module information for a package.
type Module struct {
	Path      string     // module path
	Version   string     // module version
	Time      *time.Time // time version was created
	Main      bool       // is this the main module?
	Indirect  bool       // is this module only an indirect dependency of main module?
	Dir       string     // directory holding files for this module, if any
	GoMod     string     // path to go.mod file used when loading this module, if any
	GoVersion string     // go version used in module
}
