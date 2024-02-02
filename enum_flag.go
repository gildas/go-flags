package flags

import (
	"context"
	"strings"

	"github.com/gildas/go-errors"
	"github.com/spf13/cobra"
)

type EnumFlag struct {
	Allowed     []string
	AllowedFunc func(context.Context, *cobra.Command, []string) []string
	Value       string
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
func (flag *EnumFlag) Set(value string) error {
	if flag.AllowedFunc != nil {
		flag.Value = value
		return nil
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
func (flag EnumFlag) CompletionFunc() func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if flag.AllowedFunc != nil {
			return flag.AllowedFunc(cmd.Context(), cmd, args), cobra.ShellCompDirectiveNoFileComp
		}
		return flag.Allowed, cobra.ShellCompDirectiveNoFileComp
	}
}
