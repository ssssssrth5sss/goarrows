package main

import (
	"strconv"
	"testing"
	"time"
)

// optionalInt64Flag is a flag.Value for -seed: unset means "not provided on CLI".
type optionalInt64Flag struct {
	set   bool
	value int64
}

// String implements flag.Value: empty when -seed was not passed, else the decimal seed.
func (o *optionalInt64Flag) String() string {
	if !o.set {
		return ""
	}
	return strconv.FormatInt(o.value, 10)
}

// Set implements flag.Value, parsing a base-10 int64 and marking the flag as present.
func (o *optionalInt64Flag) Set(s string) error {
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return err
	}
	o.value = v
	o.set = true
	return nil
}

// resolveProceduralSeed returns the explicit -seed value, or 0 under tests, or a time-based seed.
func resolveProceduralSeed(f *optionalInt64Flag) int64 {
	if f.set {
		return f.value
	}
	if testing.Testing() {
		return 0
	}
	return time.Now().UnixNano()
}
