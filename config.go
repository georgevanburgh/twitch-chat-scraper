package twitchchatscraper

type Config struct {
	TwitchUsername   string `json:"twitch_username"`
	TwitchOAuthToken string `json:"twitch_oauth_token"`
}

var Configuration *Config

func SetConfig(givenConfig *Config) {
	Configuration = givenConfig
}
