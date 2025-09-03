package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
)

func StartServer(ctx context.Context, handler http.Handler) (err error) {
	p, ok := os.LookupEnv("SERVER_PORT")
	if !ok {
		p = "8080"
	}

	slog.Info(fmt.Sprintf("starting server on port: %s", p))
	go func() {
		err = http.ListenAndServe(":"+p, handler)
		if errors.Is(err, http.ErrServerClosed) {
			slog.Info("server closed")
		} else if err != nil {
			err = fmt.Errorf("listen: %s", err)
		}
	}()

	<-ctx.Done()
	return
}
