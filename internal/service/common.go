package service

const defaultEndpointConnectionWait = 1800

type (
	dataSourceConfig struct {
		sourcePath             string
		endpointConnectionWait int
	}

	ConfigOption func(hp *dataSourceConfig)
)

func WithEndpointConnectionWait(wait int) ConfigOption {
	return func(cfg *dataSourceConfig) {
		if wait > 0 {
			cfg.endpointConnectionWait = wait
		}
	}
}
