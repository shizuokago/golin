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
	mean VersionMean
	m    int
	src  string
}

type VersionMean int

const (
	Major VersionMean = iota
	RC
	Beta
	MeanError
)

func (m VersionMean) String() string {
	switch m {
	case Major:
		return "Major"
	case RC:
		return "rc"
	case Beta:
		return "beta"
	case MeanError:
		return "Version Mean Error"
	}
	return "error(Mean not found)"
}

// Parse version string
// src = "1.12.1" R,V,M
// mean = major,rc,beta
func NewVersion(src string) *Version {
	v := &Version{
		mean: Major,
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
		v.mean = MeanError
	}
	return v
}

// setRevision
func (v *Version) setRevision(r string) error {
	key := Major
	if strings.Index(r, "beta") != -1 {
		key = Beta
	} else if strings.Index(r, "rc") != -1 {
		key = RC
	}

	var err error
	if key == Major {
		v.r, err = strconv.Atoi(r)
	} else {
		v.mean = key
		slice := strings.Split(r, key.String())
		if len(slice) == 2 {
			v.r, err = strconv.Atoi(slice[0])
			if err == nil {
				v.m, err = strconv.Atoi(slice[1])
			}
		}
	}

	if err != nil {
		v.mean = MeanError
	}
	return err
}

// Version less
func (src Version) Less(target *Version) bool {

	if src.mean == MeanError {
		return true
	} else if target.mean == MeanError {
		return false
	}

	if src.v != target.v {
		return src.v < target.v
	}

	if src.r != target.r {
		return src.r < target.r
	}

	//定数化したので判定できる
	if src.mean != target.mean {
		return src.mean > target.mean
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

	if len(versionList) <= 0 {
		return nil, fmt.Errorf("version not found.")
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

func getLatestVersion() (*Version, error) {

	list, err := createVersionList()
	if err != nil {
		return nil, xerrors.Errorf("create version list error: %w", err)
	}

	sort.Slice(list, func(i, j int) bool {
		return list[j].Less(list[i])
	})

	for _, elm := range list {
		if elm.mean == Major {
			return elm, nil
		}
	}
	return nil, fmt.Errorf("Major version not found.")
}
