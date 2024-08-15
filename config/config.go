package config

import (
	"golang.org/x/xerrors"
)

type Config struct {
	LinkName string //リンク名
}

const (
	DefaultLinkName    = "current"               //作成するリンク名
	GoGetLink          = "golang.org/dl"         //ダウンロード時のリンク先
	GolangDownloadPage = "https://golang.org/dl" //install時のダウンロード

	// deprecated:
	//GitHubDownloadPage = "https://github.com/golang/dl" //GitHub上のバージョンリスト

	GoDevelopDownloadPage = "https://go.dev/dl/?mode=json&include=all"
)

var gConf *Config = nil

func defaultConfig() *Config {
	conf := Config{}
	conf.LinkName = DefaultLinkName
	return &conf
}

func Get() *Config {
	if gConf == nil {
		gConf = defaultConfig()
	}
	return gConf
}

func Set(opts ...Option) error {
	gConf := Get()
	for _, opt := range opts {
		err := opt(gConf)
		if err != nil {
			return xerrors.Errorf("config set error: %w", err)
		}
	}
	return nil
}
