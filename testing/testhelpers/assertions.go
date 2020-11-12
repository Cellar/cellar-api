package testhelpers

import (
	"fmt"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
)

func AssertF(condition bool, msg string, v ...interface{}) func(t *testing.T) {
	return func(t *testing.T) {
		Assert(t, condition, msg, v)
	}
}

func Assert(tb testing.TB, condition bool, msg string, v ...interface{}) {
	if !condition {
		file, line := getCaller()
		fmt.Printf("\033[31m%s:%d: "+msg+"\033[39m\n\n", append([]interface{}{filepath.Base(file), line}, v...)...)
		tb.Fail()
	}
}

func OkF(err error) func(t *testing.T) {
	return func(t *testing.T) {
		Ok(t, err)
	}
}

func Ok(tb testing.TB, err error) {
	if err != nil {
		file, line := getCaller()
		fmt.Printf("\033[31m%s:%d: unexpected error: %s\033[39m\n\n", filepath.Base(file), line, err.Error())
		tb.FailNow()
	}
}

func EqualsF(exp, act interface{}) func(t *testing.T) {
	return func(t *testing.T) {
		Equals(t, exp, act)
	}
}

func Equals(tb testing.TB, exp, act interface{}) {
	if !reflect.DeepEqual(exp, act) {
		file, line := getCaller()
		fmt.Printf("\033[31m%s:%d:\n\n\texpected: %#v\n\n\tactual: %#v\033[39m\n\n", filepath.Base(file), line, exp, act)
		tb.FailNow()
	}
}

func NotEqualsF(exp, act interface{}) func(t *testing.T) {
	return func(t *testing.T) {
		NotEquals(t, exp, act)
	}
}

func NotEquals(tb testing.TB, exp, act interface{}) {
	if reflect.DeepEqual(exp, act) {
		file, line := getCaller()
		fmt.Printf("\033[31m%s:%d:\n\n\texpected not: %#v\033[39\n\n", filepath.Base(file), line, exp)
		tb.FailNow()
	}
}

func IsNilF(act interface{}) func(t *testing.T) {
	return func(t *testing.T) {
		IsNil(t, act)
	}
}

func IsNil(tb testing.TB, act interface{}) {
	if !reflect.ValueOf(act).IsNil() {
		file, line := getCaller()
		fmt.Printf("\033[31m%s:%d:\n\n\texpected: nil\n\n\tactual: %#v\033[39m\n\n", filepath.Base(file), line, act)
		tb.FailNow()
	}
}

func getCaller() (file string, line int) {
	_, thisFile, _, _ := runtime.Caller(0)
	i := 0
	for file == "" || file == thisFile {
		_, file, line, _ = runtime.Caller(i)
		i++
	}
	return
}
