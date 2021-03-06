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

var _ingesterHtmlTmpl = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x8c\x52\x4d\x6e\xdb\x3c\x10\xdd\xf3\x14\xf3\x71\x93\xcd\x27\x2b\x4d\xd1\x8d\x42\xa9\x8b\xa2\x05\xba\xcd\x0d\x68\x69\x24\x11\x21\x39\x02\x39\x4e\xa3\x0a\xbe\x7b\x41\x49\xb4\xdd\x20\x45\xbb\xb3\xe7\xfd\xcc\x9b\x27\xaa\xff\x3a\x6a\x79\x9e\x10\x46\x76\xb6\x11\xca\x1a\xff\x0c\x63\xc0\xbe\x96\x23\xf3\x14\xab\xb2\xec\xc9\x73\x3c\x0c\x44\x83\x45\x3d\x99\x78\x68\xc9\x95\x6d\x8c\x9f\x7b\xed\x8c\x9d\xeb\x27\x3a\x12\x93\x84\x80\xb6\x96\x91\x67\x8b\x71\x44\x64\xd9\x08\xa1\xd8\xb0\xc5\xe6\xbb\x1f\x30\x32\x06\x55\x6e\xff\x85\x50\x2b\xaf\x11\xe2\x48\xdd\x0c\x8b\x00\x48\x5b\x8a\xcd\xb1\x82\xbb\xcd\xf3\xee\x7f\x88\xda\xc7\x22\x62\x30\xfd\xa3\x38\x0b\x71\x18\x51\x77\x18\x56\xc5\xa4\xbb\xce\xf8\xa1\x38\x12\x33\xb9\x0a\x3e\xdc\x4f\xaf\x8f\x37\x73\xa6\xe9\x3a\x5c\xed\xa3\xf9\x89\x15\x3c\x7c\x4a\xa3\x64\xd6\x92\x67\x6d\xfc\x1b\x3f\x8b\x3d\x5f\x85\x4e\x87\xc1\xf8\x0a\x1e\xee\xb3\x0a\x3d\x87\xb9\x30\x7e\x3a\xf1\xaa\x73\xc6\x17\x23\x9a\x61\xe4\xc4\xca\x32\xe3\x8b\x1f\xa6\xe3\xb1\x82\x8f\x79\xd6\x99\x38\x59\x3d\x57\x70\xb4\xd4\x3e\x3f\xbe\x3d\xda\x91\xa7\x38\xe9\x16\xd7\x35\xaa\xcc\x15\xa9\xd4\x51\x23\x00\x54\x67\x5e\xa0\xb5\x3a\xc6\x5a\x5e\xa2\xcb\x84\xfc\x8e\x6d\x1d\xc9\xbd\x76\x55\x76\xe6\xa5\x11\x17\xd6\xce\xef\x29\x38\xd0\x2d\x1b\xf2\xb5\x2c\xcd\xfe\x85\x24\x38\xe4\x91\xba\x5a\x4e\x14\x79\xf7\x06\xf8\x42\x9e\xd1\x73\x05\x8a\xf1\x95\x75\x40\x0d\x5e\x3b\xac\x65\x2b\xf3\xd2\x9b\x56\x64\xb3\x2c\x87\x6f\x14\xdc\x2e\x3b\x9f\x55\x99\x75\xd9\x51\x6d\xfd\xa5\xa7\x57\xcb\x78\x3a\x3a\xc3\x12\x5e\xb4\x3d\x61\x2d\xb7\xdc\xf9\xb0\x32\x25\xcd\xbf\xaf\xa7\x2c\x0b\x98\x1e\x0e\x4f\x18\x4f\x96\xe1\x7c\xce\xbe\x37\x3d\x84\x15\xbb\x1c\xf1\x7e\x49\x9b\xc1\xee\x7c\x21\x4e\x01\xd3\x11\x1b\x98\xf2\xa7\x41\x5e\x71\xe5\x2e\x0b\xa0\xef\xd2\xf6\xdb\x4c\x5f\x43\xa0\xf0\x7e\x24\x4c\xd0\x5f\x12\xad\xf2\x3f\x04\x5a\xb1\x7f\xcc\xb3\xcf\x55\xb9\xbd\x9f\x5f\x01\x00\x00\xff\xff\x68\x22\xb4\x8a\xeb\x03\x00\x00")

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

	info := bindataFileInfo{name: "ingester.html.tmpl", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
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

	info := bindataFileInfo{name: "mixer.html.tmpl", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
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
	"mixer.html.tmpl":    mixerHtmlTmpl,
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
	"ingester.html.tmpl": {ingesterHtmlTmpl, map[string]*bintree{}},
	"mixer.html.tmpl":    {mixerHtmlTmpl, map[string]*bintree{}},
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
