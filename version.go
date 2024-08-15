package golin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/shizuokago/golin/v2/config"
	"golang.org/x/xerrors"
)

type VersionList []*Version

func (list VersionList) Exists(v *Version) bool {
	for _, src := range list {
		if src.Equals(v) {
			return true
		}
	}
	return false
}

// Version is r.v.m version
type Version struct {
	v    int
	r    int
	m    int
	mean VersionMean

	src string
}

type VersionMean int

const (
	MeanError VersionMean = iota
	MeanBeta
	MeanRC
	MeanNone
)

// Parse version string
// src = "1.12.1" v,r,m
// mean = none,rc,beta (error)
func NewVersion(src string) *Version {

	v := &Version{
		mean: MeanNone,
		src:  src,
	}
	v.v, v.r, v.m = -1, -1, -1

	base := src
	//rc,betaを確認
	idx := strings.Index(src, "rc")
	if idx != -1 {
		v.mean = MeanRC
		base = src[:idx]
	} else {
		idx = strings.Index(src, "beta")
		if idx != -1 {
			v.mean = MeanBeta
			base = src[:idx]
		}
	}

	var err error
	s := strings.Split(base, ".")

	//バージョン部分を数値化
	v.v, err = strconv.Atoi(s[0])
	if err != nil {
		v.mean = MeanError
	}

	if len(s) > 1 {
		v.r, err = strconv.Atoi(s[1])
		if err != nil {
			v.mean = MeanError
		}
	}

	if len(s) > 2 {
		v.m, err = strconv.Atoi(s[2])
		if err != nil {
			v.mean = MeanError
		}
	}

	if len(s) >= 4 {
		v.mean = MeanError
	}

	return v
}

func (v Version) IsError() bool {
	return v.mean == MeanError
}

func (v Version) Stable() bool {
	return v.mean == MeanNone
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

	if src.m > target.m {
		return 1
	} else if src.m < target.m {
		return -1
	}

	if src.mean > target.mean {
		return 1
	} else if src.mean < target.mean {
		return -1
	}

	if src.src > target.src {
		return 1
	} else if src.src < target.src {
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

func (src Version) Equals(target *Version) bool {
	return src.Compare(target) == 0
}

func (v Version) Minor() string {
	return fmt.Sprintf("%d.%d", v.v, v.r)
}

func (v Version) Revision() string {

	m := v.Minor()
	r := strings.Replace(v.src, m, "", 1)

	if len(r) == 0 {
		return "-"
	}

	if r[0] == '.' {
		return r[1:]
	}
	return r
}

func (v Version) Source() string {
	return v.src
}

func (v Version) String() string {
	return v.src
}

// dl ページのjsonからバージョンリストを取得
type GoVersion struct {
	Version string `json:"version"`
	Stable  bool   `json:"stable"`
}

func createVersionList() ([]*Version, error) {

	//https://go.dev/dl/?mode=json&include=all -> 全部
	resp, err := http.Get(config.GoDevelopDownloadPage)
	defer resp.Body.Close()

	var vers []GoVersion
	err = json.NewDecoder(resp.Body).Decode(&vers)
	if err != nil {
		return nil, xerrors.Errorf("json.Decode() error: %w", err)
	}

	versionList := make([]string, 0, len(vers))
	for _, ver := range vers {
		if strings.Index(ver.Version, "go") == 0 {
			versionList = append(versionList, ver.Version[2:])
		}
	}

	//errs := make([]error, 0)
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

// 最新のバージョンを取得
func getLatestVersion() (*Version, error) {

	list, err := createVersionList()
	if err != nil {
		return nil, xerrors.Errorf("create version list error: %w", err)
	}

	sort.Slice(list, func(i, j int) bool {
		return list[j].Less(list[i])
	})

	for _, elm := range list {
		if elm.mean == MeanNone {
			return elm, nil
		}
	}
	return nil, fmt.Errorf("Major version not found.")
}

func existsVersionList() VersionList {

	parent := filepath.Dir(os.Getenv("GOROOT"))
	entries, err := os.ReadDir(parent)
	if err != nil {
		return nil
	}

	var exists []*Version
	for _, d := range entries {
		if !d.IsDir() {
			continue
		}
		v := NewVersion(d.Name())
		if !v.IsError() {
			exists = append(exists, v)
		}
	}
	return exists
}
