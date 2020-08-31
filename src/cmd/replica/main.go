// replica is a service which writes POSTs to local storage
package main

import (
	"fmt"
	"net/http"
	"io"
	"os"
	"log"
	"flag"
	"strings"
	"path"
	"path/filepath"
	"substation/pkg/appender"
	"substation/pkg/httputil"
	"substation/pkg/config"
)

type Config struct {
	Host string `yaml:"host" json:"host"`
	Port int `yaml:"port" json:"port"`
}

var cfg = Config {
	Host: "",
	Port: 8080,
}

func init() {
	flag.IntVar(&cfg.Port, "port", cfg.Port, "port to listen on")
	flag.StringVar(&cfg.Host, "host", cfg.Host, "host to listen on")
	config.Flag(&cfg)
	flag.Parse()
	log.Println(cfg)
}
	
func handlePost(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	dir := path.Join("./", req.URL.Path)
	if strings.Contains(dir, "..") {
		httputil.HandleError(w, req, 400, "Found illegal '..' path element.")
		return
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		httputil.HandleError(w, req, 500, "Could not create log dir: %v", err)
		return
	}
	if a, err := appender.Create(dir); err != nil {
		httputil.HandleError(w, req, 500, "Could not create log file: %v", err)
		return
	} else {
		defer a.Close()
		if n, err := io.Copy(a, req.Body); err != nil {
			httputil.HandleError(w, req, 500, "Could not append to log file: %v", err)
			return
		} else {
			log.Printf("Wrote %d bytes to %s.", n, dir)
			return
		}
	}
}

func handleGet(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	dir := path.Join("./", req.URL.Path)
	if strings.Contains(dir, "..") {
		httputil.HandleError(w, req, 400, "Found illegal '..' path element.")
		return
	}
	var walker func(name string, info os.FileInfo, err error) error
	if strings.HasSuffix(dir, "/index") {
		// handle .../index requests by returning all log filenames rooted at dir
		dir = strings.TrimSuffix(dir, "index")
		walker = func(name string, info os.FileInfo, err error) error {
			if info.IsDir() {
				return nil
			}
			if path.Ext(name) != ".log" {
				log.Printf("Skipping non-log file %s.", name)
				return nil
			}
			fmt.Fprintln(w, name)
			return nil
		}
	} else {
		walker = func(name string, info os.FileInfo, err error) error {
			if info.IsDir() {
				return nil
			}
			if path.Ext(name) != ".log" {
				log.Printf("Skipping non-log file %s.", name)
				return nil
			} else if f, err := os.Open(name); err != nil {
				return fmt.Errorf("Error reading file %s: %v", name, err)
			} else if n, err := io.Copy(w, f); err != nil {
				return err
			} else {
				log.Printf("Responded with %d bytes from %s.", n, name)
				return nil
			}
		}
	}
	if err := filepath.Walk(dir, walker); err != nil {
		httputil.HandleError(w, req, 400, "Error walking path at %s: %v", dir, err)
		return
	}
}
	
func handleRequest(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "POST":
		handlePost(w, req)
	case "GET":
		handleGet(w, req)
	default:
		httputil.HandleError(w, req, 400, "Unsupported request method %s.", req.Method)
	}
}

func handleStatus(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintln(w, "GOOD")
}

func main() {
	http.HandleFunc("/status", handleStatus)
	http.HandleFunc("/", handleRequest)
	panic(http.ListenAndServe(fmt.Sprintf("%s:%d", cfg.Host, cfg.Port), nil))
}
