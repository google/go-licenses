package golicenses

type LicenseResult struct {
	Library string
	URL     string
	Path    string
	License string
	Type    string
	Errs    error
}
