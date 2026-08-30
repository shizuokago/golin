//go:build windows
// +build windows

package golin

import (
	"os"
	"path/filepath"
	"syscall"
	"unsafe"

	"golang.org/x/xerrors"
)

// ジャンクション(ディレクトリのマウントポイント)作成用の定数
//
// syscallパッケージには定義されていない為、ここで宣言する
const (
	ioReparseTagMountPoint = 0xA0000003
	fsctlSetReparsePoint   = 0x000900A4
)

// REPARSE_DATA_BUFFERのMountPointReparseBuffer部分
//
// この構造体(16バイト)の直後にUTF-16のパスが二つ並ぶ
type mountPointReparseBuffer struct {
	ReparseTag           uint32
	ReparseDataLength    uint16
	Reserved             uint16
	SubstituteNameOffset uint16
	SubstituteNameLength uint16
	PrintNameOffset      uint16
	PrintNameLength      uint16
}

// createLink is create directory link
//
// Windowsではシンボリックリンクの作成に管理者権限が必要な為、
// 権限の要らないジャンクションでリンクを作成する
//
// ジャンクションはローカルボリュームのディレクトリしか対象にできない為、
// 作成できない場合はシンボリックリンクにフォールバックする
func createLink(target, link string) error {

	err := createJunction(target, link)
	if err == nil {
		return nil
	}

	serr := os.Symlink(target, link)
	if serr != nil {
		return xerrors.Errorf("ジャンクションを作成できませんでした。(%v) : %w", err, serr)
	}
	return nil
}

// createJunction is create NTFS junction
//
// 空のディレクトリを作成し、
// そこにマウントポイントのリパースポイントを設定する
func createJunction(target, link string) error {

	abs, err := filepath.Abs(target)
	if err != nil {
		return xerrors.Errorf("filepath.Abs() error: %w", err)
	}
	abs = filepath.Clean(abs)

	//ジャンクションは空のディレクトリに対して設定する
	err = os.Mkdir(link, 0777)
	if err != nil {
		return xerrors.Errorf("os.Mkdir() error: %w", err)
	}

	//設定に失敗した場合、作成したディレクトリが残らないようにする
	success := false
	defer func() {
		if !success {
			os.Remove(link)
		}
	}()

	buf, err := mountPointBuffer(abs)
	if err != nil {
		return err
	}

	p, err := syscall.UTF16PtrFromString(link)
	if err != nil {
		return xerrors.Errorf("syscall.UTF16PtrFromString() error: %w", err)
	}

	//リパースポイントを設定する為、
	//リンクを辿らず(OPEN_REPARSE_POINT)ディレクトリを開く(BACKUP_SEMANTICS)
	h, err := syscall.CreateFile(p, syscall.GENERIC_WRITE, 0, nil, syscall.OPEN_EXISTING,
		syscall.FILE_FLAG_OPEN_REPARSE_POINT|syscall.FILE_FLAG_BACKUP_SEMANTICS, 0)
	if err != nil {
		return xerrors.Errorf("syscall.CreateFile() error: %w", err)
	}
	defer syscall.CloseHandle(h)

	var ret uint32
	err = syscall.DeviceIoControl(h, fsctlSetReparsePoint,
		&buf[0], uint32(len(buf)), nil, 0, &ret, nil)
	if err != nil {
		return xerrors.Errorf("syscall.DeviceIoControl() error: %w", err)
	}

	success = true
	return nil
}

// mountPointBuffer is REPARSE_DATA_BUFFER
//
// 引数の絶対パスに対するマウントポイントのバッファを作成する
//
// SubstituteNameはNTのパス(\??\C:\...)、
// PrintNameは表示用のパスで、それぞれNULL終端で並べる
func mountPointBuffer(abs string) ([]byte, error) {

	sub, err := syscall.UTF16FromString(`\??\` + abs)
	if err != nil {
		return nil, xerrors.Errorf("syscall.UTF16FromString() error: %w", err)
	}
	name, err := syscall.UTF16FromString(abs)
	if err != nil {
		return nil, xerrors.Errorf("syscall.UTF16FromString() error: %w", err)
	}

	//UTF16FromString()はNULL終端を含む為、名前の長さからは除く
	subLen := (len(sub) - 1) * 2
	nameLen := (len(name) - 1) * 2

	hdr := mountPointReparseBuffer{
		ReparseTag: ioReparseTagMountPoint,
		//オフセット位置(8バイト)とNULL終端(2バイト)を含めた長さ
		ReparseDataLength:    uint16(8 + subLen + 2 + nameLen + 2),
		SubstituteNameOffset: 0,
		SubstituteNameLength: uint16(subLen),
		PrintNameOffset:      uint16(subLen + 2),
		PrintNameLength:      uint16(nameLen),
	}

	//ReparseTag/ReparseDataLength/Reservedの8バイトは長さに含まない
	buf := make([]byte, 8+int(hdr.ReparseDataLength))
	*(*mountPointReparseBuffer)(unsafe.Pointer(&buf[0])) = hdr

	path := buf[unsafe.Sizeof(hdr):]
	putUTF16(path, sub)
	putUTF16(path[len(sub)*2:], name)

	return buf, nil
}

// putUTF16 is write UTF-16LE
func putUTF16(dst []byte, src []uint16) {
	for i, c := range src {
		dst[i*2] = byte(c)
		dst[i*2+1] = byte(c >> 8)
	}
}
