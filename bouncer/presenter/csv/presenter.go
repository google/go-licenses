package csv

import (
	"encoding/csv"
	"io"

	"github.com/sulaiman-coder/gobouncer/bouncer"
)

type Presenter struct {
	resultStream []bouncer.LicenseResult
}

func NewPresenter(results []bouncer.LicenseResult) Presenter {
	return Presenter{
		resultStream: results,
	}
}

func (p Presenter) Present(target io.Writer) error {
	writer := csv.NewWriter(target)
	for _, result := range p.resultStream {
		if err := writer.Write([]string{result.ModulePath, result.Type, result.License}); err != nil {
			return err
		}
	}
	writer.Flush()
	return writer.Error()
}
