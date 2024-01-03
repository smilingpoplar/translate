package middleware

import (
	"reflect"
	"testing"
)

func TestRegroupTexts(t *testing.T) {
	t.Parallel()

	list := []string{"hello", "wonderful", "world"}
	want := [][]string{{"hello", "wonderful"}, {"world"}}
	got, err := regroupTexts(list, 15)
	if err != nil {
		t.Fatal(err)
	}
	for i := range got {
		if len(got[i]) == 0 {
			t.Fatalf("got empty list %d", i)
		}
		for j := range got[i] {
			if got[i][j] != want[i][j] {
				t.Errorf("expect %s, got %s", want[i][j], got[i][j])
			}
		}
	}
}

func TestSplitLongTexts(t *testing.T) {
	t.Parallel()

	list := []string{"hello\nfunny\nworld", "wonderful"}
	want := []string{"", "wonderful", "hello\nfunny", "world"}
	wantMapping := map[int][]int{0: {2, 3}}
	got, info, err := splitLongTexts(list, 12)
	if err != nil {
		t.Fatal(err)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("expect %s, got %s", want[i], got[i])
		}
	}
	if !reflect.DeepEqual(wantMapping[0], info.Mapping[0]) {
		t.Errorf("expect %v, got %v", wantMapping[0], info.Mapping[0])
	}

	merged := mergeBack(want, info)
	for i := range merged {
		if merged[i] != list[i] {
			t.Errorf("expect %s, got %s", list[i], merged[i])
		}
	}
}
