package tile_proxy

import (
	"fmt"
	"github.com/hauke96/sigolo"
	"strings"
)

var (
	nextTraceId = 0
)

func newLogger(prefix string) *logger {
	defer func() { nextTraceId++ }() // Just increase trace-ID counter after return statement
	return &logger{
		LogTraceId: nextTraceId,
		LogPrefix:  prefix,
	}
}

type logger struct {
	LogTraceId int
	LogPrefix  string
}

func (l *logger) Log(format string, args ...interface{}) {
	sigolo.Infob(1, "%s-%X | %s", l.LogPrefix, l.LogTraceId, fmt.Sprintf(format, args...))
}

func (l *logger) Error(format string, args ...interface{}) {
	sigolo.Errorb(1, "%s-%X | %s", l.LogPrefix, l.LogTraceId, fmt.Sprintf(format, args...))
}

func (l *logger) Errorb(framesBackward int, format string, args ...interface{}) {
	sigolo.Errorb(1+framesBackward, "%s-%X | %s", l.LogPrefix, l.LogTraceId, fmt.Sprintf(format, args...))
}

func (l *logger) Debug(format string, args ...interface{}) {
	sigolo.Debugb(1, "%s-%X | %s", l.LogPrefix, l.LogTraceId, fmt.Sprintf(format, args...))
}

func (l *logger) Stack(err error) {
	sigolo.Stackb(1, err)
}

func (l *logger) LogQuery(query string, args ...interface{}) {
	for i, a := range args {
		query = strings.Replace(query, fmt.Sprintf("$%d", i+1), fmt.Sprintf("%v", a), 1)
	}

	sigolo.Debugb(1, query)
}
