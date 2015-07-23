package main

import (
	"os"
	"strconv"
	"strings"
	"time"
)

func envInt(k string, d int) int {
	x := os.Getenv(k)
	if x != "" {
		if x1, e := strconv.Atoi(x); e != nil {
			return x1
		}
	}
	return d
}

func envDuration(k string, d time.Duration) time.Duration {
	x := os.Getenv(k)
	if x != "" {
		if x1, e := strconv.Atoi(x); e != nil {
			return time.Duration(x1) * time.Second
		}
	}
	return d
}

func envString(k, d string) string {
	x := os.Getenv(k)
	if x != "" {
		return x
	}
	return d
}

// isEnv will return true if s an environment variable, that is, it
// starts with a '$' and exists in the environment.
func isEnv(s string) bool {
	if len(s) < 2 { // Need at least $<LETTER>
		return false
	}
	if s[0] != '$' {
		return false
	}
	varname := s[1:]
	for _, env := range os.Environ() {
		parts := strings.Split(env, "=")
		if len(parts) < 2 {
			continue
		}
		// Exists, but may be empty
		if parts[0] == varname {
			return true
		}
	}
	return false
}
