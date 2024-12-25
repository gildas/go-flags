package flags

import (
	"strings"

	"github.com/gildas/go-core"
	"github.com/gildas/go-errors"
	"github.com/spf13/cobra"
)

// EnumSliceFlag represents a flag that can only have values from a list of allowed values
//
// The flag can be repeated to have multiple values.
type EnumSliceFlag struct {
	Allowed     []string
	Values      []string
	Default     []string
	AllowedFunc AllowedFunc
	AllAllowed  bool
	all         bool
}

// Type returns the type of the flag
//
// implements pflag.Value
func (flag EnumSliceFlag) Type() string {
	return "stringSlice"
}

// NewEnumSliceFlag creates a new EnumSliceFlag
//
// The default values are prepended with a +.
//
// If no default value is provided, the flag will not have a default value.
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

// NewEnumSliceFlagWithFunc creates a new EnumSliceFlag
func NewEnumSliceFlagWithFunc(allowedFunc AllowedFunc, defaultvalues ...string) *EnumSliceFlag {
	return &EnumSliceFlag{
		AllowedFunc: allowedFunc,
		Default:     append([]string{}, defaultvalues...),
	}
}

// NewEnumSliceFlagWithAllAllowed creates a new EnumSliceFlag
//
// The default values are prepended with a +.
//
// If the flag is set to "all", all the allowed values are set.
//
// If no default value is provided, the flag will not have a default value.
//
// Example:
//
//	flag := flags.NewEnumSliceFlagWithAllAllowed("+one", "+two", "three")
func NewEnumSliceFlagWithAllAllowed(allowed ...string) *EnumSliceFlag {
	flag := NewEnumSliceFlag(allowed...)
	flag.AllAllowed = true
	return flag
}

// NewEnumSliceFlagWithAllAllowedAndFunc creates a new EnumSliceFlag
func NewEnumSliceFlagWithAllAllowedAndFunc(allowedFunc AllowedFunc, defaultvalues ...string) *EnumSliceFlag {
	flag := NewEnumSliceFlagWithFunc(allowedFunc, defaultvalues...)
	flag.AllAllowed = true
	return flag
}

// String returns the string representation of the flag
//
// implements fmt.Stringer and pflag.Value
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
//
// implements pflag.Value
func (flag *EnumSliceFlag) Set(value string) (err error) {
	if flag.AllowedFunc != nil && len(flag.Allowed) == 0 { // Unfortunatly, as of now, we cannot call the function to get the allowed values
		// TODO: Find a way to call the function to get the allowed values
		return flag.Append(value) // so we just add the value
	}
	if value == "all" && flag.AllAllowed {
		flag.Values = flag.Allowed
		flag.all = true
		return nil
	}
	found := false
	for _, v := range strings.Split(value, ",") {
		for _, allowed := range flag.Allowed {
			if v == allowed {
				found = true
				if !core.Contains(flag.Values, v) {
					flag.Values = append(flag.Values, v)
				}
			}
		}
	}
	if found {
		return nil
	}
	return errors.ArgumentInvalid.With("value", value, strings.Join(flag.Allowed, ", "))
}

// Append appends a value to the flag
//
// implements pflag.SliceValue
func (flag *EnumSliceFlag) Append(value string) error {
	for _, v := range strings.Split(value, ",") {
		if !core.Contains(flag.Values, v) {
			flag.Values = append(flag.Values, v)
		}
	}
	return nil
}

// Replace replaces the flag values with the given values
//
// implements pflag.SliceValue
func (flag *EnumSliceFlag) Replace(values []string) error {
	flag.Values = make([]string, 0, len(values))
	for _, value := range values {
		_ = flag.Append(value)
	}
	return nil
}

// GetSlice returns the flag value list as a slice of strings
//
// implements pflag.SliceValue
func (flag EnumSliceFlag) GetSlice() []string {
	if len(flag.Values) == 0 {
		// TODO: Find a way to call the function to get the allowed values to build the default values (if any)
		return flag.Default
	}
	if flag.all {
		return append(append([]string{}, "all"), flag.Values...)
	}
	return flag.Values
}

// CompletionFunc returns the completion function of the flag
//
// This function is used by the cobra.Command when it needs to complete the flag value.
//
// See: https://pkg.go.dev/github.com/spf13/cobra#Command.RegisterFlagCompletionFunc
func (flag EnumSliceFlag) CompletionFunc(flagName string) (string, func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective)) {
	return flagName, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		var allowed []string
		var err error

		if flag.AllowedFunc != nil {
			allowed, err = flag.AllowedFunc(cmd.Context(), cmd, args, toComplete)
			if err != nil {
				return []string{}, cobra.ShellCompDirectiveError
			}
		} else {
			allowed = make([]string, 0, len(flag.Allowed))
			if current, err := cmd.Flags().GetStringSlice(flagName); err == nil {
				for _, value := range flag.Allowed {
					if !core.Contains(current, value) {
						allowed = append(allowed, value)
					}
				}
			}
		}
		if flag.AllAllowed && len(allowed) > 0 {
			allowed = append(allowed, "all")
		}
		return allowed, cobra.ShellCompDirectiveDefault
	}
}
