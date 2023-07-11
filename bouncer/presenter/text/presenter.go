package text

import (
	"bufio"
	"fmt"
	"io"
	"sort"

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
	writer := bufio.NewWriter(target)
	results := make([]string, 0)
	for _, result := range p.resultStream {
		str := fmt.Sprintf("%-60s %-20s %-s", result.ModulePath, result.License, result.Type)
		results = append(results, str)
	}

	sort.Strings(results)

	header := fmt.Sprintf("%-60s %-20s %-s", "PACKAGE", "LICENSE", "TYPE")
	underline := fmt.Sprintf("%-60s %-20s %-s", "-------", "-------", "----")
	if _, err := writer.WriteString(header + "\n"); err != nil {
		return err
	}
	if _, err := writer.WriteString(underline + "\n"); err != nil {
		return err
	}
	for _, result := range results {
		if _, err := writer.WriteString(result + "\n"); err != nil {
			return err
		}
	}

	return writer.Flush()
}
