// Copyright (c) 2024 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package sessiondsmdemo

import "github.com/caarlos0/env/v7"

type Config struct {
	ABBaseURL      string `env:"AB_BASE_URL"`
	ABClientID     string `env:"AB_CLIENT_ID"`
	ABClientSecret string `env:"AB_CLIENT_SECRET"`
	ABNamespace    string `env:"AB_NAMESPACE"`
	GRPCServerURL  string `env:"GRPC_SERVER_URL"`
	ExtendAppName  string `env:"EXTEND_APP_NAME"`
}

func GetConfig() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
