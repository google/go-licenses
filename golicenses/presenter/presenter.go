package presenter

import (
	"io"

	"github.com/khulnasoft/go-licenses/golicenses"
	"github.com/khulnasoft/go-licenses/golicenses/presenter/csv"
	"github.com/khulnasoft/go-licenses/golicenses/presenter/json"
	"github.com/khulnasoft/go-licenses/golicenses/presenter/text"
)

type Presenter interface {
	Present(io.Writer) error
}

func GetPresenter(option Option, results <-chan golicenses.LicenseResult) Presenter {
	switch option {
	case CSVPresenter:
		return csv.NewPresenter(results)
	case JSONPresenter:
		return json.NewPresenter(results)
	case TextPresenter:
		return text.NewPresenter(results)

	default:
		return nil
	}
}
