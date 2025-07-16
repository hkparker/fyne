package camera

import (
	"io"
	"log/slog"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/internal/driver/mobile"
)

func Open() (io.Reader, error) {
	var filename string
	slog.Debug("camera.Open")

	var r io.Reader
	done := make(chan struct{})
	defer close(done)
	var openErr error
	mobile.ShowCameraOpen(func(reader io.Reader, err error) {
		slog.Debug("camera.Open: in callback", "reader", reader, "error", err)
		if err != nil {
			fyne.LogError("camera open: fyne returns error:", err)
			openErr = err
			done <- struct{}{}
			return
		}

		r = reader
		done <- struct{}{}
		slog.Debug("camera.Open: in callback: done")
	}, filename)

	slog.Debug("camera.Open: <-done")
	<-done

	slog.Debug("camera.Open: done", "error", openErr)
	return r, openErr
}
