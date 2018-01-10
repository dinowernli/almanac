// Code generated by go-bindata.
// sources:
// ingester.html.tmpl
// mixer.html.tmpl
// DO NOT EDIT!

package templates

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func bindataRead(data []byte, name string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, gz)
	clErr := gz.Close()

	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}
	if clErr != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type asset struct {
	bytes []byte
	info  os.FileInfo
}

type bindataFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
}

func (fi bindataFileInfo) Name() string {
	return fi.name
}
func (fi bindataFileInfo) Size() int64 {
	return fi.size
}
func (fi bindataFileInfo) Mode() os.FileMode {
	return fi.mode
}
func (fi bindataFileInfo) ModTime() time.Time {
	return fi.modTime
}
func (fi bindataFileInfo) IsDir() bool {
	return false
}
func (fi bindataFileInfo) Sys() interface{} {
	return nil
}

var _ingesterHtmlTmpl = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x8c\x92\xcf\x6e\xd4\x30\x10\x87\xef\x7e\x8a\xc1\x97\x5e\xd8\x75\xa9\xc4\x25\x75\xc2\x01\x81\xc4\xb5\x6f\xe0\x4d\x26\x89\x85\xed\x89\xec\xd9\x55\x97\x28\xef\x8e\x9c\x7f\xbb\x45\x45\x70\x4b\x66\xe6\xf7\xcd\x67\x27\xfa\x43\x43\x35\x5f\x07\x84\x9e\xbd\xab\x84\x76\x36\xfc\x84\x3e\x62\x5b\xca\x9e\x79\x48\x85\x52\x2d\x05\x4e\xc7\x8e\xa8\x73\x68\x06\x9b\x8e\x35\x79\x55\xa7\xf4\xa5\x35\xde\xba\x6b\xf9\x42\x27\x62\x92\x10\xd1\x95\x32\xf1\xd5\x61\xea\x11\x59\x56\x42\x68\xb6\xec\xb0\xfa\x11\x3a\x4c\x8c\x51\xab\xe5\x5d\x08\x3d\xcf\x55\x42\x9c\xa8\xb9\xc2\x28\x00\xf2\x96\xc3\x42\x2c\xe0\x61\x61\x3e\x7c\x84\x64\x42\x3a\x24\x8c\xb6\x7d\x16\x93\x10\xc7\x1e\x4d\x83\x71\x4e\x0c\xa6\x69\x6c\xe8\x0e\x27\x62\x26\x5f\xc0\xa7\xc7\xe1\xf5\xf9\xae\xce\x34\xdc\x8a\x33\x3e\xd9\x5f\x58\xc0\xd3\xe7\x5c\xca\xb0\x9a\x02\x1b\x1b\xfe\xe0\x39\x6c\xf9\x16\xf4\x26\x76\x36\x14\xf0\xf4\xb8\xa6\xb4\xda\xdc\x75\x96\xaf\x04\x80\x6e\xec\x05\x6a\x67\x52\x2a\xe5\xce\x94\xb9\xf3\xb6\xb7\xc8\xcb\xf5\x3e\xb4\x6a\xec\xa5\x12\xfb\xd4\x3a\xdf\x52\xf4\x60\x6a\xb6\x14\x4a\xa9\xec\x7a\x75\x12\x3c\x72\x4f\x4d\x29\x07\x4a\xbc\xb2\x01\xbe\x52\x60\x0c\x5c\x80\xb6\x61\x38\x33\xe4\x4f\x59\x4a\xc6\x57\x96\x10\x8c\xc7\x52\xd6\x12\x2e\xc6\x9d\xb1\x94\xe3\x78\xfc\x4e\xd1\xaf\x91\x69\xda\x21\x6f\xb2\xe9\x7c\xf2\x96\xf7\xd0\xa2\xba\x9d\x45\x65\xb9\xed\xf9\x66\x3f\x8e\x60\x5b\x38\xbe\x60\x3a\x3b\x86\x69\xda\xb8\x77\x47\x8f\x73\x6f\x5f\xf9\xfe\xbd\x2c\x80\x95\xbc\x0f\x0e\x11\xab\x71\x5c\xe9\xd3\xa4\x55\x2e\x6c\x2b\x6e\xb3\xe3\x08\x18\x9a\xbc\xfd\xde\xe9\x5b\x8c\x14\xdf\x57\xc2\xdc\xfa\x87\xd1\x1c\xff\x8b\xd0\xdc\xfb\x4f\x9f\xb5\xae\xd5\xf2\xcb\xfc\x0e\x00\x00\xff\xff\x3c\xf0\xc8\xb4\x77\x03\x00\x00")

func ingesterHtmlTmplBytes() ([]byte, error) {
	return bindataRead(
		_ingesterHtmlTmpl,
		"ingester.html.tmpl",
	)
}

func ingesterHtmlTmpl() (*asset, error) {
	bytes, err := ingesterHtmlTmplBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "ingester.html.tmpl", size: 887, mode: os.FileMode(420), modTime: time.Unix(1515609708, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _mixerHtmlTmpl = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x8c\x54\xcd\x6e\xf3\x36\x10\xbc\xeb\x29\xb6\xec\xe1\xbb\xd4\x56\x12\xb4\x40\xa1\x50\xea\x29\x3d\x14\xc8\xa1\x49\x5e\x80\x92\x56\x36\x11\x91\x54\xb8\xeb\xd4\xae\xe0\x77\xff\x40\x51\x3f\xb6\xe3\x04\x39\x52\x3b\x33\x3b\x1c\xee\x4a\xfe\x52\xbb\x8a\x0f\x1d\xc2\x96\x4d\x5b\x24\xb2\xd5\xf6\x15\xb6\x1e\x9b\x5c\x6c\x99\x3b\xca\xd2\xb4\x71\x96\x69\xbd\x71\x6e\xd3\xa2\xea\x34\xad\x2b\x67\xd2\x8a\xe8\xaf\x46\x19\xdd\x1e\xf2\x27\x57\x3a\x76\x02\x3c\xb6\xb9\x20\x3e\xb4\x48\x5b\x44\x16\x45\x92\x48\xd6\xdc\x62\xf1\xa8\xf7\xe8\x65\x1a\x0f\x89\x1c\x30\x45\x92\x94\xae\x3e\x40\x9f\x00\x84\x0e\xab\xa8\x96\xc1\x8f\xa8\xf7\xe3\x37\x20\x65\x69\x45\xe8\x75\x73\x9f\x1c\x93\x64\x5d\x39\xcb\x4a\x5b\xf4\x03\xa9\x53\x75\xad\xed\x66\xd5\x62\xc3\x19\xdc\xde\x74\xfb\xfb\x04\xc0\x28\xbf\xd1\x36\x83\xbb\xe1\x1c\x58\x5b\x54\xf5\x05\xa5\x74\xcc\xce\x2c\xa4\xe9\x3b\xbb\x6e\xf9\x38\x98\x22\xfd\x3f\x66\x70\xf7\xc7\x24\xf6\xb6\x43\x1f\x3d\x97\xce\xd7\xe8\x33\xb8\xed\xf6\x40\xae\xd5\x35\xfc\x5a\x55\xd5\x25\x71\xd4\xaa\x35\x75\xad\x3a\x64\x50\xb6\xae\x7a\x8d\x3e\xf7\xab\xff\x74\xcd\xdb\x0c\x7e\xbf\x99\xbd\x7a\xa4\x5d\xcb\x74\xd5\xec\xdd\x47\x5f\xb7\x7f\x9e\x13\x4f\x8c\x2d\x77\xbc\xf0\x17\xd0\xac\x0d\x12\x2b\xd3\x65\x99\x6a\x78\x0c\x27\xa4\x8b\x96\x33\x10\x99\x88\x30\x83\x44\x6a\x83\x1f\x9f\x48\x18\x67\x1d\x75\xaa\xc2\x11\x49\xa8\x7c\xb5\xfd\xb6\xed\x29\x4e\x99\x8e\xa3\x20\xc3\x28\x14\x09\x80\xac\xf5\x3b\x54\xad\x22\xca\xc5\xfc\xdc\x22\x54\xce\x6b\xb1\xe1\x58\x38\x2f\xc5\xf7\x16\xc5\xf3\x00\x91\x69\xad\xdf\x8b\x64\xc2\x35\xce\x1b\x50\x15\x6b\x67\x73\x91\x9a\x30\x97\xb3\x08\xc0\x8b\x36\x08\x5e\xd9\x0d\x66\x20\xb5\xed\x76\x0c\x61\x37\x72\xc1\xb8\x67\x01\x56\x19\xcc\x05\x09\x78\x57\xed\x0e\xf3\xbe\x5f\xff\xed\xbc\x79\x66\xe5\xf9\x91\x8e\xc7\x02\x56\x9f\xb3\xf0\x82\xf5\x60\xeb\xc8\x91\xa5\x4f\x17\x07\xff\x86\xf9\xfa\xa2\xf9\xdb\x85\xcc\x80\x1f\x64\x4e\x19\xb4\x2b\x8d\xe6\x09\x2a\x9e\x2f\xc2\x4a\x43\x0a\x63\xa6\x27\xf1\xf4\x3d\xe8\x06\xd6\x0f\xde\x3b\x0f\xc7\xe3\x95\x68\x31\x94\x4e\xf2\xba\x16\xfb\x40\x1f\x65\x67\x5c\xe7\xb1\xe8\xfb\x28\x7d\x3c\xca\x34\x9c\x67\x37\x33\xb4\xef\x01\x6d\x1d\x5a\x9f\xfa\x79\x42\xea\x9c\x25\x9c\x2c\x9d\x36\x1d\xf7\xe5\xcb\x39\x78\x8a\x98\xf3\x41\xe8\xfb\xf8\xce\x8b\xfc\xfa\xc1\xb2\xd7\x48\xcb\xcd\xaf\xb5\x3a\xb9\x3c\x80\xa4\x4e\xd9\xa9\x3e\xef\x94\x28\xfa\x1e\xd6\x2f\xd3\xf1\x31\x28\xca\x34\x40\x3f\xe5\x8e\x8b\x16\x99\xc1\xc7\xe1\x1f\x72\xf6\x0a\xef\x2c\xd7\x25\xae\x6f\xe6\xf8\xb6\x43\xe2\x6b\x31\xd6\x58\xee\x36\x5f\x86\x38\xcc\xd9\x59\x77\x59\xb9\x1a\x27\xd8\xf0\x5b\x8c\xf6\x97\x36\x32\x0d\x90\xe2\x73\x77\xf3\x67\x99\xc6\xfd\xff\x19\x00\x00\xff\xff\xc4\xbf\x5e\x6b\x8e\x06\x00\x00")

func mixerHtmlTmplBytes() ([]byte, error) {
	return bindataRead(
		_mixerHtmlTmpl,
		"mixer.html.tmpl",
	)
}

func mixerHtmlTmpl() (*asset, error) {
	bytes, err := mixerHtmlTmplBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "mixer.html.tmpl", size: 1678, mode: os.FileMode(420), modTime: time.Unix(1515609708, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("Asset %s can't read by error: %v", name, err)
		}
		return a.bytes, nil
	}
	return nil, fmt.Errorf("Asset %s not found", name)
}

// MustAsset is like Asset but panics when Asset would return an error.
// It simplifies safe initialization of global variables.
func MustAsset(name string) []byte {
	a, err := Asset(name)
	if err != nil {
		panic("asset: Asset(" + name + "): " + err.Error())
	}

	return a
}

// AssetInfo loads and returns the asset info for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func AssetInfo(name string) (os.FileInfo, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("AssetInfo %s can't read by error: %v", name, err)
		}
		return a.info, nil
	}
	return nil, fmt.Errorf("AssetInfo %s not found", name)
}

// AssetNames returns the names of the assets.
func AssetNames() []string {
	names := make([]string, 0, len(_bindata))
	for name := range _bindata {
		names = append(names, name)
	}
	return names
}

// _bindata is a table, holding each asset generator, mapped to its name.
var _bindata = map[string]func() (*asset, error){
	"ingester.html.tmpl": ingesterHtmlTmpl,
	"mixer.html.tmpl": mixerHtmlTmpl,
}

// AssetDir returns the file names below a certain
// directory embedded in the file by go-bindata.
// For example if you run go-bindata on data/... and data contains the
// following hierarchy:
//     data/
//       foo.txt
//       img/
//         a.png
//         b.png
// then AssetDir("data") would return []string{"foo.txt", "img"}
// AssetDir("data/img") would return []string{"a.png", "b.png"}
// AssetDir("foo.txt") and AssetDir("notexist") would return an error
// AssetDir("") will return []string{"data"}.
func AssetDir(name string) ([]string, error) {
	node := _bintree
	if len(name) != 0 {
		cannonicalName := strings.Replace(name, "\\", "/", -1)
		pathList := strings.Split(cannonicalName, "/")
		for _, p := range pathList {
			node = node.Children[p]
			if node == nil {
				return nil, fmt.Errorf("Asset %s not found", name)
			}
		}
	}
	if node.Func != nil {
		return nil, fmt.Errorf("Asset %s not found", name)
	}
	rv := make([]string, 0, len(node.Children))
	for childName := range node.Children {
		rv = append(rv, childName)
	}
	return rv, nil
}

type bintree struct {
	Func     func() (*asset, error)
	Children map[string]*bintree
}
var _bintree = &bintree{nil, map[string]*bintree{
	"ingester.html.tmpl": &bintree{ingesterHtmlTmpl, map[string]*bintree{}},
	"mixer.html.tmpl": &bintree{mixerHtmlTmpl, map[string]*bintree{}},
}}

// RestoreAsset restores an asset under the given directory
func RestoreAsset(dir, name string) error {
	data, err := Asset(name)
	if err != nil {
		return err
	}
	info, err := AssetInfo(name)
	if err != nil {
		return err
	}
	err = os.MkdirAll(_filePath(dir, filepath.Dir(name)), os.FileMode(0755))
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(_filePath(dir, name), data, info.Mode())
	if err != nil {
		return err
	}
	err = os.Chtimes(_filePath(dir, name), info.ModTime(), info.ModTime())
	if err != nil {
		return err
	}
	return nil
}

// RestoreAssets restores an asset under the given directory recursively
func RestoreAssets(dir, name string) error {
	children, err := AssetDir(name)
	// File
	if err != nil {
		return RestoreAsset(dir, name)
	}
	// Dir
	for _, child := range children {
		err = RestoreAssets(dir, filepath.Join(name, child))
		if err != nil {
			return err
		}
	}
	return nil
}

func _filePath(dir, name string) string {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	return filepath.Join(append([]string{dir}, strings.Split(cannonicalName, "/")...)...)
}

