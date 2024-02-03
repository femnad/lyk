package notify

import (
	"github.com/esiqveland/notify"
	"github.com/godbus/dbus/v5"
)

const (
	name = "lyk"
	icon = "emblem-favorite-symbolic"
)

func Send(summary, body string) error {
	conn, err := dbus.SessionBusPrivate()
	if err != nil {
		return err
	}
	defer conn.Close()

	err = conn.Auth(nil)
	if err != nil {
		return err
	}

	err = conn.Hello()
	if err != nil {
		return err
	}

	n := notify.Notification{
		AppIcon:       icon,
		AppName:       name,
		ExpireTimeout: notify.ExpireTimeoutSetByNotificationServer,
		Summary:       summary,
		Body:          body,
	}
	_, err = notify.SendNotification(conn, n)
	return err
}
