package golin

import (
	"fmt"
	"log"

	"golang.org/x/xerrors"
)

// PrintGoVersionList is download list printing
//
// インストール可能なバージョンリストを元に並び替えを行い表示します
// 存在するバージョンには「*」を表示します
//
// GitHubのDlページを参照していたが、JSに依存するサイトだった為、
// go.dev/dl ページに移行した。
// 現在動作しているOS、arc のみを表示するように変更
func PrintGoVersionList(all bool) error {

	//https://go.dev/dl/?mode=json
	verList, err := createVersionList()
	if err != nil {
		return xerrors.Errorf("createVersionList() error: %w", err)
	}

	//dl site
	// 1.5.4 1.6.4 1.7.6
	dlVers := VersionList{
		NewVersion("1.5.4"), NewVersion("1.6.4"), NewVersion("1.7.6")}
	//dl site older version 1.8
	older := NewVersion("1.8rc9")

	vers := make(VersionList, 0, len(verList))
	for _, ver := range verList {
		//DLに存在するバージョンより　小さい場合
		if ver.Less(older) {
			//特定バージョンじゃない場合、処理しない
			if !dlVers.Exists(ver) {
				continue
			}
		}
		vers = append(vers, ver)
	}

	//ファイルシステムの検索
	existVersions := existsVersionList()

	if all {
		for _, ver := range vers {
			mark := ""
			//存在確認
			if existVersions.Exists(ver) {
				mark = " *  "
			}
			fmt.Printf("%4s%s\n", mark, ver.Source())
		}
	} else {
		p := NewVersionPrinter(vers)
		p.addHave(existVersions)
		fmt.Println(
			"The revision is in parentheses.\n" +
				"\"-\" is a version without a revision number. \n" +
				"  exp) 1.19 (-) = 1.19, 1.23 (0) = 1.23.0, 1.22 (rc2) = 1.22rc2\n" +
				"Versions that exist are indented.")
		fmt.Println("-----------------------------------------------------------")
		p.Print()
		fmt.Println("-----------------------------------------------------------")
	}

	return nil
}

type VersionPrinter struct {
	minorMap map[string]*MinorVersion
	keys     VersionList
}

func NewVersionPrinter(vers VersionList) *VersionPrinter {
	var p VersionPrinter
	p.minorMap = make(map[string]*MinorVersion)
	for _, v := range vers {
		p.add(v)
	}
	return &p
}

func (p *VersionPrinter) addHave(exists VersionList) {
	for _, v := range exists {
		key := createMinorKey(v)
		m, ok := p.minorMap[key]
		if ok {
			m.addHave(v)
		} else {
			log.Println("impossible")
		}
	}
}

func (p *VersionPrinter) add(v *Version) {
	//2つめの"."以降を削除する
	key := createMinorKey(v)
	m, ok := p.minorMap[key]
	if !ok {
		m = NewMinorVersion(key)
		p.minorMap[key] = m
		p.keys = append(p.keys, NewVersion(key))
	}
	m.add(v)
}

func (p *VersionPrinter) Print() {
	for _, ver := range p.keys {
		m := p.minorMap[ver.Source()]
		fmt.Printf("%v", m)
	}
}

func createMinorKey(v *Version) string {
	return v.Minor()
}

type MinorVersion struct {
	version     string
	versionList VersionList
	have        VersionList
}

func NewMinorVersion(ver string) *MinorVersion {
	var m MinorVersion
	m.version = ver
	return &m
}

func (m *MinorVersion) add(v *Version) {
	m.versionList = append(m.versionList, v)
}

func (m *MinorVersion) addHave(v *Version) {
	m.have = append(m.have, v)
}

func (m *MinorVersion) String() string {

	//表示用のリストを作成
	fmt.Printf("%-6s", m.version)
	var stables VersionList
	var unstables VersionList
	for _, v := range m.versionList {
		if v.Stable() {
			stables = append(stables, v)
		} else {
			unstables = append(unstables, v)
		}
	}

	fmt.Printf(" ( ")
	for idx, v := range stables {
		if idx != 0 {
			fmt.Printf(",")
		}
		fmt.Printf(v.Revision())
	}

	if len(stables) != 0 && len(unstables) != 0 {
		fmt.Printf(" | ")
	}

	for idx, v := range unstables {
		if idx != 0 {
			fmt.Printf(",")
		}
		fmt.Printf(v.Revision())
	}

	fmt.Printf(" )\n")

	for _, v := range m.have {
		fmt.Printf("  %s\n", v.Source())
	}
	return ""
}
