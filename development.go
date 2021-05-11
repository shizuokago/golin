package golin

const CompileSDK = "compile_sdk"

//
// CompileGoSDK is Compile from the latest repository to Create GoSDK
//
// Create()にCompileSDKを渡すことで開発用のgotipの実行を行います
//
func CompileLatestSDK() error {
	return Create(CompileSDK)
}
