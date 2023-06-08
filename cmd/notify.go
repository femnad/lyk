package cmd

import (
	"fmt"
	"github.com/femnad/lyk/notify"
)

const Name = "lyk"

func sendErrorNotification(msg string) error {
	name := fmt.Sprintf("%s error", Name)
	return notify.Send(name, msg)
}
