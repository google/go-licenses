package presenter

import (
	"io"

	"github.com/sulaiman-coder/gobouncer/bouncer"
	"github.com/sulaiman-coder/gobouncer/bouncer/presenter/csv"
	"github.com/sulaiman-coder/gobouncer/bouncer/presenter/json"
	"github.com/sulaiman-coder/gobouncer/bouncer/presenter/text"
)

type Presenter interface {
	Present(io.Writer) error
}

func GetPresenter(option Option, results []bouncer.LicenseResult) Presenter {
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
