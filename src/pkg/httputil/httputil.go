package httputil

import (
	"log"
	"fmt"
	"strconv"
	"time"
	"net/http"
)

func HandleError(w http.ResponseWriter, req *http.Request, status int, f string, args ...interface{}) {
	w.WriteHeader(status)
	msg := fmt.Sprintf(f, args...)
	fmt.Fprintln(w, msg)
	log.Printf("%s %s -> %d: %s", req.Method, req.URL.String(), status, msg)
}

func IntParam(req *http.Request, k string, def int64) (int64, error) {
	v := req.FormValue(k)
	if v == "" {
		return def, nil
	} else if i, err := strconv.ParseInt(v, 10, 64); err != nil {
		return def, fmt.Errorf("Could not parse integer parameter '%s': %v", k, err)
	} else {
		return i, nil
	}
}

// TimeParam only supports nanoseconds from epoch, for now.
func TimeParam(req *http.Request, k string, def time.Time) (time.Time, error) {
	v := req.FormValue(k)
	if v == "" {
		return def, nil
	} else if i, err := strconv.ParseInt(v, 10, 64); err != nil {
		return def, fmt.Errorf("Could not parse time parameter '%s': %v", k, err)
	} else {
		return time.Unix(0, i), nil
	}
}

func DurationParam(req *http.Request, k string, def time.Duration) (time.Duration, error) {
	v := req.FormValue(k)
	if v == "" {
		return def, nil
	} else if d, err := time.ParseDuration(v); err != nil {
		return def, fmt.Errorf("Could not parse duration parameter '%s': %v", k, err)
	} else {
		return d, nil
	}
}
