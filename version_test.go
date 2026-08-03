package golin_test

import (
	"testing"

	"github.com/shizuokago/golin/v2"
)

func TestVersion(t *testing.T) {

	v18 := golin.NewVersion("1.8")
	v1110 := golin.NewVersion("1.11.0")
	v1116 := golin.NewVersion("1.11.6")
	v112 := golin.NewVersion("1.12.1")
	v2 := golin.NewVersion("2.0beta1")

	if !v18.Less(v1110) {
		t.Errorf("Version less error. 1.8 < 1.11.0")
	}
	if !v18.Less(v2) {
		t.Errorf("Version less error. 1.8 < 2.0")
	}

	if v1116.Less(v1110) {
		t.Errorf("Version less error. 1.11.6 > 1.11.0")
	}

	if !v1116.Less(v112) {
		t.Errorf("Version less error. 1.11.6 < 1.12")
	}

	if !v112.Less(v2) {
		t.Errorf("Version less error. 1.12 < 2.0")
	}

}

// beta,rcの連番の並び順
//
// 文字列比較では beta10 < beta9 になってしまう
func TestVersionPreNumber(t *testing.T) {

	data := []struct {
		small string
		large string
	}{
		{"1.22beta9", "1.22beta10"},
		{"1.22rc1", "1.22rc2"},
		{"1.22rc9", "1.22rc10"},
		{"1.22beta10", "1.22rc1"},
		{"1.22rc10", "1.22"},
		{"1.22beta2", "1.22.0"},
	}

	for _, d := range data {
		s := golin.NewVersion(d.small)
		l := golin.NewVersion(d.large)
		if !s.Less(l) {
			t.Errorf("Version less error. %s < %s", d.small, d.large)
		}
		if l.Less(s) {
			t.Errorf("Version less error. %s > %s", d.large, d.small)
		}
	}
}
