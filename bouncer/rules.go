package bouncer

import (
	"fmt"
	"regexp"
)

const (
	UnknownAction Action = iota
	AllowAction
	DenyAction
)

var actionStr = []string{
	"UnknownAction",
	"Allow",
	"Deny",
}

type Action int

type Rules struct {
	Action     Action
	Patterns   []*regexp.Regexp
	IgnorePkgs []*regexp.Regexp
}

func NewRules(act Action, patterns []string, ignore ...string) (Rules, error) {
	if act == UnknownAction {
		return Rules{}, fmt.Errorf("bad action given: %+v", act)
	}

	regexPatterns := make([]*regexp.Regexp, len(patterns))
	for idx, a := range patterns {
		pattern, err := regexp.Compile(a)
		if err != nil {
			return Rules{}, fmt.Errorf("bad rule (%s): %w", a, err)
		}
		regexPatterns[idx] = pattern
	}

	ignorePatterns := make([]*regexp.Regexp, len(ignore))
	for idx, a := range ignore {
		pattern, err := regexp.Compile(a)
		if err != nil {
			return Rules{}, fmt.Errorf("bad ignore pattern (%s): %w", a, err)
		}
		ignorePatterns[idx] = pattern
	}

	return Rules{
		Action:     act,
		Patterns:   regexPatterns,
		IgnorePkgs: ignorePatterns,
	}, nil
}

func (r Rules) Evaluate(results ...LicenseResult) (bool, []LicenseResult, error) {
	mismatched := make([]LicenseResult, 0)
	matched := make([]LicenseResult, 0)
resultsLoop:
	for _, result := range results {
		licenseName := result.License
		libName := result.ModulePath

		for _, i := range r.IgnorePkgs {
			if i.Match([]byte(libName)) {
				continue resultsLoop
			}
		}

		match := false
	patternLoop:
		for _, p := range r.Patterns {
			if p.Match([]byte(licenseName)) {
				match = true
				break patternLoop
			}
		}
		if match {
			matched = append(matched, result)
		} else {
			mismatched = append(mismatched, result)
		}
	}

	switch r.Action {
	case AllowAction:
		return len(mismatched) == 0, mismatched, nil
	case DenyAction:
		return len(matched) == 0, matched, nil
	}
	return false, nil, fmt.Errorf("could not evaluate action: %s", r.Action)
}

func (o Action) String() string {
	if int(o) >= len(actionStr) || o < 0 {
		return actionStr[0]
	}

	return actionStr[o]
}
