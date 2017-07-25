package mimime

import (
	"fmt"
	"io"
)

func logErr(w io.Writer, err error) {
	fmt.Fprintf(w, "Fatal error: %s", err)
}

func logRequest(req *request) {
	fmt.Printf("Requesting image %s\n", req.imgUrl)

	sslString := "disabled"
	if req.reqOpts.setOpts[sslOption] {
		sslString = "enabled"
	}
	fmt.Printf("SSL option is %s\n", sslString)
	forceString := "disabled"
	if req.reqOpts.setOpts[forceReloadOption] {
		forceString = "enabled"
	}
	fmt.Printf("Force reload option is %s\n", forceString)

	if req.reqOpts.setOpts[fileSizeOption] {
		fmt.Printf("Option fileSize is set with value: %s\n", req.reqOpts.fs)
	} else {
		fmt.Printf("No fileSize set using default value: %s\n", req.reqOpts.fs)
	}
}
