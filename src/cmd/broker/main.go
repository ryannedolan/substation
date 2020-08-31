// Brokers accept incoming byte streams and write to one or more replicas.
package main

import (
	"context"
	"fmt"
	"flag"
	"net/http"
	"net/url"
	"io"
	"strings"
	"log"
	"substation/pkg/broadcast"
	"substation/pkg/httputil"
	"substation/pkg/config"
)

type Config struct {
	Host string `yaml:"host" json:"host"`
	Port int `yaml:"port" json:"port"`
	Replicas []Endpoint `yaml:"replicas" json:"replicas"`
	WaitFor int `yaml:"waitFor" json:"waitFor"`
}

type Endpoint struct {
	Target string
}

var cfg = Config {
	Host: "",
	Port: 8080,
	WaitFor: -1,
}

func init() {
	flag.IntVar(&cfg.Port, "port", cfg.Port, "port to listen on")
	flag.StringVar(&cfg.Host, "host", cfg.Host, "host to listen on")
	config.Flag(&cfg)
	flag.Parse()
	log.Println(cfg)
	if len(cfg.Replicas) == 0 {
		log.Println("WARNING: Broker has zero replicas. Writing to broker will have no effect.")
	}
	for i, e := range cfg.Replicas {
		if e.Target == "" {
			log.Fatalf("Replica[%d] has no Target.", i)
		} else if _, err := url.Parse(e.Target); err != nil {
			// we don't use the URL object, just parse for sanity check
			log.Fatalf("Could not parse URL %s in Replica[%d]: %v", e.Target, i, err)
		}
	}
}

func handlePost(w http.ResponseWriter, req *http.Request) {
	r := req.Body
	n := len(cfg.Replicas)
	defer r.Close()
	dir := req.URL.Path
	if strings.Contains(dir, "..") {
		httputil.HandleError(w, req, 400, "Found illegal '..' path element.")
		return
	}
	if err := broadcast.Broadcast(req.Context(), replicate(dir, r, n), n, cfg.WaitFor); err != nil {
		httputil.HandleError(w, req, 500, "Error broadcasting to %d replicas: %v", n, err)
		return
	}
}

func handleRequest(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "POST":
		handlePost(w, req)
	default:
		httputil.HandleError(w, req, 400, "Unsupported request method %s.", req.Method)
	}
}

func handleStatus(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintln(w, "GOOD")
}

/* Hand-rolled multiwriter
func split(r io.Reader, n int) (rs []io.Reader) {
	var ws []io.WriteCloser
	if n == 0 {
		return
	}
	for i := 0; i < n; i++ {
		r2, w := io.Pipe()
		r = io.TeeReader(r, w)
		rs = append(rs, r2)	
		ws = append(ws, w)	
	}
	go func() {
		io.Copy(ioutil.Discard, r)
		for _, w := range ws {
			w.Close()
		}	
	} ()
	return
}
*/

func split(r io.Reader, n int) (rs []io.Reader) {
	var ws []io.Writer
	// create n pipes
	for i := 0; i < n; i++ {
		r2, w := io.Pipe() 
		rs = append(rs, r2)
		ws = append(ws, w)
	}
	mw := io.MultiWriter(ws...)
	go func() {
		io.Copy(mw, r)
		for _, w := range ws {
			w.(io.WriteCloser).Close()
		}
	}()
	return
}

func replicate(dir string, r io.Reader, n int) func(ctx context.Context, i int) error {
	rs := split(r, n)
	return func(ctx context.Context, i int) error {
		t := cfg.Replicas[i].Target + dir
		log.Printf("POSTing to replica at %s.", t)
		if req, err := http.NewRequestWithContext(ctx, "POST", t, rs[i]); err != nil {
			return fmt.Errorf("Could not create request to replica at %s: %v", t, err)
		} else if resp, err := http.DefaultClient.Do(req); err != nil {
			return fmt.Errorf("Could not POST to replica at %s: %v", t, err)
		} else if resp.StatusCode / 100 != 2 {
			return fmt.Errorf("Got error response from replica at %s: %v", t, resp.Status)
		}
		return nil
	}
}

func main() {
	http.HandleFunc("/status", handleStatus)
	http.HandleFunc("/", handleRequest)
	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%d", cfg.Host, cfg.Port), nil))
}
