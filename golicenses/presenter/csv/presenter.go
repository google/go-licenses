package csv

import (
	"encoding/csv"
	"io"

	"github.com/khulnasoft/go-licenses/golicenses"
)

type Presenter struct {
	resultStream []golicenses.LicenseResult
}

func NewPresenter(results []golicenses.LicenseResult) Presenter {
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
