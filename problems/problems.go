package problems

import (
    "fmt"
    "regexp"
    "strings"
)

type MessageLevel int

const (
    NA      = -1
    INFO    = 0
    WARNING = 1
    ERROR   = 2
)

func (lvl MessageLevel) String() string {
    switch lvl {
    case INFO:
        return "info"
    case WARNING:
        return "warning"
    case ERROR:
        return "error"
    default:
        return "n/a"
    }
}

type Message struct {
    Line  int
    Col   int
    Level MessageLevel
    Text  string
}

func (msg *Message) String() string {
    return fmt.Sprintf("%d:%d: %v: %s", msg.Line, msg.Col, msg.Level, msg.Text)
}

type MessageFilter interface {
    Filter(msg *Message) bool
}

type Scanner interface {
    ScanFile(filter MessageFilter, path string) ([]*Message, error)
}

type RegexpFilter struct {
    reject *regexp.Regexp
}

func (filter RegexpFilter) Filter(msg *Message) bool {
    return !filter.reject.MatchString(msg.Text)
}

func NewRegexpFilter(patterns []string) (*RegexpFilter, error) {
    wrappedPatterns := make([]string, len(patterns))
    for i, pat := range patterns {
        wrappedPatterns[i] = fmt.Sprintf("(%s)", pat)
    }
    comboPattern := strings.Join(wrappedPatterns, "|")
    reject, err := regexp.Compile(comboPattern)
    if err != nil {
        return nil, err
    }
    filter := new(RegexpFilter)
    filter.reject = reject
    return filter, nil
}
