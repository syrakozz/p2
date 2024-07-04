// Package config handles configuration and environment vars.
package config

import (
	"bytes"
	_ "embed"
	"log"

	"github.com/spf13/viper"
)

var (
	//go:embed config.env
	configFile []byte

	// VARS is the environment shared across all packages
	VARS struct {
		Port                        string `mapstructure:"PORT"`
		Env                         string `mapstructure:"DIS_ENV"`
		BuildImage                  string `mapstructure:"DIS_BUILD_IMAGE"`
		BuildCommit                 string `mapstructure:"DIS_BUILD_COMMIT"`
		BuildDateTime               string `mapstructure:"DIS_BUILD_DATETIME"`
		HTTPRetryCount              int    `mapstructure:"DIS_HTTP_RETRY_COUNT"`
		HTTPRetryMaxWaitTimeSeconds int64  `mapstructure:"DIS_HTTP_RETRY_MAX_WAIT_TIME_SECONDS"`
		LoggingLevel                string `mapstructure:"DIS_LOGGING_LEVEL"`
		LoggingOutput               string `mapstructure:"DIS_LOGGING_OUTPUT"`
		AnyMailFinderKey            string `mapstructure:"DIS_ANYMAILFINDER_KEY"`
		JWTSessionSecret            string `mapstructure:"DIS_JWT_SESSION_SECRET"`
		ERCDataRoot                 string `mapstructure:"DIS_ERC_DATA_ROOT"`
		ERCSpannerProject           string `mapstructure:"DIS_ERC_SPANNER_PROJECT"`
		ERCSpannerInstance          string `mapstructure:"DIS_ERC_SPANNER_INSTANCE"`
		ERCSpannerDatabase          string `mapstructure:"DIS_ERC_SPANNER_DATABASE"`
		MondayAuthorization         string `mapstructure:"DIS_MONDAY_AUTHORIZATION"`
	}
)

func init() {
	viper.SetConfigType("env")

	if err := viper.ReadConfig(bytes.NewReader(configFile)); err != nil {
		log.Fatalf("unable to read config: %v\n", err)
	}

	viper.AutomaticEnv()

	if err := viper.Unmarshal(&VARS); err != nil {
		log.Fatalf("failed to unmarshal config: %v\n", err)
	}
}
