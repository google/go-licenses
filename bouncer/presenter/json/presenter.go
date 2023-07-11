package json

import (
	"encoding/json"
	"io"

	"github.com/hashicorp/go-multierror"
	"github.com/sulaiman-coder/gobouncer/bouncer"
)

type jsonResult struct {
	Pkg string `json:"package"`
	URL string `json:"url"`
	// Path     string   `json:"local-path"`
	Name     string   `json:"name"`
	Type     string   `json:"type"`
	Warnings []string `json:"warnings,omitempty"`
}

type Presenter struct {
	resultStream <-chan bouncer.LicenseResult
}

func NewPresenter(results <-chan bouncer.LicenseResult) Presenter {
	return Presenter{
		resultStream: results,
	}
}

func unwrap(err error) []error {
	errs := make([]error, 0)
	if err == nil {
		return errs
	}

	if mErr, ok := err.(*multierror.Error); ok {
		if mErr == nil {
			return errs
		}
		for _, err := range mErr.Errors {
			errs = append(errs, unwrap(err)...)
		}
	} else {
		errs = append(errs, err)
	}
	return errs
}

func (p Presenter) Present(target io.Writer) error {
	writer := json.NewEncoder(target)
	writer.SetEscapeHTML(false)
	writer.SetIndent("", "  ")

	results := make([]jsonResult, 0)
	for result := range p.resultStream {
		warnings := make([]string, 0)
		if result.Errs != nil {
			for _, err := range unwrap(result.Errs) {
				warnings = append(warnings, err.Error())
			}
		}
		results = append(results, jsonResult{
			Pkg:  result.Library,
			URL:  result.URL,
			Name: result.License,
			Type: result.Type,
			//Path:     result.Path,
			Warnings: warnings,
		})
	}

	return writer.Encode(&results)
}
