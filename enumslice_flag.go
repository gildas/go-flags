package flags

import (
	"strings"

	"github.com/gildas/go-core"
	"github.com/gildas/go-errors"
	"github.com/spf13/cobra"
)

type EnumSliceFlag struct {
	Allowed    []string
	Values     []string
	Default    []string
	AllAllowed bool
	all        bool
}

// Type returns the type of the flag
func (flag EnumSliceFlag) Type() string {
	return "stringSlice"
}

// NewEnumSliceFlag creates a new EnumSliceFlag
//
// The default values are prepended with a +
//
// # If no default value is provided, the flag will not have a default value
//
// Example:
//
//	flag := flags.NewEnumSliceFlag("+one", "+two", "three")
func NewEnumSliceFlag(allowed ...string) *EnumSliceFlag {
	var allowedValues []string
	var defaultValues []string

	for _, value := range allowed {
		if strings.HasPrefix(value, "+") {
			defaultValues = append(defaultValues, strings.TrimPrefix(value, "+"))
			allowedValues = append(allowedValues, strings.TrimPrefix(value, "+"))
		} else {
			allowedValues = append(allowedValues, value)
		}
	}
	return &EnumSliceFlag{
		Allowed: allowedValues,
		Default: defaultValues,
	}
}

// NewEnumSliceFlagWithAllAllowed creates a new EnumSliceFlag
//
// The default values are prepended with a +
//
// # If no default value is provided, the flag will not have a default value
//
// Example:
//
//	flag := flags.NewEnumSliceFlag("+one", "+two", "three")
func NewEnumSliceFlagWithAllAllowed(allowed ...string) *EnumSliceFlag {
	flag := NewEnumSliceFlag(allowed...)
	flag.AllAllowed = true
	return flag
}

// String returns the string representation of the flag
func (flag EnumSliceFlag) String() string {
	var result strings.Builder

	result.WriteString("[")
	for i, value := range flag.Values {
		if i > 0 {
			result.WriteString(",")
		}
		result.WriteString(value)
	}
	result.WriteString("]")
	return result.String()
}

// Set sets the flag value
func (flag *EnumSliceFlag) Set(value string) error {
	if value == "all" && flag.AllAllowed {
		flag.Values = flag.Allowed
		flag.all = true
		return nil
	}
	for _, allowed := range flag.Allowed {
		if value == allowed {
			for _, existing := range flag.Values {
				if existing == value {
					return nil
				}
			}
			flag.Values = append(flag.Values, value)
			return nil
		}
	}
	return errors.ArgumentInvalid.With("value", value, strings.Join(flag.Allowed, ", "))
}

// Get returns the flag value
func (flag EnumSliceFlag) Get() []string {
	if len(flag.Values) == 0 {
		return flag.Default
	}
	return flag.Values
}

// Contains returns true if the flag contains the given value
func (flag EnumSliceFlag) Contains(value string) bool {
	if flag.all && value == "all" {
		return true
	}
	if !core.Contains(flag.Allowed, value) {
		return false
	}
	if flag.all {
		return true
	}
	return core.Contains(flag.Values, value)
}

// CompletionFunc returns the completion function of the flag
func (flag EnumSliceFlag) CompletionFunc() func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
		return flag.Allowed, cobra.ShellCompDirectiveDefault
	}
}
