package golin

import (
	"io"
	"os"
)

//実行オプション
type Option struct {
	LinkName string    //リンク名
	StdIn    io.Reader //エラー時の出力場所
	StdErr   io.Writer //エラー時の出力場所
	StdOut   io.Writer //出力場所
}

var option *Option

//
// DefaultOptoon is golin option
//
//
func DefaultOption() *Option {
	return &Option{
		LinkName: defaultLinkName,
		StdIn:    os.Stdin,
		StdOut:   os.Stdout,
		StdErr:   os.Stderr,
	}
}

//
// SetOption is
//
//
//
func SetOption(op *Option) {
	option = op
}

func getOption() *Option {
	if option == nil {
		option = DefaultOption()
	}
	return option
}
