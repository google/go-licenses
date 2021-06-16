package deps

type GoModule struct {
	// go import path, example: github.com/google/licenseclassifier/v2
	ImportPath string
	// version, example: v1.2.3, v0.0.0-20201021035429-f5854403a974
	Version string
	// local directory of dependency's source code, example on MacOS:
	// /Users/username/go/pkg/mod/github.com/!puerkito!bio/goquery@v1.6.1
	SrcDir string
}
