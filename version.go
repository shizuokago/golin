package golin

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/xerrors"
)

// Version is r.v.m version
type Version struct {
	v    int
	r    int
	mean string
	m    int
	src  string
}

// Parse version string
// src = "1.12.1" R,V,M
// mean = major,rc,beta
func NewVersion(src string) *Version {
	v := &Version{
		mean: "major",
		src:  src,
	}
	var err error
	slice := strings.Split(src, ".")
	if len(slice) > 0 {
		v.v, err = strconv.Atoi(slice[0])
		if len(slice) > 1 && err == nil {
			r := slice[1]
			err = v.setRevision(r)
			if len(slice) > 2 && err == nil {
				v.m, err = strconv.Atoi(slice[2])
			}
		}
	}

	if err != nil {
		v.mean = "error"
	}
	return v
}

// setRevision
func (v *Version) setRevision(r string) error {
	key := ""
	if strings.Index(r, "beta") != -1 {
		key = "beta"
	} else if strings.Index(r, "rc") != -1 {
		key = "rc"
	}

	var err error
	if key == "" {
		v.r, err = strconv.Atoi(r)
	} else {
		v.mean = key
		slice := strings.Split(r, key)
		if len(slice) == 2 {
			v.r, err = strconv.Atoi(slice[0])
			if err == nil {
				v.m, err = strconv.Atoi(slice[1])
			}
		}
	}

	if err != nil {
		v.mean = "error"
	}
	return err
}

// Version less
func (src Version) Less(target *Version) bool {

	if src.mean == "error" {
		return true
	} else if target.mean == "error" {
		return false
	}

	if src.v != target.v {
		return src.v < target.v
	}

	if src.r != target.r {
		return src.r < target.r
	}

	if src.mean != target.mean {

		if src.mean == "beta" {
			return true
		} else if target.mean == "beta" {
			return false
		} else if src.mean == "rc" {
			return true
		} else if target.mean == "rc" {
			return false
		}

	} else {
		return src.m < target.m
	}

	return false
}

// print source
func (v Version) String() string {
	return v.src
}

func createVersionList() ([]*Version, error) {

	doc, err := goquery.NewDocument("https://github.com/golang/dl")
	if err != nil {
		return nil, xerrors.Errorf("error github page: %w", err)
	}

	//main#js-repo-pjax-container
	//div.Box
	versionList := make([]string, 0, 100)

	boxRow := doc.Find("div.Box-row")
	if boxRow.Length() == 0 {
		return nil, xerrors.Errorf("Box-row is empty")
	}

	errs := make([]error, 0)

	boxRow.Each(func(_ int, s *goquery.Selection) {
		link := s.Find("a.js-navigation-open")
		if link.Length() == 0 {
			errs = append(errs, fmt.Errorf("link(a.js-navigation-open) not found."))
			return
		}
		name := link.First().Text()
		if isGo(name) {
			versionList = append(versionList, name[2:])
		}
	})

	if len(errs) > 0 {
		msg := "link list error:"
		for _, elm := range errs {
			msg += "\n    " + elm.Error()
		}
		return nil, fmt.Errorf(msg)
	}

	v := make([]*Version, 0, 100)
	for _, elm := range versionList {
		v = append(v, NewVersion(elm))
	}

	sort.Slice(v, func(i, j int) bool {
		return v[i].Less(v[j])
	})

	return v, nil
}

func isGo(name string) bool {

	if len(name) < 3 {
		return false
	}

	if strings.Index(name, "go") != 0 {
		return false
	}

	if _, err := strconv.Atoi(name[2:3]); err != nil {
		return false
	}

	return true
}
