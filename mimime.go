package mimime

import (
	"io"
	"net/http"
)

const applicationName string = "mimime"

//todo: export handler instead. make handler work for non "/" prefixes
func RunServer() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}

func handler(w http.ResponseWriter, r *http.Request) {
	req, err := parseRequest(r.URL.Path)
	if err != nil {
		logErr(w, err)
		return
	}

	go logRequest(req)

	//todo: make nonblocking?
	err = retrieveOriginal(req)
	if err != nil {
		logErr(w, err)
		return
	}

	/*
	 * <- gather info
	 */

	cmd, err := minificationCommand(req)
	if err != nil {
		logErr(w, err)
		return
	}

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		logErr(w, err)
		return
	}

	err = cmd.Start()
	if err != nil {
		logErr(w, err)
		return
	}

	_, err = io.Copy(w, stdoutPipe)
	if err != nil {
		logErr(w, err)
		return
	}
	cmd.Wait()
}
