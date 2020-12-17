package golin_test

import (
	"testing"

	"github.com/shizuokago/golin"
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
