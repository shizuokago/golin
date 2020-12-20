package golin

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/cheggaaa/pb/v3"
	"golang.org/x/xerrors"
)

type CompressType int

const (
	CompressZip CompressType = iota
	CompressTarGz
	CompressNotSupported
)

func getCompressType(n string) CompressType {
	if strings.Index(n, ".zip") == len(n)-4 {
		return CompressZip
	} else if strings.Index(n, ".tar.gz") == len(n)-7 {
		return CompressTarGz
	}
	return CompressNotSupported
}

func DecompressURL(url string, dir string) error {

	resp, err := http.Get(url)
	if err != nil {
		return xerrors.Errorf("http Get error: %w", err)
	}
	defer resp.Body.Close()

	switch getCompressType(url) {
	case CompressZip:
		return decompressZip(resp.Body, dir)
	case CompressTarGz:
		return decompressTarGz(resp.Body, dir)
	default:
		return fmt.Errorf("Decompress NotSupported: %s", url)
	}

	return fmt.Errorf("Not Reachable.")
}

func decompressZip(r io.Reader, dir string) error {

	err := os.Mkdir(dir, 0777)
	if err != nil {
		return xerrors.Errorf("make directory error: %w", err)
	}

	body, err := ioutil.ReadAll(r)
	if err != nil {
		return xerrors.Errorf("ioutil.ReadAll() error: %w", err)
	}

	fmt.Println("Downloaded!")
	zr, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil {
		return xerrors.Errorf("zip.NewReader() error: %w", err)
	}

	fmt.Println("Decompress...")

	bar := pb.StartNew(len(zr.File))
	for _, f := range zr.File {

		fn := filepath.Clean(dir + f.Name[2:])

		info := f.FileInfo()
		if info.IsDir() {
			err = os.MkdirAll(fn, 0777)
			if err != nil {
				return xerrors.Errorf("make directory error: %w", err)
			}
		} else {
			err = createZipFile(f, fn)
			if err != nil {
				return xerrors.Errorf("createZipFile error: %w", err)
			}
		}
		bar.Increment()
	}

	bar.Finish()

	return nil
}

func createZipFile(zf *zip.File, n string) error {
	fo, err := os.Create(n)
	if err != nil {
		return xerrors.Errorf("file create: %w", err)
	}
	defer fo.Close()

	f, err := zf.Open()
	if err != nil {
		return xerrors.Errorf("zip file open: %w", err)
	}
	defer f.Close()

	_, err = io.Copy(fo, f)
	if err != nil {
		return xerrors.Errorf("file copy: %w", err)
	}

	return nil
}

func decompressTarGz(r io.Reader, dir string) error {

	gzr, err := gzip.NewReader(r)
	if err != nil {
		return xerrors.Errorf("gzip.NewReader() error: %w", err)
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	//dirを作成

	for {
		th, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}

		fmt.Println(th.Name)
	}

	return nil
}
