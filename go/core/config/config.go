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
		Port                             string  `mapstructure:"PORT"`
		Domain                           string  `mapstructure:"DOMAIN"`
		Env                              string  `mapstructure:"DIS_ENV"`
		BuildImage                       string  `mapstructure:"DIS_BUILD_IMAGE"`
		BuildCommit                      string  `mapstructure:"DIS_BUILD_COMMIT"`
		BuildDateTime                    string  `mapstructure:"DIS_BUILD_DATETIME"`
		BuildTag                         string  `mapstructure:"DIS_BUILD_TAG"`
		DisableCaches                    bool    `mapstructure:"DIS_DISABLE_CACHES"`
		GPT35TurboPromptCost             float64 `mapstructure:"DIS_GPT_35_TURBO_PROMPT_COST"`
		GPT35TurboResponseCost           float64 `mapstructure:"DIS_GPT_35_TURBO_RESPONSE_COST"`
		GPT4TurboPromptCost              float64 `mapstructure:"DIS_GPT_4_TURBO_PROMPT_COST"`
		GPT4TurboResponseCost            float64 `mapstructure:"DIS_GPT_4_TURBO_RESPONSE_COST"`
		LoggingLevel                     string  `mapstructure:"DIS_LOGGING_LEVEL"`
		LoggingOutput                    string  `mapstructure:"DIS_LOGGING_OUTPUT"`
		AnthropicKey                     string  `mapstructure:"DIS_ANTHROPIC_KEY"`
		AnyMailFinderKey                 string  `mapstructure:"DIS_ANYMAILFINDER_KEY"`
		FirebaseProject                  string  `mapstructure:"DIS_FIREBASE_PROJECT"`
		CoquiKey                         string  `mapstructure:"DIS_COQUI_KEY"`
		ElevenLabsKey                    string  `mapstructure:"DIS_ELEVENLABS_KEY"`
		JWTSessionSecret                 string  `mapstructure:"DIS_JWT_SESSION_SECRET"`
		DeepgramHost                     string  `mapstructure:"DIS_DEEPGRAM_HOST"`
		DeepgramKey                      string  `mapstructure:"DIS_DEEPGRAM_KEY"`
		DeepgramOnPremKey                string  `mapstructure:"DIS_DEEPGRAM_ONPREM_KEY"`
		ERCDataRoot                      string  `mapstructure:"DIS_ERC_DATA_ROOT"`
		ERCSpannerProject                string  `mapstructure:"DIS_ERC_SPANNER_PROJECT"`
		ERCSpannerInstance               string  `mapstructure:"DIS_ERC_SPANNER_INSTANCE"`
		ERCSpannerDatabase               string  `mapstructure:"DIS_ERC_SPANNER_DATABASE"`
		MailgunDomain                    string  `mapstructure:"DIS_MAILGUN_DOMAIN"`
		MailgunNotificationsFrom         string  `mapstructure:"DIS_MAILGUN_NOTIFICATIONS_FROM"`
		MailgunLowBalanceNotificationsTo string  `mapstructure:"DIS_MAILGUN_LOWBALANCE_NOTIFICATION_TO"`
		MailgunKey                       string  `mapstructure:"DIS_MAILGUN_KEY"`
		MondayKey                        string  `mapstructure:"DIS_MONDAY_KEY"`
		OpenAIKey                        string  `mapstructure:"DIS_OPENAI_KEY"`
		PineconeKey                      string  `mapstructure:"DIS_PINECONE_KEY"`
		RocketReachKey                   string  `mapstructure:"DIS_ROCKETREACH_KEY"`
		StabilityAIKey                   string  `mapstructure:"DIS_STABILITYAI_KEY"`
		STTRegion                        string  `mapstructure:"DIS_STT_REGION"`
		UserAgent                        string  `mapstructure:"DIS_USER_AGENT"`
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
