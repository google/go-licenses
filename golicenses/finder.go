package golicenses

import (
	"context"
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/google/licenseclassifier"
	"github.com/markbates/pkger"

	"github.com/hashicorp/go-multierror"
	"github.com/khulnasoft/go-licenses/golicenses/licenses"
)

type LicenseFinder struct {
	Paths               []string
	ConfidenceThreshold float64
	GitRemotes          []string
}

func NewLicenseFinder(paths, gitRemotes []string, threshold float64) LicenseFinder {
	return LicenseFinder{
		Paths:               paths,
		GitRemotes:          gitRemotes,
		ConfidenceThreshold: threshold,
	}
}

func licenseDBArchiveFetcher() ([]byte, error) {
	f, err := pkger.Open("/assets/licenses.db")
	if err != nil {
		return nil, fmt.Errorf("unable to open license.db: %w", err)
	}

	defer f.Close()
	return io.ReadAll(f)
}

func (r LicenseFinder) Find() (<-chan LicenseResult, error) {
	// suppress log events from go-licenses
	flag.Parse()
	_ = flag.Lookup("logtostderr").Value.Set("false")

	dbFetcherOpt := licenseclassifier.ArchiveFunc(licenseDBArchiveFetcher)
	classifier, err := licenses.NewClassifier(r.ConfidenceThreshold, dbFetcherOpt)
	if err != nil {
		return nil, err
	}

	libs, err := licenses.Libraries(context.Background(), classifier, r.Paths...)
	if err != nil {
		return nil, err
	}

	results := make(chan LicenseResult)

	go func() {
		defer close(results)
		for _, lib := range libs {
			var licenseURL, licenseName string
			var classification licenses.Type
			var errs error

			if lib.LicensePath != "" {
				licenseURL, err = findLicenseURL(lib, r.GitRemotes...)
				if err != nil {
					errs = multierror.Append(errs, fmt.Errorf("failed to locate license URL (%s): %w", lib.LicensePath, err))
					licenseURL = ""
				}

				licenseName, classification, err = classifier.Identify(lib.LicensePath)
				if err != nil {
					errs = multierror.Append(errs, fmt.Errorf("failed to identify license (%s): %w", lib.LicensePath, err))
					licenseName = ""
				}
			}

			results <- LicenseResult{
				Library: unvendor(lib.Name()),
				URL:     licenseURL,
				Path:    lib.LicensePath,
				License: licenseName,
				Type:    classification.String(),
				Errs:    errs,
			}
		}
	}()

	return results, nil
}

func findLicenseURL(lib *licenses.Library, gitRemotes ...string) (string, error) {
	// find a URL for the license file, based on the URL of a remote for the git repository.
	repo, err := licenses.FindGitRepo(lib.LicensePath)
	if err != nil {
		// can't find git repo (possibly a go module?) - derive URL from lib name instead.
		lURL, err := lib.FileURL(lib.LicensePath)
		if err != nil {
			return "", err
		}
		return lURL.String(), nil
	}

	var errs error
	for _, remote := range gitRemotes {
		url, err := repo.FileURL(lib.LicensePath, remote)
		if err != nil {
			errs = multierror.Append(errs, err)
			continue
		}
		return url.String(), errs
	}

	return "", multierror.Append(errs, fmt.Errorf("failed to find license URL"))
}

func unvendor(importPath string) string {
	// Remove the "*/vendor/" prefix from the library name for conciseness.
	if vendorerAndVendoree := strings.SplitN(importPath, "/vendor/", 2); len(vendorerAndVendoree) == 2 {
		return vendorerAndVendoree[1]
	}
	return importPath
}
