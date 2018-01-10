package index

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	pb_almanac "dinowernli.me/almanac/proto"
)

// Serialize returns a proto with the contents of the supplied index.
func Serialize(index *Index) (*pb_almanac.BleveIndex, error) {
	buffer := &bytes.Buffer{}
	zipWriter := zip.NewWriter(buffer)

	info, err := os.Stat(index.path)
	if err != nil {
		return nil, fmt.Errorf("unable to stat index root %s: %v", index.path, err)
	}

	if !info.IsDir() {
		return nil, fmt.Errorf("can't work with non-dir index root %s", index.path)
	}

	err = filepath.Walk(index.path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("got error while walking at path %s: %v", path, err)
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return fmt.Errorf("could not create zip header: %v", err)
		}

		// The 'path' string is of the form '/tmp/root/somedir/foo' for a root dir
		// /tmp/root, so what we want is the suffix '/somedire/foo'.
		header.Name = strings.TrimPrefix(path, index.path)
		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(writer, file)

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk index directory: %v", err)
	}

	zipWriter.Close()
	return &pb_almanac.BleveIndex{
		DirectoryZip: buffer.Bytes(),
	}, nil
}

// Deserialize returns an instance of indexService which has loaded the
// supplied index proto.
func Deserialize(proto *pb_almanac.BleveIndex) (*Index, error) {
	bytesReader := bytes.NewReader(proto.DirectoryZip)
	bytesNum := int64(len(proto.DirectoryZip))
	zipReader, err := zip.NewReader(bytesReader, bytesNum)
	if err != nil {
		return nil, fmt.Errorf("unable to create zip reader for bytes")
	}

	root, err := ioutil.TempDir("", "alm-index-")
	if err != nil {
		return nil, fmt.Errorf("failed to create root dir for index: %v", err)
	}

	for _, file := range zipReader.File {
		path := filepath.Join(root, file.Name)
		if file.FileInfo().IsDir() {
			os.MkdirAll(path, file.Mode())
			continue
		}

		fileReader, err := file.Open()
		if err != nil {
			return nil, fmt.Errorf("failed to open zipped file %s: %v", file.Name, err)
		}
		defer fileReader.Close()

		destFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return nil, fmt.Errorf("failed to create file %s: %v", destFile.Name(), err)
		}
		defer destFile.Close()

		if _, err := io.Copy(destFile, fileReader); err != nil {
			return nil, fmt.Errorf("failed to copy into file %s: %v", destFile.Name(), err)
		}
	}

	result, err := openIndex(root)
	if err != nil {
		return nil, fmt.Errorf("failed to create index service: %v", err)
	}
	return result, nil
}
