package slogloki

import (
	"fmt"
	"log/slog"

	"github.com/grafana/loki-client-go/loki"
)

func NewLokiLogger(lokiUrl string, logLevel slog.Level) *slog.Logger {
	config, err := loki.NewDefaultConfig(lokiUrl)
	if err != nil {
		panic(fmt.Sprintf("failed to create new loki config:%v", err))
	}

	client, err := loki.New(config)
	if err != nil {
		panic(fmt.Sprintf("failed to create new loki client from config:\n%v\n%v", config, err))
	}

	return slog.New(Option{Level: logLevel, Client: client}.NewLokiHandler())
}
