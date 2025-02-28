package slogloki

import (
	"fmt"
	"log/slog"

	"github.com/grafana/loki-client-go/loki"
	"github.com/grafana/loki-client-go/pkg/labelutil"
	"github.com/prometheus/common/model"
)

func NewLokiLogger(serviceName string, lokiUrl string, logLevel slog.Level) *slog.Logger {
	config, err := loki.NewDefaultConfig(lokiUrl)
	if err != nil {
		panic(fmt.Sprintf("failed to create new loki config:%v", err))
	}

	config.ExternalLabels = labelutil.LabelSet{
		model.LabelSet{
			"service_name": model.LabelValue(serviceName),
		},
	}

	client, err := loki.New(config)
	if err != nil {
		panic(fmt.Sprintf("failed to create new loki client from config:\n%v\n%v", config, err))
	}

	return slog.New(Option{Level: logLevel, Client: client}.NewLokiHandler())
}
