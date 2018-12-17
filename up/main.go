package main

import (
	"context"
	"net/http"
	"os"

	"github.com/apex/log"
	"github.com/unee-t/env"
	"github.com/unee-t/invite"
)

func main() {

	ctx := context.Background()
	h, err := invite.New(ctx)
	if err != nil {
		log.WithError(err).Fatal("error setting configuration")
		return
	}

	defer h.DB.Close()

	addr := ":" + os.Getenv("PORT")
	app := h.BasicEngine()

	if err := http.ListenAndServe(addr, env.Protect(app, h.APIAccessToken)); err != nil {
		log.WithError(err).Fatal("error listening")
	}

}
