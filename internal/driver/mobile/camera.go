package mobile

import (
	"bytes"
	"io"
	"log/slog"

	"fyne.io/fyne/v2"
)

type hasCameraOpen interface {
	ShowCameraOpen(callback func(path string, closer func()), filename string)
}

// ShowCameraOpen loads the native file save dialog and returns the chosen file path via the callback func.
func ShowCameraOpen(callback func(r io.Reader, err error), filename string) {
	drv, ok := fyne.CurrentApp().Driver().(*driver)
	if !ok {
		return
	}

	a, ok := drv.app.(hasCameraOpen)
	if !ok {
		return
	}

	a.ShowCameraOpen(func(path string, closer func()) {
		slog.Debug("app show camera open", "path", path)
		if path == "" {
			callback(nil, nil)
			return
		}

		buf := bytes.NewBufferString(path)

		slog.Debug("app show camera open: call callback")

		callback(buf, nil)
	}, filename)
}
