package fileserver

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/Bearnie-H/easy-tls/header"
	"github.com/Bearnie-H/easy-tls/server"
)

// ModifiedTimeFormat defines the time format used by the LastModified time
// value when converting from time.Time values to strings to be written to the
// network.
const ModifiedTimeFormat = time.Stamp

// fileDetails is the set of file properties returned as headers during GET
// and HEAD calls. This is only a slightly modified copy of os.FileInfo, to
// assert HTTP Header encoding as specific types and with a clear naming
// scheme.
type fileDetails struct {
	Filename     string
	Size         int64  `easytls:"File-Size"`
	Permissions  string `easytls:"File-Mode"`
	LastModified string `easytls:"Last-Modified-Time"`
	IsDirectory  bool   `easytls:"-"`
}

// Handlers will return the standard full set of HTTP handlers to fully
// implement a simple file-system backed HTTP(S) file server.
//
// The standard set of handlers are:
//	GET:	Read the file from disk.
//	HEAD:	Write out basic details about the file.
//	POST:	Write the contents of the request body to disk as a new file.
//	PUT:	Overwrite an existing file on disk with the contents of the request body.
//	PATCH:	Append the request body to the existing file on disk.
//	DELETE:	Delete the file from disk.
//
// The server will be based out of the given ServeBase folder.
func Handlers(URLBase, ServeBase string) []server.SimpleHandler {
	return []server.SimpleHandler{
		Get(URLBase, ServeBase),
		Head(URLBase, ServeBase),
		Post(URLBase, ServeBase),
		Put(URLBase, ServeBase),
		Patch(URLBase, ServeBase),
		Delete(URLBase, ServeBase),
	}
}

// Get will attempt to read out the requested file from disk.
func Get(URLBase, ServeBase string) server.SimpleHandler {
	return server.SimpleHandler{
		Path:    URLBase,
		Methods: []string{http.MethodGet},
		Handler: func(w http.ResponseWriter, r *http.Request) {
			Filename := path.Join(ServeBase, strings.TrimPrefix(r.URL.Path, URLBase))
			Details, err := describeFile(Filename)
			if err != nil {
				w.WriteHeader(http.StatusNotFound)
				log.Println(err)
				w.Write([]byte(err.Error()))
				return
			}
			RespHeader := w.Header()

			H, err := header.DefaultEncode(*Details)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				log.Println(err)
				w.Write([]byte(err.Error()))
				return
			}

			header.Merge(&RespHeader, &H)

			f, err := os.Open(Filename)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				log.Println(err)
				w.Write([]byte(err.Error()))
				return
			}
			defer f.Close()

			if Details.IsDirectory {
				stats, err := f.Readdir(-1)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					log.Println(err)
					w.Write([]byte(err.Error()))
					return
				}
				sort.Slice(stats, func(i, j int) bool {
					return stats[i].Name() < stats[j].Name()
				})
				for _, stat := range stats {
					name := ""
					if stat.IsDir() {
						name = stat.Name() + "/"
					} else {
						name = stat.Name()
					}
					name = fmt.Sprintf("<a href=\"%s%s\">%s</a><br/>\n", r.URL.Path, name, name)
					w.Write([]byte(name))
				}
			} else {
				if _, err := io.Copy(w, f); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					log.Println(err)
					w.Write([]byte(err.Error()))
				}
			}
		},
	}
}

// Head will...
func Head(URLBase, ServeBase string) server.SimpleHandler {
	return server.SimpleHandler{
		Path:    URLBase,
		Methods: []string{http.MethodHead},
		Handler: func(w http.ResponseWriter, r *http.Request) {
			Filename := path.Join(ServeBase, strings.TrimPrefix(r.URL.Path, URLBase))
			Details, err := describeFile(Filename)
			if err != nil {
				w.WriteHeader(http.StatusNotFound)
				log.Println(err)
				w.Write([]byte(err.Error()))
				return
			}
			RespHeader := w.Header()

			H, err := header.DefaultEncode(*Details)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				log.Println(err)
				w.Write([]byte(err.Error()))
				return
			}

			header.Merge(&RespHeader, &H)
			w.WriteHeader(http.StatusOK)
		},
	}
}

// Post will...
func Post(URLBase, ServeBase string) server.SimpleHandler {
	return server.SimpleHandler{
		Path:    URLBase,
		Methods: []string{http.MethodPost},
		Handler: func(w http.ResponseWriter, r *http.Request) {
			Filename := path.Join(ServeBase, strings.TrimPrefix(r.URL.Path, URLBase))

			if err := os.MkdirAll(path.Dir(Filename), 0755); err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}

			f, err := os.Create(Filename)
			if err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}
			defer f.Close()

			if _, err := io.Copy(f, r.Body); err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}

			w.WriteHeader(http.StatusCreated)
		},
	}
}

// Put will...
func Put(URLBase, ServeBase string) server.SimpleHandler {
	return server.SimpleHandler{
		Path:    URLBase,
		Methods: []string{http.MethodPut},
		Handler: func(w http.ResponseWriter, r *http.Request) {
			Filename := path.Join(ServeBase, strings.TrimPrefix(r.URL.Path, URLBase))

			if err := os.Truncate(Filename, 0); err != nil {
				log.Printf("Truncate error: %s", err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}

			f, err := os.Create(Filename)
			if err != nil {
				log.Printf("Open error: %s", err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}
			defer f.Close()

			if _, err := io.Copy(f, r.Body); err != nil {
				log.Printf("Write error: %s", err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}

			w.WriteHeader(http.StatusAccepted)
		},
	}
}

// Patch will...
func Patch(URLBase, ServeBase string) server.SimpleHandler {
	return server.SimpleHandler{
		Path:    URLBase,
		Methods: []string{http.MethodPatch},
		Handler: func(w http.ResponseWriter, r *http.Request) {
			Filename := path.Join(ServeBase, strings.TrimPrefix(r.URL.Path, URLBase))

			f, err := os.OpenFile(Filename, os.O_APPEND|os.O_WRONLY, 0755)
			if err != nil {
				log.Printf("Open error: %s", err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}
			defer f.Close()

			if _, err := io.Copy(f, r.Body); err != nil {
				log.Printf("Write error: %s", err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}

			w.WriteHeader(http.StatusAccepted)
		},
	}
}

// Delete will...
func Delete(URLBase, ServeBase string) server.SimpleHandler {
	return server.SimpleHandler{
		Path:    URLBase,
		Methods: []string{http.MethodDelete},
		Handler: func(w http.ResponseWriter, r *http.Request) {
			Filename := path.Join(ServeBase, strings.TrimPrefix(r.URL.Path, URLBase))
			if err := os.Remove(Filename); err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}

			w.WriteHeader(http.StatusNoContent)
		},
	}
}

// describeFile will attempt to describe the given filename, returning a
// struct to be encoded into the returned HTTP headers of the response.
func describeFile(Filename string) (*fileDetails, error) {

	stat, err := os.Stat(Filename)
	if err != nil {
		return nil, err
	}

	return &fileDetails{
		Filename:     stat.Name(),
		Size:         stat.Size(),
		LastModified: stat.ModTime().Format(ModifiedTimeFormat),
		Permissions:  stat.Mode().String(),
		IsDirectory:  stat.IsDir(),
	}, nil
}
