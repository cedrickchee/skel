package main

import (
	"reflect"
	"testing"
)

// *****************************************************************************
// TEST HELPERS
// *****************************************************************************

func assertEqual(t *testing.T, a, b interface{}) {
	t.Helper()
	if !reflect.DeepEqual(a, b) {
		t.Fatalf("expecting values to be equal but got: '%v' and '%v'", a, b)
	}
}
