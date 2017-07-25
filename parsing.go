package mimime

import (
	"strconv"
	"strings"
)

func parseUrl(unparsedUrl string) (parsedUrl string, ssl bool, err error) {
	parsedUrl = unparsedUrl
	if strings.HasPrefix(unparsedUrl, "http:/") {
		parsedUrl = unparsedUrl[6:]
	} else if strings.HasPrefix(unparsedUrl, "https:/") {
		parsedUrl = unparsedUrl[7:]
		ssl = true
	}
	return
}

func parseRequest(path string) (*request, error) {
	urlSplit := strings.SplitN(path, "/u", 2)
	var unparsedImgUrl string
	var options []string
	var sslFlag bool
	var err error
	if len(urlSplit) > 1 {
		unparsedImgUrl = urlSplit[1]
		options = strings.Split(urlSplit[0], "/")
	} else {
		unparsedImgUrl = urlSplit[0][1:]
		options = []string{""}
	}

	imgUrl, sslFlag, err := parseUrl(unparsedImgUrl)
	if err != nil {
		return nil, err
	}

	reqOpts, err := parseRequestOptions(options[1:])
	if err != nil {
		return nil, err
	}
	if sslFlag {
		reqOpts.setOpts[sslOption] = true
	}
	return &request{imgUrl, "", *reqOpts}, nil
}

func parseRequestOptions(unparsedOptions []string) (*requestOptions, error) {
	setOpts := make(map[option]bool)
	ro := &requestOptions{setOpts: setOpts}
	for _, el := range unparsedOptions {
		err := registerOptions(ro, el)
		if err != nil {
			return nil, err
		}
	}
	return ro, nil
}

func parseFileSize(unparsedFileSize string) (*fileSize, error) {
	splitIndex := len(unparsedFileSize)
	for index, character := range unparsedFileSize {
		if !strings.Contains("1234567890.", string(character)) {
			splitIndex = index
			break
		}
	}
	size, err := strconv.ParseFloat(unparsedFileSize[:splitIndex], 64)
	if err != nil {
		return nil, err
	}
	fu, err := parseFileUnit(unparsedFileSize[splitIndex:])
	if err != nil {
		return nil, err
	}
	return &fileSize{size, *fu}, nil
}

func parseFileUnit(unparsedFileUnit string) (*fileUnit, error) {
	var fu fileUnit
	switch strings.ToLower(unparsedFileUnit) {
	case "b":
		fu = bFU
	case "kb":
		fu = kbFU
	case "mb":
		fu = mbFU
	case "gb":
		fu = gbFU
	default:
		fu = kbFU
	}
	return &fu, nil
}
