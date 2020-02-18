package remodel

import (
	"reflect"
	"testing"
)

func Equals(t *testing.T, a, b interface{}) {
	if !reflect.DeepEqual(a, b) {
		t.Fatalf("Not equals %v and %v", a, b)
	}
}

func NotEquals(t *testing.T, a, b interface{}) {
	if reflect.DeepEqual(a, b) {
		t.Fatalf("Equals %v and %v", a, b)
	}
}

func Len(t *testing.T, a interface{}, expect int) {
	v := reflect.ValueOf(a)
	actual := v.Len()
	if actual != expect {
		t.Fatalf("Not match length: %d, expect:%d", actual, expect)
	}
}

func True(t *testing.T, b bool) {
	if !b {
		t.Fatal("Invalid boolean: False")
	}
}

func False(t *testing.T, b bool) {
	if b {
		t.Fatal("Invalid boolean: True")
	}
}
