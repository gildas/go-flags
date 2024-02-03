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

func TestPullRequestSuite(t *testing.T) {
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
	return &cobra.Command{Use: "root", RunE: func(cmd *cobra.Command, args []string) error {
		value, err := cmd.Flags().GetString("state")
		if err != nil {
			suite.Logger.Errorf("Error getting flag: %s", err)
			return err
		}
		suite.Logger.Infof("State Value: %s", value)
		cmd.Print(value)
		return nil
	}}
}

func (suite *FlagSuite) NewCommandWithSlice() *cobra.Command {
	return &cobra.Command{Use: "root", RunE: func(cmd *cobra.Command, args []string) error {
		values, err := cmd.Flags().GetStringSlice("state")
		if err != nil {
			suite.Logger.Errorf("Error getting flag: %s", err)
			return err
		}
		suite.Logger.Infof("State Values (%d items): %s", len(values), values)
		cmd.Print(values)
		return nil
	}}
}

// *****************************************************************************

func (suite *FlagSuite) TestEnumFlag() {
	root := suite.NewCommand()
	root.Flags().Var(&state, "state", "State of the flag")
	state := flags.NewEnumFlag("+one", "two", "three")
	root.Flags().Var(state, "state", "State of the flag")
	_ = root.RegisterFlagCompletionFunc("state", state.CompletionFunc())

	output, err := suite.Execute(root, "__complete", "--state", "")
	suite.Require().NoError(err)
	suite.Assert().Equal("one\ntwo\nthree\n:4\nCompletion ended with directive: ShellCompDirectiveNoFileComp\n", output)

	output, err = suite.Execute(root, "--state", "one")
	suite.Require().NoError(err)
	suite.Assert().Equal("one", output)

	_, err = suite.Execute(root, "--state", "four")
	suite.Require().Error(err)
}

func (suite *FlagSuite) TestEnumFlagWithFunc() {
	root := suite.NewCommand()
	state := flags.NewEnumFlagWithFunc("one", func(context.Context, *cobra.Command, []string) []string { return []string{"one", "two", "three"} })
	root.Flags().Var(state, "state", "State of the flag")
	_ = root.RegisterFlagCompletionFunc("state", state.CompletionFunc())

	output, err := suite.Execute(root, "__complete", "--state", "")
	suite.Require().NoError(err)
	suite.Assert().Equal("one\ntwo\nthree\n:4\nCompletion ended with directive: ShellCompDirectiveNoFileComp\n", output)

	output, err = suite.Execute(root, "--state", "one")
	suite.Require().NoError(err)
	suite.Assert().Equal("one", output)
}

func (suite *FlagSuite) TestEnumSliceFlag() {
	state := flags.EnumSliceFlag{Allowed: []string{"one", "two", "three"}, AllAllowed: true, Default: []string{"one"}}
	root := suite.NewCommandWithSlice()
	root.Flags().Var(&state, "state", "State of the flag")
	_ = root.RegisterFlagCompletionFunc("state", state.CompletionFunc())

	values := state.Get()
	suite.Assert().Equal([]string{"one"}, values)

	output, err := suite.Execute(root, "__complete", "--state", "")
	suite.Require().NoError(err)
	suite.Assert().Equal("one\ntwo\nthree\n:4\nCompletion ended with directive: ShellCompDirectiveNoFileComp\n", output)

	output, err = suite.Execute(root, "--state", "one", "--state", "two", "--state", "one")
	suite.Require().NoError(err)
	suite.Assert().Equal("[one two]", output)
	values = state.Get()
	suite.Assert().Equal([]string{"one", "two"}, values)
	suite.Assert().True(state.Contains("one"))
	suite.Assert().True(state.Contains("two"))
	suite.Assert().False(state.Contains("three"))
	suite.Assert().False(state.Contains("four"))

	output, err = suite.Execute(root, "--state", "all")
	suite.Require().NoError(err)
	suite.Assert().Equal("[one two three]", output)
	values = state.Get()
	suite.Assert().Equal([]string{"one", "two", "three"}, values)
	suite.Assert().True(state.Contains("one"))
	suite.Assert().True(state.Contains("two"))
	suite.Assert().True(state.Contains("three"))
	suite.Assert().False(state.Contains("four"))

	_, err = suite.Execute(root, "--state", "four")
	suite.Require().Error(err)
}
