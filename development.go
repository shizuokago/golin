package golin

// CompileSDK is dev用の内部バージョン名
//
// golin devが指定する値で、実際のディレクトリ名にはならない
// readyPath()でTipVersionに読み替えられる
const CompileSDK = "compile_sdk"

// TipVersion is 開発版の配置先ディレクトリ名
//
// golin devのビルド結果はここに配置され、
// golin tipはこのディレクトリを参照する
const TipVersion = "tip"

//
// CompileGoSDK is Compile from the latest repository to Create GoSDK
//
// Create()にCompileSDKを渡すことで開発用のgotipの実行を行います
//
func CompileLatestSDK() error {
	return Create(CompileSDK)
}
