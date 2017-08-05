package mimime

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type option int
type optionRegistration (func(r *requestOptions, arg string) error)

const (
	fileSizeOption    option = iota
	sslOption         option = iota
	forceReloadOption option = iota
	grayScaleOption   option = iota
	qualityOption     option = iota
	resizeOption      option = iota
)

var registeredOptions map[string]optionRegistration

func init() {
	registeredOptions = make(map[string]optionRegistration)
	addOption(registerFileSizeOption, "s", "-size", "-filesize")
	addOption(genToggleOptionRegistration(sslOption), "p", "-ssl")
	addOption(genToggleOptionRegistration(forceReloadOption), "f", "-force")
	addOption(genToggleOptionRegistration(grayScaleOption), "g", "-gray")
	addOption(registerQualityOption, "q", "-quality")
	addOption(registerResizeOption, "r", "-resize")
}

func addOption(or optionRegistration, prefixes ...string, ) {
	for _, el := range prefixes {
		registeredOptions[el] = or
	}
}

func registerOptions(ro *requestOptions, arg string) error {
	for pre, or := range registeredOptions {
		if strings.HasPrefix(arg, pre) {
			return or(ro, strings.TrimPrefix(arg, pre))
		}
	}
	errMsg := fmt.Sprintf("Unknown option: %s", arg)
	return errors.New(errMsg)
}

func genToggleOptionRegistration(opt option) optionRegistration {
	return func(r *requestOptions, arg string) error {
		r.setOpts[opt] = true
		return nil
	}
}

func registerFileSizeOption(r *requestOptions, arg string) error {
	fs, err := parseFileSize(arg)
	if err != nil {
		return err
	}
	r.setOpts[fileSizeOption] = true
	r.fs = *fs
	return nil
}

func registerQualityOption(r *requestOptions, arg string) error {
	quality, err := strconv.ParseFloat(arg, 64)
	if err != nil {
		return err
	}
	r.setOpts[qualityOption] = true
	r.qual = quality
	return nil
}

func registerResizeOption(r *requestOptions, arg string) error {
	if strings.Contains(arg, "x") {
		sizes := strings.Split(arg, "x")
		if len(sizes) != 2 {
			return errors.New(fmt.Sprintf("Invalid amount of size parameters: %s", arg))
		}
		var width int64
		var height int64
		var err error

		parseFsm := newResizeFsm()

		if sizes[0] == "" {
			parseFsm.Advance(noLeftParseInput)
		} else {
			width, err = strconv.ParseInt(sizes[0], 10, 64)
			if err != nil {
				return err
			}
			parseFsm.Advance(leftParseInput)
		}

		if sizes[1] == "" {
			parseFsm.Advance(noRightParseInput)
		} else {
			height, err = strconv.ParseInt(sizes[1], 10, 64)
			if err != nil {
				return err
			}
			parseFsm.Advance(rightParseInput)
		}

		finalState, err := parseFsm.Finalize()
		if err != nil {
			return errors.New("Invalid size option given.")
		}

		flag, err := resizeFlagFromState(finalState)
		if err != nil {
			return err
		}
		r.setOpts[resizeOption] = true
		r.re = imgSize{rf: flag, width: width, height: height}
		return nil
	}
	perc, err := strconv.ParseFloat(arg, 64)
	if err != nil {
		return err
	}
	r.setOpts[resizeOption] = true
	r.re = imgSize{rf: rFPerc, perc: perc}
	return nil
}
