package firebase

import (
	"context"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
)

type Firebase struct {
	app *firebase.App
}

func NewFirebase(app *firebase.App) Firebase {
	return Firebase{
		app: app,
	}
}

func (f Firebase) SendPush(fcmToken string, title string, body string) error {
	ctx := context.Background()

	client, err := f.app.Messaging(ctx)
	if err != nil {
		return err
	}

	message := &messaging.Message{
		Token: fcmToken,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
	}

	_, err = client.Send(ctx, message)
	if err != nil {
		return err
	}

	return nil
}
