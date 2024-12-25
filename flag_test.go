package flags_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/gildas/go-errors"
	"github.com/gildas/go-flags"
	"github.com/gildas/go-logger"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/suite"
)

type FlagSuite struct {
	suite.Suite
	Name   string
	Logger *logger.Logger
	Start  time.Time
}

func TestFlagSuite(t *testing.T) {
	suite.Run(t, new(FlagSuite))
}

// *****************************************************************************
// Suite Tools

func (suite *FlagSuite) SetupSuite() {
	_ = godotenv.Load()
	suite.Name = strings.TrimSuffix(reflect.TypeOf(suite).Elem().Name(), "Suite")
	suite.Logger = logger.Create("test",
		&logger.FileStream{
			Path:         fmt.Sprintf("./log/test-%s.log", strings.ToLower(suite.Name)),
			Unbuffered:   true,
			SourceInfo:   true,
			FilterLevels: logger.NewLevelSet(logger.TRACE),
		},
	).Child("test", "test")
	suite.Logger.Infof("Suite Start: %s %s", suite.Name, strings.Repeat("=", 80-14-len(suite.Name)))
}

func (suite *FlagSuite) TearDownSuite() {
	suite.Logger.Debugf("Tearing down")
	if suite.T().Failed() {
		suite.Logger.Warnf("At least one test failed, we are not cleaning")
		suite.T().Log("At least one test failed, we are not cleaning")
	} else {
		suite.Logger.Infof("All tests succeeded, we are cleaning")
	}
	suite.Logger.Infof("Suite End: %s %s", suite.Name, strings.Repeat("=", 80-12-len(suite.Name)))
}

func (suite *FlagSuite) BeforeTest(suiteName, testName string) {
	suite.Logger.Infof("Test Start: %s %s", testName, strings.Repeat("-", 80-13-len(testName)))
	suite.Start = time.Now()
}

func (suite *FlagSuite) AfterTest(suiteName, testName string) {
	duration := time.Since(suite.Start)
	if suite.T().Failed() {
		suite.Logger.Errorf("Test %s failed", testName)
	}
	suite.Logger.Record("duration", duration.String()).Infof("Test End: %s %s", testName, strings.Repeat("-", 80-11-len(testName)))
}

func (suite *FlagSuite) LoadTestData(filename string) []byte {
	data, err := os.ReadFile(fmt.Sprintf("../../testdata/%s", filename))
	if err != nil {
		suite.T().Fatal(err)
	}
	return data
}

func (suite *FlagSuite) UnmarshalData(filename string, v interface{}) error {
	data := suite.LoadTestData(filename)
	suite.Logger.Infof("Loaded %s: %s", filename, string(data))
	return json.Unmarshal(data, v)
}

func (suite *FlagSuite) Execute(cmd *cobra.Command, args ...string) (string, error) {
	output := new(strings.Builder)
	cmd.SetOut(output)
	cmd.SetErr(output)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return output.String(), err
}

func (suite *FlagSuite) NewCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "root", RunE: func(cmd *cobra.Command, args []string) error {
		value, err := cmd.Flags().GetString("state")
		if err != nil {
			suite.Logger.Errorf("Error getting flag: %s", err)
			return err
		}
		suite.Logger.Infof("State Value: %s", value)
		cmd.Print(value)
		return nil
	}}
	cmd.SetContext(suite.Logger.ToContext(context.Background()))
	return cmd
}

func (suite *FlagSuite) NewCommandWithSlice() *cobra.Command {
	cmd := &cobra.Command{Use: "root", RunE: func(cmd *cobra.Command, args []string) error {
		values, err := cmd.Flags().GetStringSlice("state")
		if err != nil {
			suite.Logger.Errorf("Error getting flag: %s", err)
			return err
		}
		suite.Logger.Infof("State Values (%d items): %s", len(values), values)
		cmd.Print(values)
		return nil
	}}
	cmd.SetContext(suite.Logger.ToContext(context.Background()))
	return cmd
}

// *****************************************************************************

func (suite *FlagSuite) TestEnumFlag() {
	root := suite.NewCommand()
	state := flags.NewEnumFlag("+one", "two", "three")
	root.Flags().Var(state, "state", "State of the flag")
	_ = root.RegisterFlagCompletionFunc(state.CompletionFunc("state"))

	suite.Assert().Equal("one", state.Value)

	output, err := suite.Execute(root, "__complete", "--state", "")
	suite.Require().NoError(err)
	suite.Assert().Equal("one\ntwo\nthree\n:0\nCompletion ended with directive: ShellCompDirectiveDefault\n", output)

	output, err = suite.Execute(root, "--state", "one")
	suite.Require().NoError(err)
	suite.Assert().Equal("one", output)

	// See enum_flag.go for the commented code
	// _, err = suite.Execute(root, "--state", "four")
	// suite.Require().Error(err)
}

func (suite *FlagSuite) TestEnumFlagWithFunc() {
	root := suite.NewCommand()
	state := flags.NewEnumFlagWithFunc("one", func(context.Context, *cobra.Command, []string, string) ([]string, error) {
		return []string{"one", "two", "three"}, nil
	})
	root.Flags().Var(state, "state", "State of the flag")
	_ = root.RegisterFlagCompletionFunc(state.CompletionFunc("state"))

	suite.Assert().Equal("one", state.Value)

	output, err := suite.Execute(root, "--state", "two")
	suite.Require().NoError(err)
	suite.Assert().Equal("two", output)

	output, err = suite.Execute(root, "__complete", "--state", "")
	suite.Require().NoError(err)
	suite.Assert().Equal("one\ntwo\nthree\n:0\nCompletion ended with directive: ShellCompDirectiveDefault\n", output)

	output, err = suite.Execute(root, "--state", "one")
	suite.Require().NoError(err)
	suite.Assert().Equal("one", output)

	// See enum_flag.go for the commented code
	// _, err = suite.Execute(root, "--state", "four")
	// suite.Require().Error(err)
}

func (suite *FlagSuite) TestEnumFlagWithFuncReturningError() {
	root := suite.NewCommand()
	state := flags.NewEnumFlagWithFunc("one", func(context.Context, *cobra.Command, []string, string) ([]string, error) {
		return []string{}, errors.NotImplemented
	})
	root.Flags().Var(state, "state", "State of the flag")
	_ = root.RegisterFlagCompletionFunc(state.CompletionFunc("state"))

	output, err := suite.Execute(root, "__complete", "--state", "")
	suite.Require().NoError(err)
	suite.Assert().Equal(":1\nCompletion ended with directive: ShellCompDirectiveError\n", output)
}

func (suite *FlagSuite) TestEnumSliceFlag() {
	root := suite.NewCommandWithSlice()
	state := flags.NewEnumSliceFlag("+one", "+two", "three")
	root.Flags().Var(state, "state", "State of the flag")
	_ = root.RegisterFlagCompletionFunc(state.CompletionFunc("state"))

	values := state.GetSlice()
	suite.Assert().Equal([]string{"one", "two"}, values)

	output, err := suite.Execute(root, "__complete", "--state", "")
	suite.Require().NoError(err)
	suite.Assert().Equal("one\ntwo\nthree\n:0\nCompletion ended with directive: ShellCompDirectiveDefault\n", output)

	output, err = suite.Execute(root, "__complete", "--state", "one", "--state", "")
	suite.Require().NoError(err)
	suite.Assert().Equal("two\nthree\n:0\nCompletion ended with directive: ShellCompDirectiveDefault\n", output)

	_, err = suite.Execute(root, "--state", "four")
	suite.Require().Error(err, "four should not be allowed")

	output, err = suite.Execute(root, "--state", "one", "--state", "two", "--state", "one")
	suite.Require().NoError(err)
	suite.Assert().Equal("[one two]", output)
	values = state.GetSlice()
	suite.Assert().Equal([]string{"one", "two"}, values)

	output, err = suite.Execute(root, "--state", "one,two")
	suite.Require().NoError(err)
	suite.Assert().Equal("[one two]", output)
	values = state.GetSlice()
	suite.Assert().Equal([]string{"one", "two"}, values)
}

func (suite *FlagSuite) TestEnumSliceFlagWithAllAllowed() {
	root := suite.NewCommandWithSlice()
	state := flags.NewEnumSliceFlagWithAllAllowed("+one", "two", "+three")
	root.Flags().Var(state, "state", "State of the flag")
	_ = root.RegisterFlagCompletionFunc(state.CompletionFunc("state"))

	values := state.GetSlice()
	suite.Assert().Equal([]string{"one", "three"}, values)

	output, err := suite.Execute(root, "--state", "one", "--state", "two", "--state", "one")
	suite.Require().NoError(err)
	suite.Assert().Equal("[one two]", output)

	values = state.GetSlice()
	suite.Assert().Equal([]string{"one", "two"}, values)

	_, err = suite.Execute(root, "--state", "four")
	suite.Require().Error(err, "four should not be allowed")

	output, err = suite.Execute(root, "--state", "all")
	suite.Require().NoError(err)
	suite.Assert().Equal("[one two three]", output)
	values = state.GetSlice()
	suite.Assert().Equal([]string{"all", "one", "two", "three"}, values)

	_, err = suite.Execute(root, "--state", "four")
	suite.Require().Error(err)
}

func (suite *FlagSuite) TestEnumSliceFlagWithAllAllowedCompletion() {
	root := suite.NewCommandWithSlice()
	state := flags.NewEnumSliceFlagWithAllAllowed("+one", "two", "+three")
	root.Flags().Var(state, "state", "State of the flag")
	_ = root.RegisterFlagCompletionFunc(state.CompletionFunc("state"))

	values := state.GetSlice()
	suite.Assert().Equal([]string{"one", "three"}, values)

	output, err := suite.Execute(root, "__complete", "--state", "")
	suite.Require().NoError(err)
	suite.Assert().Equal("one\ntwo\nthree\nall\n:0\nCompletion ended with directive: ShellCompDirectiveDefault\n", output)

	output, err = suite.Execute(root, "__complete", "--state", "one", "--state", "")
	suite.Require().NoError(err)
	suite.Assert().Equal("two\nthree\nall\n:0\nCompletion ended with directive: ShellCompDirectiveDefault\n", output)

	output, err = suite.Execute(root, "__complete", "--state", "all", "--state", "")
	suite.Require().NoError(err)
	suite.Assert().Equal(":0\nCompletion ended with directive: ShellCompDirectiveDefault\n", output)

	output, err = suite.Execute(root, "__complete", "--state", "one", "--state", "all", "--state", "")
	suite.Require().NoError(err)
	suite.Assert().Equal(":0\nCompletion ended with directive: ShellCompDirectiveDefault\n", output)
}

func (suite *FlagSuite) TestEnumSliceFlagWithFuncShouldHaveDefault() {
	root := suite.NewCommandWithSlice()
	state := flags.NewEnumSliceFlagWithFunc(func(context.Context, *cobra.Command, []string, string) ([]string, error) {
		return []string{"one", "two", "three"}, nil
	}, "one", "two")
	root.Flags().Var(state, "state", "State of the flag")
	_ = root.RegisterFlagCompletionFunc(state.CompletionFunc("state"))

	values := state.GetSlice()
	suite.Assert().Equal([]string{"one", "two"}, values)
}

func (suite *FlagSuite) TestEnumSliceFlagWithFuncShouldAcceptOneValue() {
	root := suite.NewCommandWithSlice()
	state := flags.NewEnumSliceFlagWithFunc(func(context.Context, *cobra.Command, []string, string) ([]string, error) {
		return []string{"one", "two", "three"}, nil
	}, "one", "two")
	root.Flags().Var(state, "state", "State of the flag")
	_ = root.RegisterFlagCompletionFunc(state.CompletionFunc("state"))

	output, err := suite.Execute(root, "--state", "one")
	suite.Require().NoError(err)
	suite.Assert().Equal("[one]", output)
}

func (suite *FlagSuite) TestEnumSliceFlagWithFuncShouldNotRepeatValues() {
	root := suite.NewCommandWithSlice()
	state := flags.NewEnumSliceFlagWithFunc(func(context.Context, *cobra.Command, []string, string) ([]string, error) {
		return []string{"one", "two", "three"}, nil
	}, "one", "two")
	root.Flags().Var(state, "state", "State of the flag")
	_ = root.RegisterFlagCompletionFunc(state.CompletionFunc("state"))

	output, err := suite.Execute(root, "--state", "one", "--state", "two", "--state", "one")
	suite.Require().NoError(err)
	suite.Assert().Equal("[one two]", output)
	values := state.GetSlice()
	suite.Assert().Equal([]string{"one", "two"}, values)
}

func (suite *FlagSuite) TestEnumSliceFlagWithFuncShouldAcceptCommaSeparatedValues() {
	root := suite.NewCommandWithSlice()
	state := flags.NewEnumSliceFlagWithFunc(func(context.Context, *cobra.Command, []string, string) ([]string, error) {
		return []string{"one", "two", "three"}, nil
	}, "one", "two")
	root.Flags().Var(state, "state", "State of the flag")
	_ = root.RegisterFlagCompletionFunc(state.CompletionFunc("state"))

	output, err := suite.Execute(root, "--state", "one", "--state", "one,two")
	suite.Require().NoError(err)
	suite.Assert().Equal("[one two]", output)
	values := state.GetSlice()
	suite.Assert().Equal([]string{"one", "two"}, values)
}

func (suite *FlagSuite) TestEnumSliceFlagWithFuncShouldNotAcceptNonAllowedValues() {
	suite.T().Skip("We cannot test this reliably yet (see bitbucket cli)")
	root := suite.NewCommandWithSlice()
	state := flags.NewEnumSliceFlagWithFunc(func(context.Context, *cobra.Command, []string, string) ([]string, error) {
		return []string{"one", "two", "three"}, nil
	}, "one", "two")
	root.Flags().Var(state, "state", "State of the flag")
	_ = root.RegisterFlagCompletionFunc(state.CompletionFunc("state"))

	_, err := suite.Execute(root, "--state", "four")
	suite.Require().Error(err, "four should not be allowed")
}

func (suite *FlagSuite) TestEnumSliceFlagWithFuncShouldComplete() {
	root := suite.NewCommandWithSlice()
	state := flags.NewEnumSliceFlagWithFunc(func(context.Context, *cobra.Command, []string, string) ([]string, error) {
		return []string{"one", "two", "three"}, nil
	}, "one", "two")
	root.Flags().Var(state, "state", "State of the flag")
	_ = root.RegisterFlagCompletionFunc(state.CompletionFunc("state"))

	output, err := suite.Execute(root, "__complete", "--state", "")
	suite.Require().NoError(err)
	suite.Assert().Equal("one\ntwo\nthree\n:0\nCompletion ended with directive: ShellCompDirectiveDefault\n", output)
}

func (suite *FlagSuite) TestEnumSliceFlagWithAllAllowedAndFuncShouldHaveDefault() {
	root := suite.NewCommandWithSlice()
	state := flags.NewEnumSliceFlagWithAllAllowedAndFunc(func(context.Context, *cobra.Command, []string, string) ([]string, error) {
		return []string{"one", "two", "three"}, nil
	}, "one", "three")
	root.Flags().Var(state, "state", "State of the flag")
	_ = root.RegisterFlagCompletionFunc(state.CompletionFunc("state"))

	values := state.GetSlice()
	suite.Assert().Equal([]string{"one", "three"}, values)
}

func (suite *FlagSuite) TestEnumSliceFlagWithAllAllowedAndFuncShouldNotRepeaseValues() {
	root := suite.NewCommandWithSlice()
	state := flags.NewEnumSliceFlagWithAllAllowedAndFunc(func(context.Context, *cobra.Command, []string, string) ([]string, error) {
		return []string{"one", "two", "three"}, nil
	}, "one", "three")
	root.Flags().Var(state, "state", "State of the flag")
	_ = root.RegisterFlagCompletionFunc(state.CompletionFunc("state"))

	output, err := suite.Execute(root, "--state", "one", "--state", "two", "--state", "one")
	suite.Require().NoError(err)
	suite.Assert().Equal("[one two]", output)
	values := state.GetSlice()
	suite.Assert().Equal([]string{"one", "two"}, values)
}

func (suite *FlagSuite) TestEnumSliceFlagWithAllAllowedAndFuncShouldAcceptAllAsValue() {
	root := suite.NewCommandWithSlice()
	state := flags.NewEnumSliceFlagWithAllAllowedAndFunc(func(context.Context, *cobra.Command, []string, string) ([]string, error) {
		return []string{"one", "two", "three"}, nil
	}, "one", "three")
	root.Flags().Var(state, "state", "State of the flag")
	_ = root.RegisterFlagCompletionFunc(state.CompletionFunc("state"))

	output, err := suite.Execute(root, "--state", "all")
	suite.Require().NoError(err)
	// suite.Assert().Equal("[all one two three]", output)
	suite.Assert().Equal("[all]", output)
	values := state.GetSlice()
	// suite.Assert().Equal([]string{"all", "one", "two", "three"}, values)
	suite.Assert().Equal([]string{"all"}, values)
}

func (suite *FlagSuite) TestEnumSliceFlagWithAllAllowedAndFuncNotAcceptNotAllowedValues() {
	suite.T().Skip("We cannot test this reliably yet (see bitbucket cli)")
	root := suite.NewCommandWithSlice()
	state := flags.NewEnumSliceFlagWithAllAllowedAndFunc(func(context.Context, *cobra.Command, []string, string) ([]string, error) {
		return []string{"one", "two", "three"}, nil
	}, "one", "three")
	root.Flags().Var(state, "state", "State of the flag")
	_ = root.RegisterFlagCompletionFunc(state.CompletionFunc("state"))

	_, err := suite.Execute(root, "--state", "four")
	suite.Require().Error(err, "four should not be allowed")
}

func (suite *FlagSuite) TestEnumSliceFlagWithFuncReturningError() {

	root := suite.NewCommandWithSlice()
	state := flags.NewEnumSliceFlagWithFunc(func(context.Context, *cobra.Command, []string, string) ([]string, error) {
		return []string{}, errors.NotImplemented
	}, "one", "two")
	root.Flags().Var(state, "state", "State of the flag")
	_ = root.RegisterFlagCompletionFunc(state.CompletionFunc("state"))

	output, err := suite.Execute(root, "__complete", "--state", "")
	suite.Require().NoError(err)
	suite.Assert().Equal(":1\nCompletion ended with directive: ShellCompDirectiveError\n", output)
}

func (suite *FlagSuite) TestEnumSliceFlagCanReplaceValues() {
	root := suite.NewCommandWithSlice()
	state := flags.NewEnumSliceFlag("+one", "+two", "three")
	root.Flags().Var(state, "state", "State of the flag")
	_ = root.RegisterFlagCompletionFunc(state.CompletionFunc("state"))

	values := state.GetSlice()
	suite.Assert().Equal([]string{"one", "two"}, values)

	err := state.Replace([]string{"one", "three"})
	suite.Require().NoError(err)
	values = state.GetSlice()
	suite.Assert().Equal([]string{"one", "three"}, values)
}
