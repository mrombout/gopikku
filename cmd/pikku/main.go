package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	blackfriday "gopkg.in/russross/blackfriday.v2"
)

type dir string

func (d dir) open(fileName string) ([]byte, error) {
	content, err := ioutil.ReadFile(string(d) + fileName)
	if err != nil {
		return nil, err
	}

	return content, err
}

func (d dir) stat(fileName string) (os.FileInfo, error) {
	info, err := os.Stat(string(d) + fileName)
	if err != nil {
		return nil, err
	}

	return info, err
}

type fileSystem interface {
	open(fileName string) ([]byte, error)
	stat(fileName string) (os.FileInfo, error)
}

type pikkuHandler struct {
	root fileSystem
}

func main() {
	panic(http.ListenAndServe(":8080", pikkuServer(dir("./"))))
}

func pikkuServer(root fileSystem) http.Handler {
	return &pikkuHandler{root}
}

func (p *pikkuHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Println("resolving")
	path := r.URL.Path
	file, err := p.root.stat(path)
	if err != nil {
		msg, err := toHTTPError(err)
		http.Error(w, msg, err)
		return
	}

	fmt.Println("dirring")
	if file.Mode().IsDir() {
		path = path + "/README.md"
	}

	fmt.Println("opening")
	content, err := p.root.open(path)
	if err != nil {
		msg, err := toHTTPError(err)
		http.Error(w, msg, err)
		return
	}

	fmt.Println("serving")
	contentType := http.DetectContentType(content)
	if strings.HasPrefix(contentType, "text/plain") {
		output := blackfriday.Run(content)
		fmt.Println(string(output))
		w.Write(output)
	} else {
		w.Header().Set("Content-Type", contentType)
		w.Write(content)
	}
}

func toHTTPError(err error) (msg string, httpStatus int) {
	if os.IsNotExist(err) {
		return "404 page not found", http.StatusNotFound
	}

	if os.IsPermission(err) {
		return "403 Forbidden", http.StatusForbidden
	}

	// Default:
	return "500 Internal Server Error", http.StatusInternalServerError
}
