package parser

import (
	"fmt"

	"github.com/khulnasoft/go-licenses/golicenses/event"

	"github.com/khulnasoft/go-licenses/golicenses/presenter"
	"github.com/khulnasoft/go-progress"
	"github.com/khulnasoft/go-pulsebus"
)

type ErrBadPayload struct {
	Type  pulsebus.EventType
	Field string
	Value interface{}
}

func (e *ErrBadPayload) Error() string {
	return fmt.Sprintf("event='%s' has bad event payload field='%v': '%+v'", string(e.Type), e.Field, e.Value)
}

func newPayloadErr(t pulsebus.EventType, field string, value interface{}) error {
	return &ErrBadPayload{
		Type:  t,
		Field: field,
		Value: value,
	}
}

func checkEventType(actual, expected pulsebus.EventType) error {
	if actual != expected {
		return newPayloadErr(expected, "Type", actual)
	}
	return nil
}

func ParseModuleScanStarted(e pulsebus.Event) (progress.StagedProgressable, error) {
	if err := checkEventType(e.Type, event.ModuleScanStarted); err != nil {
		return nil, err
	}

	p, ok := e.Value.(progress.StagedProgressable)
	if !ok {
		return nil, newPayloadErr(e.Type, "Value", e.Value)
	}

	return p, nil
}

func ParseModuleScanResult(e pulsebus.Event) (presenter.Presenter, error) {
	if err := checkEventType(e.Type, event.ModuleScanResult); err != nil {
		return nil, err
	}

	pres, ok := e.Value.(presenter.Presenter)
	if !ok {
		return nil, newPayloadErr(e.Type, "Value", e.Value)
	}

	return pres, nil
}
