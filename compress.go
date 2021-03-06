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
	"time"

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

func Compress(w io.Writer, files ...string) error {

	zw := zip.NewWriter(w)
	defer zw.Close()

	for _, name := range files {
		err := addFiles(zw, name)
		if err != nil {
			return xerrors.Errorf("addFile() error: %w", err)
		}
	}

	return nil
}

func addFiles(w *zip.Writer, name string) error {
	info, err := os.Stat(name)
	if err != nil {
		return xerrors.Errorf("os.Stat(): %w", err)
	}

	if info.IsDir() {
		infos, err := ioutil.ReadDir(name)
		if err != nil {
			return xerrors.Errorf("ioutil.ReadDir() error: %w", err)
		}
		for _, elm := range infos {
			err := addFiles(w, elm.Name())
			if err != nil {
				return xerrors.Errorf("addFile() error: %w", err)
			}
		}
	} else {
		err = addFile(w, name)
		if err != nil {
			return xerrors.Errorf("addFile() error: %w", err)
		}
	}
	return nil
}

func addFile(w *zip.Writer, name string) error {

	data, err := ioutil.ReadFile(name)
	if err != nil {
		return xerrors.Errorf("ioutil.ReadFile() error: %w", err)
	}

	info, err := os.Lstat(name)
	if err != nil {
		return xerrors.Errorf("os.Lstat() error: %w", err)
	}

	h, err := zip.FileInfoHeader(info)
	if err != nil {
		return xerrors.Errorf("zip.FileInfoHeader() error: %w", err)
	}

	h.Name = name
	h.Method = zip.Deflate

	writer, err := w.CreateHeader(h)
	if err != nil {
		return xerrors.Errorf("writer CreateHeader() error: %w", err)
	}

	_, err = writer.Write(data)
	if err != nil {
		return xerrors.Errorf("writer Write() error: %w", err)
	}

	return nil
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

		//goを変換
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

	err := os.Mkdir(dir, 0777)
	if err != nil {
		return xerrors.Errorf("make directory error: %w", err)
	}

	fmt.Println("Downloaded!")

	gzr, err := gzip.NewReader(r)
	if err != nil {
		return xerrors.Errorf("gzip.NewReader() error: %w", err)
	}
	defer gzr.Close()

	fmt.Println("Decompress...")

	tr := tar.NewReader(gzr)

	fmt.Println(time.Now())

	bar := pb.StartNew(10000)
	for {
		th, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}

		name := th.Name[2:]
		//goを変換
		fn := filepath.Clean(dir + name)

		if th.Typeflag == tar.TypeDir {
			err = os.MkdirAll(fn, 0777)
			if err != nil {
				return xerrors.Errorf("make directory error: %w", err)
			}
		} else {
			err = createTarFile(tr, fn)
			if err != nil {
				return xerrors.Errorf("createTarFile(): %w", err)
			}

			err = os.Chmod(fn, os.FileMode(th.Mode))
			if err != nil {
				return xerrors.Errorf("bin file os.Chmod(): %w", err)
			}
		}
		bar.Increment()
	}

	bar.Finish()
	fmt.Println(time.Now())

	return nil
}

func createTarFile(r io.Reader, f string) error {

	fo, err := os.Create(f)
	if err != nil {
		return xerrors.Errorf("file create: %w", err)
	}
	defer fo.Close()

	_, err = io.Copy(fo, r)
	if err != nil {
		return xerrors.Errorf("file copy: %w", err)
	}

	return nil
}
