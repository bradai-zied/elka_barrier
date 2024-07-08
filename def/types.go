package def

// YAMLConfig represents the structure of the YAML configuration file
type YAMLConfig struct {
	AppPort              int           `yaml:"APP_PORT"`
	DefaultNotification  []string      `yaml:"DefaultNotification"`
	Barriers             []YAMLBarrier `yaml:"Barriers"`
	TimeOutHttpResp      int           `yaml:"TimeOutHttpResp"`
	TimeoutConnectSec    int           `yaml:"TimeoutConnectSec"`
	MaxRetries           int           `yaml:"MaxRetries"`
	TimeBetweenRetrySec  int           `yaml:"TimeBetweenRetrySec"`
	TimeRetryAfterMaxSec int           `yaml:"TimeRetryAfterMaxSec"`
}

// YAMLBarrier represents a barrier as defined in the YAML file
type YAMLBarrier struct {
	IP          string `yaml:"ip"`
	Name        string `yaml:"name"`
	ID          []int  `yaml:"id"`
	BarrierType string `yaml:"barrierType"`
	Port        int    `yaml:"port"`
}

// Barrier represents a barrier for API responses
type Barrier struct {
	IP          string `json:"ip"`
	Name        string `json:"name"`
	ID          int    `json:"id"`
	BarrierType string `json:"barrierType"`
	Port        int    `json:"port"`
}

// BarrierResponse represents the response for the get all barriers endpoint
type BarrierResponse struct {
	Barriers []Barrier `json:"barriers"`
}
