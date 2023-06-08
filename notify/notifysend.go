package notify

import (
	"fmt"

	marecmd "github.com/femnad/mare/cmd"
)

const executable = "notify-send"

func Send(summary, body string) error {
	cmd := fmt.Sprintf("%s '%s' '%s'", executable, summary, body)
	in := marecmd.Input{Command: cmd}

	_, err := marecmd.RunFormatError(in)
	if err != nil {
		return fmt.Errorf("error sending notification body '%s', summary '%s': %v", summary, body, err)
	}

	return nil
}
