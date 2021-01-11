package test

import (
	"testing"

	"github.com/kr/pretty"
)

type OrderedTests struct {
	TestDataSet DataSet
	OrderedList OrderedTestList
}

type DataSet map[string]Data
type OrderedTestList []string

type Data struct {
	Expected interface{}
	Data     interface{}
	Mock     interface{}
}

var TestResultString = "\n%s test failed.\n\nReturned:\n%+v\n\nExpected:\n%+v\nDiff:\n%+v"

func CheckResult(outputA interface{}, outputB interface{}, errA interface{}, errB interface{}, testCaseString string, t *testing.T) {
	if diff := pretty.Diff(outputA, outputB); len(diff) != 0 {
		t.Errorf(TestResultString, testCaseString, outputA, outputB, diff)
		return
	}

	if diff := pretty.Diff(errA, errB); len(diff) != 0 {
		t.Errorf(TestResultString, testCaseString, errA, errB, diff)
		return
	}
}
