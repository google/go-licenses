package presenter

import "strings"

const (
	UnknownPresenter Option = iota
	CSVPresenter
	JSONPresenter
	TextPresenter
)

var optionStr = []string{
	"UnknownPresenter",
	"csv",
	"json",
	"text",
}

var Options = []Option{
	CSVPresenter,
	JSONPresenter,
	TextPresenter,
}

type Option int

func ParseOption(userStr string) Option {
	switch strings.ToLower(userStr) {
	case strings.ToLower(CSVPresenter.String()):
		return CSVPresenter
	case strings.ToLower(JSONPresenter.String()):
		return JSONPresenter
	case strings.ToLower(TextPresenter.String()):
		return TextPresenter
	default:
		return UnknownPresenter
	}
}

func (o Option) String() string {
	if int(o) >= len(optionStr) || o < 0 {
		return optionStr[0]
	}

	return optionStr[o]
}
