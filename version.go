package golin

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/shizuokago/golin/v2/config"
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
		//バージョン部分を数値化
		v.v, err = strconv.Atoi(slice[0])
		//エラーがない場合
		if len(slice) > 1 && err == nil {
			//リビジョンを設定
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

func (src Version) Compare(target *Version) int {

	if src.mean == MeanError {
		return -1
	} else if target.mean == MeanError {
		return 1
	}

	if src.v > target.v {
		return 1
	} else if src.v < target.v {
		return -1
	}

	if src.r > target.r {
		return 1
	} else if src.r < target.r {
		return -1
	}

	if src.mean > target.mean {
		return 1
	} else if src.mean < target.mean {
		return -1
	}

	if src.m > target.m {
		return 1
	} else if src.m < target.m {
		return -1
	}

	return 0
}

// Version less
func (src Version) Less(target *Version) bool {
	if src.Compare(target) < 0 {
		return true
	}
	return false
}

func (v Version) String() string {
	return v.src
}

// GitHubのバージョン解析用のタグ
const (
	firstTag  = "div.Box-row"
	secondTag = "a.js-navigation-open"
)

// GitHubページ(dl)からバージョンを確認して、
// 可能なバージョンのスライスを取得
func createVersionList() ([]*Version, error) {

	doc, err := goquery.NewDocument(config.GitHubDownloadPage)
	if err != nil {
		return nil, xerrors.Errorf("error github page: %w", err)
	}

	//main#js-repo-pjax-container
	//div.Box

	boxRow := doc.Find(firstTag)
	if boxRow.Length() == 0 {
		return nil, xerrors.Errorf("Box-row is empty")
	}

	versionList := make([]string, 0, 100)
	errs := make([]error, 0)

	boxRow.Each(func(_ int, s *goquery.Selection) {
		link := s.Find(secondTag)
		if link.Length() == 0 {
			errs = append(errs, fmt.Errorf("link(a.js-navigation-open) not found."))
			return
		}
		name := link.First().Text()
		if isVersion(name) {
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

//バージョンを表すか？
func isVersion(name string) bool {

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

//最新のバージョンを取得
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
