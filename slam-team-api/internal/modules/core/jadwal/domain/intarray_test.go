package domain

import (
	"reflect"
	"testing"
)

func TestIntArrayScan(t *testing.T) {
	cases := []struct {
		name string
		src  any
		want IntArray
		err  bool
	}{
		{"null", nil, nil, false},
		{"empty", "{}", IntArray{}, false},
		{"single", "{3}", IntArray{3}, false},
		{"many", "{0,2,4}", IntArray{0, 2, 4}, false},
		{"bytes", []byte("{1, 6}"), IntArray{1, 6}, false},
		{"bad literal", "1,2", nil, true},
		{"bad element", "{1,x}", nil, true},
		{"bad type", 42, nil, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var got IntArray
			err := got.Scan(tc.src)
			if tc.err {
				if err == nil {
					t.Fatalf("expected error, got %v", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("got %#v, want %#v", got, tc.want)
			}
		})
	}
}

func TestIntArrayValue(t *testing.T) {
	if v, _ := IntArray(nil).Value(); v != nil {
		t.Fatalf("nil slice should be NULL, got %v", v)
	}
	if v, _ := (IntArray{}).Value(); v != "{}" {
		t.Fatalf("empty slice should be '{}', got %v", v)
	}
	if v, _ := (IntArray{1, 3, 5}).Value(); v != "{1,3,5}" {
		t.Fatalf("got %v", v)
	}
}

func TestIntArrayRoundTrip(t *testing.T) {
	in := IntArray{0, 1, 6}
	v, err := in.Value()
	if err != nil {
		t.Fatal(err)
	}
	var out IntArray
	if err := out.Scan(v); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(in, out) {
		t.Fatalf("round trip mismatch: %v vs %v", in, out)
	}
	if !out.Contains(6) || out.Contains(2) {
		t.Fatal("Contains is wrong")
	}
}
