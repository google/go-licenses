package csv

import (
	"encoding/csv"
	"io"

	"github.com/sulaiman-coder/gobouncer/bouncer"
)

type Presenter struct {
	resultStream <-chan bouncer.LicenseResult
}

func NewPresenter(results <-chan bouncer.LicenseResult) Presenter {
	return Presenter{
		resultStream: results,
	}
}

func (p Presenter) Present(target io.Writer) error {
	writer := csv.NewWriter(target)
	for result := range p.resultStream {
		if err := writer.Write([]string{result.Library, result.URL, result.Type, result.License}); err != nil {
			return err
		}
	}
	writer.Flush()
	return writer.Error()
}
