package fileserver

import (
	"errors"
	"fmt"
	"html"
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

// HandlerLogger is the reference to the default logger to use for the FileServer handlers
var HandlerLogger *log.Logger

// fileDetails is the set of file properties returned as headers during GET
// and HEAD calls. This is only a slightly modified copy of os.FileInfo, to
// assert HTTP Header encoding as specific types and with a clear naming
// scheme.
type fileDetails struct {
	Filename     string
	Size         int64  `easytls:"File-Size"`
	Permissions  string `easytls:"File-Mode"`
	LastModified string `easytls:"Last-Modified-Time"`
	IsDirectory  bool   `easytls:"Directory"`
}

// Handlers will return the standard full set of HTTP handlers to fully
// implement a simple file-system backed HTTP(S) file server.
//
// The standard set of handlers are:
//
//	GET:	Read the file from disk.
//	HEAD:	Read out basic details about the file.
//	POST:	Write the contents of the request body to disk as a new file.
//	PUT:	Overwrite an existing file on disk with the contents of the request body.
//	PATCH:	Append the request body to the existing file on disk.
//	DELETE:	Delete the file from disk.
//
// The server will be based out of the given ServeBase folder.
func Handlers(URLBase, ServeBase string, ShowHidden bool, Logger *log.Logger) ([]server.SimpleHandler, error) {
	HandlerLogger = Logger

	if !strings.HasSuffix(URLBase, "/") {
		URLBase += "/"
	}

	if !strings.HasSuffix(ServeBase, "/") {
		ServeBase += "/"
	}

	if strings.Count(ServeBase, ".") > 0 {
		return nil, errors.New("file-server error: ServeBase cannot contain [ . ] characters")
	}

	return []server.SimpleHandler{
		Get(URLBase, ServeBase, ShowHidden),
		Head(URLBase, ServeBase),
		Post(URLBase, ServeBase),
		Put(URLBase, ServeBase),
		Patch(URLBase, ServeBase),
		Delete(URLBase, ServeBase),
	}, nil
}

// ExitHandler is the generic function to simplify failing out of a HTTP Handler within a plugin
//
// The general use of this is:
//
//	if err := foo(); err != nil {
//		ExitHandler(w, http.StatusInternalServerError, "Failed to foo the bar for ID [ %s ] with index [ %d ]", err, ID, index)
//		return
//	}
//
// This will write the status code to the response, as well as the result of
// fmt.Sprintf(Message, args...) to the response, and to the logger
func ExitHandler(w http.ResponseWriter, StatusCode int, Message string, err error, args ...interface{}) {
	w.WriteHeader(StatusCode)
	w.Write([]byte(html.EscapeString(fmt.Sprintf(Message, args...))))
	if HandlerLogger != nil {
		if err == nil {
			HandlerLogger.Printf(Message, args...)
		} else {
			HandlerLogger.Print(fmt.Sprintf(Message, args...) + " - " + err.Error())
		}
	}
}

// Get will attempt to read out the requested file from disk.
func Get(URLBase, ServeBase string, ShowHidden bool) server.SimpleHandler {
	return server.SimpleHandler{
		Path:        URLBase,
		Methods:     []string{http.MethodGet},
		Description: "Attempt to read and serve the requested file from the filesystem.",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			RelFilename := strings.TrimPrefix(r.URL.Path, URLBase)

			Filename := path.Join(ServeBase, RelFilename)

			if !strings.HasPrefix(Filename, ServeBase) {
				ExitHandler(w, http.StatusBadRequest, "file-server error: Filename attempted to move out of Server filesystem", nil)
				return
			}

			Details, err := describeFile(Filename)
			if os.IsNotExist(err) {

				if strings.Count(Filename, "..") > 0 {
					ExitHandler(w, http.StatusBadRequest, "file-server error: Filename contains parent directory reference in path", nil)
					return
				}
				ExitHandler(w, http.StatusNotFound, "file-server error: File [ %s ] does not exist", err, RelFilename)
				return
			} else if err != nil {
				ExitHandler(w, http.StatusInternalServerError, "file-server error: Failed to format HTTP Headers of file details", err)
				return
			}
			RespHeader := w.Header()

			H, err := header.DefaultEncode(*Details)
			if err != nil {
				ExitHandler(w, http.StatusInternalServerError, "file-server error: Failed to encode HTTP Headers of file details", err)
				return
			}

			header.Merge(&RespHeader, &H)

			f, err := os.Open(Filename)
			if os.IsNotExist(err) {
				ExitHandler(w, http.StatusNotFound, "file-server error: File [ %s ] could not be found", err, Filename)
				return
			} else if err != nil {
				ExitHandler(w, http.StatusInternalServerError, "file-server error: Error occurred while opening file [ %s ]", err, Filename)
				return
			}
			defer f.Close()

			if Details.IsDirectory {

				stats, err := f.Readdir(-1)
				if err != nil {
					ExitHandler(w, http.StatusInternalServerError, "file-server error: Failed to read directory contents for folder [ %s ]", err, path.Base(Filename))
					return
				}

				sort.Slice(stats, func(i, j int) bool {
					return stats[i].Name() < stats[j].Name()
				})

				if len(stats) == 0 {
					w.Write([]byte("No contents in directory\n"))
					return
				}

				for _, stat := range stats {
					name := ""
					if stat.IsDir() {
						name = stat.Name() + "/"
					} else {
						name = stat.Name()
					}

					// Hide names similar to how most file browsers do if the first character is a period.
					if name[0] == '.' && !ShowHidden {
						continue
					}

					name = fmt.Sprintf("<a href=\"%s%s\">%s</a><br/>\n", html.EscapeString(r.URL.Path), html.EscapeString(name), html.EscapeString(name))
					w.Write([]byte(name))
				}
				HandlerLogger.Printf("Successfully served directory [ %s ]", Filename)
			} else {
				if _, err := io.Copy(w, f); err != nil {
					ExitHandler(w, http.StatusInternalServerError, "file-server error: Failed to write file [ %s ] to network", err, Filename)
					return
				}
				HandlerLogger.Printf("Successfully served file [ %s ]", Filename)
			}
		}),
	}
}

// Head will...
func Head(URLBase, ServeBase string) server.SimpleHandler {
	return server.SimpleHandler{
		Path:        URLBase,
		Methods:     []string{http.MethodHead},
		Description: "Read out the permissions, access time, and other meta-data about the requested file.",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			RelFilename := strings.TrimPrefix(r.URL.Path, URLBase)

			Filename := path.Join(ServeBase, RelFilename)

			if !strings.HasPrefix(Filename, ServeBase) {
				ExitHandler(w, http.StatusBadRequest, "file-server error: Filename attempted to move out of Server filesystem", nil)
				return
			}

			Details, err := describeFile(Filename)
			if os.IsNotExist(err) {

				if strings.Count(Filename, "..") > 0 {
					ExitHandler(w, http.StatusBadRequest, "file-server error: Filename contains parent directory reference in path", nil)
					return
				}
				ExitHandler(w, http.StatusNotFound, "file-server error: File [ %s ] does not exist", err, RelFilename)
				return
			} else if err != nil {
				ExitHandler(w, http.StatusInternalServerError, "file-server error: Failed to format HTTP Headers of file details", err)
				return
			}
			RespHeader := w.Header()

			H, err := header.DefaultEncode(*Details)
			if err != nil {
				ExitHandler(w, http.StatusInternalServerError, "file-server error: Failed to encode HTTP Headers of file details", err)
				return
			}

			header.Merge(&RespHeader, &H)
			HandlerLogger.Printf("Successfully served HTTP Headers for file [ %s ]", Filename)
			w.WriteHeader(http.StatusOK)
		}),
	}
}

// Post will...
func Post(URLBase, ServeBase string) server.SimpleHandler {
	return server.SimpleHandler{
		Path:        URLBase,
		Methods:     []string{http.MethodPost},
		Description: "Write the incoming request body to the filesystem based on the filename of the URL.",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			RelFilename := strings.TrimPrefix(r.URL.Path, URLBase)

			if strings.Count(RelFilename, "..") > 0 {
				ExitHandler(w, http.StatusBadRequest, "file-server error: Filename contains parent directory reference in path", nil)
				return
			}

			if RelFilename == "" {
				RelFilename = "/"
			}
			Filename := path.Join(ServeBase, RelFilename)

			if !strings.HasPrefix(Filename, ServeBase) {
				ExitHandler(w, http.StatusBadRequest, "file-server error: Filename attempted to move out of Server filesystem", nil)
				return
			}

			if err := os.MkdirAll(path.Dir(Filename), 0755); err != nil {
				ExitHandler(w, http.StatusInternalServerError, "file-server error: Failed to assert directory exists for file [ %s ]", err, RelFilename)
				return
			}

			f, err := os.Create(Filename)
			if err != nil {
				ExitHandler(w, http.StatusInternalServerError, "file-server error: Failed to create file [ %s ]", err, RelFilename)
				return
			}
			defer f.Close()

			if _, err := io.Copy(f, r.Body); err != nil {
				ExitHandler(w, http.StatusInternalServerError, "file-server error: Failed to write file [ %s ]", err, RelFilename)
				return
			}

			ExitHandler(w, http.StatusCreated, "Successfully created file [ %s ]", nil, RelFilename)
		}),
	}
}

// Put will...
func Put(URLBase, ServeBase string) server.SimpleHandler {
	return server.SimpleHandler{
		Path:        URLBase,
		Methods:     []string{http.MethodPut},
		Description: "Overwrite the existing file with the contents of this request, failing if the file does not exist.",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			RelFilename := strings.TrimPrefix(r.URL.Path, URLBase)

			if strings.Count(RelFilename, "..") > 0 {
				ExitHandler(w, http.StatusBadRequest, "file-server error: Filename contains parent directory reference in path", nil)
				return
			}

			if RelFilename == "" {
				RelFilename = "/"
			}
			Filename := path.Join(ServeBase, RelFilename)

			if !strings.HasPrefix(Filename, ServeBase) {
				ExitHandler(w, http.StatusBadRequest, "file-server error: Filename attempted to move out of Server filesystem", nil)
				return
			}

			f, err := os.Create(Filename)
			if os.IsNotExist(err) {
				ExitHandler(w, http.StatusNotFound, "file-server error: File [ %s ] does not exist", err, RelFilename)
				return
			} else if err != nil {
				ExitHandler(w, http.StatusInternalServerError, "file-server error: Failed to open file [ %s ]", err, RelFilename)
				return
			}
			defer f.Close()

			if _, err := io.Copy(f, r.Body); err != nil {
				ExitHandler(w, http.StatusInternalServerError, "file-server error: Failed to write file [ %s ]", err, RelFilename)
				return
			}

			ExitHandler(w, http.StatusAccepted, "Successfully updated contents of file [ %s ]", nil, RelFilename)
			w.WriteHeader(http.StatusAccepted)
		}),
	}
}

// Patch will...
func Patch(URLBase, ServeBase string) server.SimpleHandler {
	return server.SimpleHandler{
		Path:        URLBase,
		Methods:     []string{http.MethodPatch},
		Description: "Append the contents of the request to the end of the specified file, failing if the file does not exist.",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			RelFilename := strings.TrimPrefix(r.URL.Path, URLBase)

			if strings.Count(RelFilename, "..") > 0 {
				ExitHandler(w, http.StatusBadRequest, "file-server error: Filename contains parent directory reference in path", nil)
				return
			}

			if RelFilename == "" {
				RelFilename = "/"
			}
			Filename := path.Join(ServeBase, RelFilename)

			if !strings.HasPrefix(Filename, ServeBase) {
				ExitHandler(w, http.StatusBadRequest, "file-server error: Filename attempted to move out of Server filesystem", nil)
				return
			}

			f, err := os.OpenFile(Filename, os.O_APPEND|os.O_WRONLY, 0755)
			if os.IsNotExist(err) {
				ExitHandler(w, http.StatusNotFound, "file-server error: File [ %s ] does not exist", err, RelFilename)
				return
			} else if err != nil {
				ExitHandler(w, http.StatusInternalServerError, "file-server error: Failed to open file [ %s ]", err, RelFilename)
				return
			}
			defer f.Close()

			if _, err := io.Copy(f, r.Body); err != nil {
				ExitHandler(w, http.StatusInternalServerError, "file-server error: Failed to write file [ %s ]", err, RelFilename)
				return
			}

			ExitHandler(w, http.StatusAccepted, "Successfully appended to file [ %s ]", nil, RelFilename)
		}),
	}
}

// Delete will...
func Delete(URLBase, ServeBase string) server.SimpleHandler {
	return server.SimpleHandler{
		Path:        URLBase,
		Methods:     []string{http.MethodDelete},
		Description: "Delete the specified file from the filesystem.",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			RelFilename := strings.TrimPrefix(r.URL.Path, URLBase)

			if strings.Count(RelFilename, "..") > 0 {
				ExitHandler(w, http.StatusBadRequest, "file-server error: Filename contains parent directory reference in path", nil)
				return
			}

			if RelFilename == "" {
				RelFilename = "/"
			}
			Filename := path.Join(ServeBase, RelFilename)

			if !strings.HasPrefix(Filename, ServeBase) {
				ExitHandler(w, http.StatusBadRequest, "file-server error: Filename attempted to move out of Server filesystem", nil)
				return
			}

			if err := os.Remove(Filename); err != nil {
				ExitHandler(w, http.StatusInternalServerError, "file-server error: Failed to delete file [ %s ]", err, Filename)
				return
			}

			ExitHandler(w, http.StatusNoContent, "Successfully deleted file [ %s ]", nil, RelFilename)
		}),
	}
}

// describeFile will attempt to describe the given filename, returning a
// struct to be encoded into the returned HTTP headers of the response.
func describeFile(Filename string) (*fileDetails, error) {

	if strings.Count(Filename, "..") > 0 {
		return nil, errors.New("file-server error: Filename contains parent directory reference")
	}

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
