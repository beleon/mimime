package mimime

import (
	"errors"
	"fmt"
	"os/exec"
)

func minificationCommand(req *request) (*exec.Cmd, error) {
	fn := "convert"
	args := []string{}
	for key, value := range req.reqOpts.setOpts {
		if value {
			switch key {
			case fileSizeOption:
				args = append(args, "-define", "jpeg:extent="+req.reqOpts.fs.String())
			case grayScaleOption:
				args = append(args, "-colorspace", "Gray")
			case qualityOption:
				args = append(
					args,
					"-quality",
					fmt.Sprintf("%.6f%%", req.reqOpts.qual))
			case resizeOption:
				re := req.reqOpts.re
				switch re.rf {
				case rFLeft:
					args = append(args, "-resize", fmt.Sprintf("%dx", re.width))
				case rFRight:
					args = append(args, "-resize", fmt.Sprintf("x%d", re.height))
				case rFBoth:
					args = append(
						args,
						"-resize",
						fmt.Sprintf("%dx%d", re.width, re.height))
				case rFPerc:
					args = append(
						args,
						"-resize",
						fmt.Sprintf("%.6f%%", re.perc))
				default:
					return nil, errors.New("Invalid resize flag.")
				}
			}
		}
	}
	args = append(args, req.originalImagePath())
	args = append(args, "jpeg:-")
	return exec.Command(fn, args...), nil
}
