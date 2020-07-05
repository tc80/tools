package sentry

import (
	"os"
	"time"

	"github.com/cdnjs/tools/util"

	"github.com/getsentry/sentry-go"
)

// Init Sentry client
func Init() {
	if sentryDsn, ok := os.LookupEnv("SENTRY_DSN"); ok {
		util.Check(sentry.Init(sentry.ClientOptions{
			Dsn:              sentryDsn,
			AttachStacktrace: true,
		}))
	}
}

// PanicHandler registers panic handler to record the error in Sentry
func PanicHandler() {
	err := recover()

	if err != nil {
		NotifyError(err)
		panic(err)
	}
}

// NotifyError notifies sentry of an error
func NotifyError(err interface{}) {
	sentry.CurrentHub().Recover(err)
	sentry.Flush(time.Second * 5)
}
