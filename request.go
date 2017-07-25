package mimime

import (
	"crypto/md5"
	"fmt"
	"path/filepath"
)

type resizeFlag int
type fileUnit string

type request struct {
	imgUrl  string
	_imgId  string
	reqOpts requestOptions
}

type requestOptions struct {
	setOpts map[option]bool
	fs      fileSize
	qual    float64
	re      imgSize
}

type fileSize struct {
	value float64
	unit  fileUnit
}

type imgSize struct {
	rf     resizeFlag
	width  int64
	height int64
	perc   float64
}

const (
	rFLeft  resizeFlag = iota
	rFRight resizeFlag = iota
	rFBoth  resizeFlag = iota
	rFPerc  resizeFlag = iota
)

const (
	bFU  fileUnit = "B"
	kbFU fileUnit = "KB"
	mbFU fileUnit = "MB"
	gbFU fileUnit = "GB"
)

func (fs fileSize) String() string {
	return fmt.Sprintf("%f%s", fs.value, fs.unit)
}

func (r *request) originalImagePath() string {
	return filepath.Join(cacheOrigPath, r.imgId())
}

func (r *request) imgId() string {
	if r._imgId == "" {
		r._imgId = fmt.Sprintf("%x", md5.Sum([]byte(r.imgUrl)))
	}
	return r._imgId
}
