package godog

import (
	"fmt"
	"testing"

	"github.com/spf13/viper"
)

type Config struct {
	Environment      string  `mapstructure:"environment"`
	URL              string  `mapstructure:"url"`
	KeystoreFile     string  `mapstructure:"keystore_file"`
	KeystorePassword string  `mapstructure:"keystore_password"`
	ChainId          string  `mapstructure:"chain_id"`
	Bech32Prefix     string  `mapstructure:"bech32_prefix"`
	GasAdjustment    float64 `mapstructure:"gas_adjustment"`
	GasPrice         string  `mapstructure:"gas_price"`
	Bin              string  `mapstructure:"bin"`
	TrustingPeriod   string  `mapstructure:"trusting_period"`
}

func LoadConfig(path string) Config {
	var config Config
	vp := viper.New()
	vp.SetConfigName("config")
	vp.SetConfigType("json")
	vp.AddConfigPath(path)

	vp.ReadInConfig()
	vp.Unmarshal(&config)
	return config
}

func TestConfig(t *testing.T) {
	config := LoadConfig(".")
	fmt.Println(config.Environment)
}
