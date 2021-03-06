package log

import (
	"fmt"
	"io"
	"strings"

	"k8s.io/klog"
)

// Logger is a simple interface that is roughly equivalent to klog.
type Logger interface {
	Is(level int) bool
	V(level int) Logger
	Infof(format string, args ...interface{})
}

// ToFile creates a logger that will log any items at level or below to file, and defer
// any other output to klog (no matter what the level is.)
func ToFile(w io.Writer, level int) Logger {
	return file{w, level}
}

var (
	// None implements the Logger interface but does nothing with the log output
	None Logger = discard{}
	// Log implements the Logger interface for Glog
	Log Logger = klogger{}
)

// discard is a Logger that outputs nothing.
type discard struct{}

func (discard) Is(level int) bool                { return false }
func (discard) V(level int) Logger               { return None }
func (discard) Infof(_ string, _ ...interface{}) {}

// klogger outputs log messages to klog
type klogger struct{}

func (klogger) Is(level int) bool {
	return bool(klog.V(klog.Level(level)))
}

func (klogger) V(level int) Logger {
	return kverbose{klog.V(klog.Level(level))}
}

func (klogger) Infof(format string, args ...interface{}) {
	klog.InfoDepth(2, fmt.Sprintf(format, args...))
}

// kverbose handles klog.V(x) calls
type kverbose struct {
	klog.Verbose
}

func (kverbose) Is(level int) bool {
	return bool(klog.V(klog.Level(level)))
}

func (kverbose) V(level int) Logger {
	if klog.V(klog.Level(level)) {
		return Log
	}
	return None
}

func (g kverbose) Infof(format string, args ...interface{}) {
	if g.Verbose {
		klog.InfoDepth(2, fmt.Sprintf(format, args...))
	}
}

// file logs the provided messages at level or below to the writer, or delegates
// to klog.
type file struct {
	w     io.Writer
	level int
}

func (f file) Is(level int) bool {
	return level <= f.level || bool(klog.V(klog.Level(level)))
}

func (f file) V(level int) Logger {
	// only log things that klog allows
	if !klog.V(klog.Level(level)) {
		return None
	}
	// send anything above our level to klog
	if level > f.level {
		return Log
	}
	return f
}

func (f file) Infof(format string, args ...interface{}) {
	fmt.Fprintf(f.w, format, args...)
	if !strings.HasSuffix(format, "\n") {
		fmt.Fprintln(f.w)
	}
}
