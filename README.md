# go-flags

![GoVersion](https://img.shields.io/github/go-mod/go-version/gildas/go-flags)
[![GoDoc](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/github.com/gildas/go-flags)
[![License](https://img.shields.io/github/license/gildas/go-flags)](https://github.com/gildas/go-flags/blob/master/LICENSE)
[![Report](https://goreportcard.com/badge/github.com/gildas/go-flags)](https://goreportcard.com/report/github.com/gildas/go-flags)  

![master](https://img.shields.io/badge/branch-master-informational)
[![Test](https://github.com/gildas/go-flags/actions/workflows/test.yml/badge.svg?branch=master)](https://github.com/gildas/go-flags/actions/workflows/test.yml)
[![codecov](https://codecov.io/gh/gildas/go-flags/branch/master/graph/badge.svg?token=gFCzS9b7Mu)](https://codecov.io/gh/gildas/go-flags/branch/master)

![dev](https://img.shields.io/badge/branch-dev-informational)
[![Test](https://github.com/gildas/go-flags/actions/workflows/test.yml/badge.svg?branch=dev)](https://github.com/gildas/go-flags/actions/workflows/test.yml)
[![codecov](https://codecov.io/gh/gildas/go-flags/branch/dev/graph/badge.svg?token=gFCzS9b7Mu)](https://codecov.io/gh/gildas/go-flags/branch/dev)

go-flags is a library that provides [pflag.FlagSet](https://pkg.go.dev/github.com/spf13/pflag#FlagSet) objects for cobra commands.

## Installation

This package is available through `go get`:

```console
go get github.com/gildas/go-flags
```

## Usage

### EnumFlag

You can use the `EnumFlag` to define a flag that can only take a set of predefined values:

```go
cmd := &cobra.Command{
    Use: "myapp",
    . . .
}

state := flags.NewEnumFlag("+one", "two", "three")
cmd.Flags().Var(state, "state", "State of the flag")
_ = cmd.RegisterFlagCompletionFunc(state.CompletionFunc("state"))
```

As you can see, the `EnumFlag` is created with a list of strings that are the only values the flag can take. The `RegisterFlagCompletionFunc` is used to provide completion for the flag.

The default value is prepended with a `+`.

Instead of values, you can provided instead a function that will be called to get the list of allowed values:

```go
cmd := &cobra.Command{
    Use: "myapp",
    . . .
}

state := flags.NewEnumFlagWithFunc("one", func(context.Context, *cobra.Command, []string) []string {
    return []string{"one", "two", "three"}
})
cmd.Flags().Var(state, "state", "State of the flag")
_ = cmd.RegisterFlagCompletionFunc(state.CompletionFunc("state"))
```

Note that the default value is not prepended with a `+` in this case.

### EnumSliceFlag

You can use the `EnumSliceFlag` to define a flag that can only take a set of predefined values:

```go
cmd := &cobra.Command{
    Use: "myapp",
    . . .
}

state := flags.NewEnumSliceFlag("+one", "+two", "three")
cmd.Flags().Var(state, "state", "State of the flag")
_ = cmd.RegisterFlagCompletionFunc(state.CompletionFunc("state"))
```

The default values are prepended with a `+`.

If all values can be provided with an `all` value, you can use:

```go
cmd := &cobra.Command{
    Use: "myapp",
    . . .
}

state := flags.NewEnumSliceFlagWithAllAllowed("one", "two", "three")
cmd.Flags().Var(state, "state", "State of the flag")
_ = cmd.RegisterFlagCompletionFunc(state.CompletionFunc("state"))
```

Note that there is no need to add the `all` value to the list of allowed values.
