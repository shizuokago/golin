package config

//実行オプション
type Option func(*Config) error

//シンボリックリンクの名称
func SetLinkName(l string) Option {
	return func(conf *Config) error {
		conf.LinkName = l
		return nil
	}
}
