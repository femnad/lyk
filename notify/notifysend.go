package notify

import (
	"fmt"
	"strings"

	marecmd "github.com/femnad/mare/cmd"
)

const executable = "notify-send"

func escapeQuotes(s string) string {
	return strings.ReplaceAll(s, "'", "\\'")
}

func Send(summary, body string) error {
	summary = escapeQuotes(summary)
	body = escapeQuotes(body)
	cmd := fmt.Sprintf("%s '%s' '%s'", executable, summary, body)
	in := marecmd.Input{Command: cmd}

	_, err := marecmd.RunFormatError(in)
	if err != nil {
		return fmt.Errorf("error sending notification body '%s', summary '%s': %v", summary, body, err)
	}

	return nil
}
