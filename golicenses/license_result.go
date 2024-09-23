package golicenses

type LicenseResult struct {
	ModulePath  string
	LicensePath string
	License     string
	Type        string
	Errs        error
}
