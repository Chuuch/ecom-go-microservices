package probes

type Probes struct {
	ReadinessPath        string `mapstructure:"readinessPath"`
	LivenessPath         string `mapstructure:"livenessPath"`
	Port                 string `mapstructure:"port"`
	Pprof                string `mapstructure:"pprof"`
	PrometheusPath       string `mapstructure:"prometheusPath"`
	PrometheusPort       string `mapstructure:"prometheusPort"`
	CheckIntervalSeconds int    `mapstructure:"checkIntervalSeconds"`
}
