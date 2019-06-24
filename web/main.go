package main

import (
	"context"
	"net/http"
	"os"

	"github.com/apex/log"
	jsonloghandler "github.com/apex/log/handlers/json"
	"github.com/apex/log/handlers/text"
	"github.com/unee-t/invite"
)

func init() {
	if os.Getenv("UP_STAGE") != "" {
		log.SetHandler(jsonloghandler.Default)
	} else {
		log.SetHandler(text.Default)
	}
}

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
	if err := http.ListenAndServe(addr, app); err != nil {
		log.WithError(err).Fatal("error listening")
	}
}
