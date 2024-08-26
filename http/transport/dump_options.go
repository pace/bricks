package transport

import (
	"context"
	"fmt"
)

func NewDumpOptions(opts ...DumpOption) (DumpOptions, error) {
	dumpOptions := DumpOptions(map[string]bool{})
	for _, opt := range opts {
		err := opt(dumpOptions)
		if err != nil {
			return nil, err
		}
	}
	return dumpOptions, nil
}

type DumpOptions map[string]bool

func (o DumpOptions) IsEnabled(option string) bool {
	isEnabled := o[option]
	return isEnabled
}

func (o DumpOptions) AnyEnabled(options ...string) bool {
	for _, option := range options {
		if o.IsEnabled(option) {
			return true
		}
	}
	return false
}

type DumpOption func(o DumpOptions) error

func WithDumpOption(option string, enabled bool) DumpOption {
	return func(o DumpOptions) error {
		if !isDumpOptionValid(option) {
			return fmt.Errorf("invalid dump option %q", option)
		}
		o[option] = enabled
		return nil
	}
}

const (
	DumpRoundTripperOptionRequest     = "request"
	DumpRoundTripperOptionResponse    = "response"
	DumpRoundTripperOptionRequestHEX  = "request-hex"
	DumpRoundTripperOptionResponseHEX = "response-hex"
	DumpRoundTripperOptionBody        = "body"
	DumpRoundTripperOptionNoRedact    = "no-redact"
)

func isDumpOptionValid(k string) bool {
	switch k {
	case DumpRoundTripperOptionRequest,
		DumpRoundTripperOptionRequestHEX,
		DumpRoundTripperOptionResponse,
		DumpRoundTripperOptionResponseHEX,
		DumpRoundTripperOptionBody,
		DumpRoundTripperOptionNoRedact:
		return true
	default:
		return false
	}
}

func mergeDumpOptions(globalOptions, reqOptions DumpOptions) DumpOptions {
	// No request dump options, use global ones
	if len(reqOptions) == 0 {
		return globalOptions
	}

	// Request options exists, update with info from global options, but prioritize options from request over the global ones
	for globalKey, globalVal := range globalOptions {
		_, ok := reqOptions[globalKey]
		if ok {
			// req option already exists, ignore the global one
			continue
		}
		reqOptions[globalKey] = globalVal
	}
	return reqOptions
}

type dumpRoundTripperCtxKey struct{}

func CtxWithDumpRoundTripperOptions(ctx context.Context, opts DumpOptions) context.Context {
	if opts == nil {
		return ctx
	}
	return context.WithValue(ctx, dumpRoundTripperCtxKey{}, opts)
}

func DumpRoundTripperOptionsFromCtx(ctx context.Context) DumpOptions {
	do := ctx.Value(dumpRoundTripperCtxKey{})
	dumpOptions, _ := do.(DumpOptions)
	return dumpOptions
}
