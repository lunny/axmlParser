package axmlParser

import (
	"archive/zip"
	"io/ioutil"
)

func ParseApk(apkpath string, listener Listener) (*Parser, error) {
	r, err := zip.OpenReader(apkpath)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	var xmlf *zip.File

	for _, f := range r.File {
		if f.Name != "AndroidManifest.xml" {
			continue
		}
		xmlf = f
		break
	}

	if xmlf == nil {
		return nil, err
	}

	rc, err := xmlf.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	bs, err := ioutil.ReadAll(rc)
	if err != nil {
		return nil, err
	}

	parser := New(listener)
	err = parser.Parse(bs)
	if err != nil {
		return nil, err
	}
	return parser, nil
}

func ParseAxml(axmlpath string, listener Listener) (*Parser, error) {
	bs, err := ioutil.ReadFile(axmlpath)
	if err != nil {
		return nil, err
	}
	parser := New(listener)
	err = parser.Parse(bs)
	if err != nil {
		return nil, err
	}
	return parser, nil
}
