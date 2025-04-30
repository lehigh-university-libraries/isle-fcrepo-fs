package handler

import (
	"net/http"
	"strings"

	"github.com/lehigh-university-libraries/isle-fcrepo-fs/fcrepo"
)

func Download(w http.ResponseWriter, r *http.Request) {
	uri := strings.TrimPrefix(r.URL.Path, "/")
	if uri == "" {
		http.NotFound(w, r)
		return
	}
	uri = "info:fedora/" + uri
	realPath := fcrepo.RealPath(uri)
	if realPath == "" {
		http.NotFound(w, r)
		return
	}

	http.ServeFile(w, r, realPath)
}
