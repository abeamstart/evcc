package easee

import (
	"fmt"
	"strings"

	"github.com/philippseith/signalr"
	"github.com/thoas/go-funk"
)

// Logger is a simple logger interface
type Logger interface {
	Println(v ...interface{})
}

type logger struct {
	b   strings.Builder
	log Logger
}

func SignalrLogger(log Logger) signalr.StructuredLogger {
	return &logger{log: log}
}

var skipKeys = []string{"class", "ts"}

func (l *logger) Log(keyVals ...interface{}) error {
	var skip bool
	fmt.Println(keyVals...)
	for i, v := range keyVals {
		// fmt.Printf("---- %d,%v\n", i, v)
		if i%2 == 0 {
			if funk.Contains(skipKeys, v) {
				skip = true
				continue
			}

			if l.b.Len() > 0 {
				l.b.WriteRune(' ')
			}
			l.b.WriteString(fmt.Sprintf("%v", v))
			l.b.WriteRune('=')
		} else {
			if skip {
				skip = false
				continue
			}

			l.b.WriteString(fmt.Sprintf("%v", v))
		}
	}

	l.log.Println(l.b.String())
	l.b.Reset()

	return nil
}
