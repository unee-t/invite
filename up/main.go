package main

import (
	"net/http"
	"os"

	"github.com/apex/log"
	"github.com/unee-t/env"
	"github.com/unee-t/invite"
)

func main() {

	h, err := invite.New()
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
