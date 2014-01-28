package problems

import (
	"bufio"
	"container/list"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
    "strings"
)

// TODO: all of this file assumes UTF-8 output from subprocess

type Tidy struct {
	msgRE *regexp.Regexp
}

func NewTidy() (*Tidy, error) {
	msgRE, err := regexp.Compile(`^line (\d+) column (\d+) - (\w+): (.+)$`)
	if err != nil {
		return nil, err
	}
	tidy := new(Tidy)
	tidy.msgRE = msgRE

	// TODO: check for "HTML5" in output of `tidy -version`

	return tidy, nil
}

func (tidy *Tidy) parseMsg(line string) (*Message, error) {
	groups := tidy.msgRE.FindStringSubmatch(line)
	if groups == nil {
		return nil, fmt.Errorf("can't parse Tidy output line: %q", line)
	}

	msg := new(Message)
	var err error

	msg.Line, err = strconv.Atoi(groups[1])
	if err != nil {
		return nil, err
	}

	msg.Col, err = strconv.Atoi(groups[2])
	if err != nil {
		return nil, err
	}

    level := strings.ToLower(groups[3])
	switch level {
	case "info":
		msg.Level = INFO
	case "warning":
		msg.Level = WARNING
	case "error":
		msg.Level = ERROR
	default:
		return nil, fmt.Errorf("can't parse Tidy message level: %q", groups[3])
	}

	msg.Text = groups[4]

	return msg, nil
}

func (tidy *Tidy) ScanFile(filter MessageFilter, path string) ([]*Message, error) {
	cmd := exec.Command("tidy", "-quiet", "-errors", "-utf8", path)

	// get stderr pipe
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	// handle subprocess start errors
    err = cmd.Start()
    if err != nil {
        return nil, err
    }

    // try to parse every line as a message
    msgList := list.New()
    lines := bufio.NewScanner(stderr)
	for lines.Scan() {
		line := lines.Text()
		msg, msgErr := tidy.parseMsg(line)
		if msgErr != nil {
            return nil, msgErr
		}
        keep := filter.Filter(msg)
        if keep {
            msgList.PushBack(msg)
        }
	}

    var msgs []*Message
    if msgList.Len() > 0 {
        msgs = make([]*Message, msgList.Len())
    }
    i := 0
	for el := msgList.Front(); el != nil; el = el.Next() {
		msg := el.Value.(*Message)
		msgs[i] = msg
		i += 1
	}

    // handle subprocess exit errors
    err = cmd.Wait()
    if err != nil {
        _, isExitErr := err.(*exec.ExitError)
        // tidy returns with exit code 1 if there were errors
        if !isExitErr || len(msgs) == 0 {
            return nil, err
        }
    }

    // handle scanner errors
    err = lines.Err()
    if err != nil {
        return nil, err
    }

	return msgs, nil
}
