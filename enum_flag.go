package flags

import (
	"context"
	"strings"

	"github.com/gildas/go-errors"
	"github.com/gildas/go-logger"
	"github.com/spf13/cobra"
)

// AllowedFunc is a function that returns the allowed values for a flag
type AllowedFunc func(context.Context, *cobra.Command, []string) ([]string, error)

// EnumFlag represents a flag that can only have a value from a list of allowed values
//
// If the AllowedFunc is set, the Allowed values are ignored and the function is called to get the allowed values.
type EnumFlag struct {
	Allowed     []string
	AllowedFunc AllowedFunc
	Value       string
}

// NewEnumFlag creates a new EnumFlag
//
// The default value is prepended with a +.
//
// If no default value is provided, the flag will not have a default value.
//
// If more than one default value is provided, the first one is used.
//
// Example:
//
//	flag := flags.NewEnumFlag("one", "+two", "three")
func NewEnumFlag(allowed ...string) *EnumFlag {
	var allowedValues []string
	var defaultValue string

	for _, value := range allowed {
		if strings.HasPrefix(value, "+") && defaultValue == "" {
			defaultValue = strings.TrimPrefix(value, "+")
			allowedValues = append(allowedValues, strings.TrimPrefix(value, "+"))
		} else {
			allowedValues = append(allowedValues, value)
		}
	}
	return &EnumFlag{
		Allowed: allowedValues,
		Value:   defaultValue,
	}
}

// NewEnumFlagWithFunc creates a new EnumFlag with a function to get the allowed values
func NewEnumFlagWithFunc(defaultValue string, allowedFunc AllowedFunc) *EnumFlag {
	return &EnumFlag{
		AllowedFunc: allowedFunc,
		Value:       defaultValue,
	}
}

// Type returns the type of the flag
func (flag EnumFlag) Type() string {
	return "string"
}

// String returns the string representation of the flag
func (flag EnumFlag) String() string {
	return flag.Value
}

// Set sets the flag value
func (flag *EnumFlag) Set(value string) (err error) {
	if flag.AllowedFunc != nil && len(flag.Allowed) == 0 {
		log := logger.Create("Flags", &logger.NilStream{})
		flag.Allowed, _ = flag.AllowedFunc(log.ToContext(context.Background()), nil, nil)
	}
	for _, allowed := range flag.Allowed {
		if value == allowed {
			flag.Value = value
			return nil
		}
	}
	return errors.ArgumentInvalid.With("value", value, strings.Join(flag.Allowed, ", "))
}

// CompletionFunc returns the completion function of the flag
func (flag *EnumFlag) CompletionFunc(flagName string) func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if flag.AllowedFunc != nil {
			var err error

			flag.Allowed, err = flag.AllowedFunc(cmd.Context(), cmd, args)
			if err != nil {
				return []string{}, cobra.ShellCompDirectiveError
			}
		}
		return flag.Allowed, cobra.ShellCompDirectiveDefault
	}
}
