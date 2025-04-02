// Copyright (c) 2024 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package sessiondsmdemo

import (
	"fmt"
	"strconv"
	"time"

	"github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/factory"
	"github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/repository"
	"github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/service/session"
	"github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/utils"
	"github.com/AccelByte/accelbyte-go-sdk/session-sdk/pkg/sessionclient/configuration_template"
	"github.com/AccelByte/accelbyte-go-sdk/session-sdk/pkg/sessionclient/game_session"
	"github.com/AccelByte/accelbyte-go-sdk/session-sdk/pkg/sessionclientmodels"
)

type SessionDataUnit struct {
	CLIConfig  *Config
	ConfigRepo repository.ConfigRepository
	TokenRepo  repository.TokenRepository
	sessionID  string
	configName string
}

var prefix = "session_dsm_grpc_go"

func (p *SessionDataUnit) CreateSessionConfiguration() error {
	servicePluginCfgWrapper := session.ConfigurationTemplateService{
		Client:           factory.NewSessionClient(p.ConfigRepo),
		ConfigRepository: p.ConfigRepo,
		TokenRepository:  p.TokenRepo,
	}

	p.configName = fmt.Sprintf("%s_%s", prefix, RandomString("1234567890", 3))

	body := &sessionclientmodels.ApimodelsCreateConfigurationTemplateRequest{
		ClientVersion:    Ptr("test"),
		Deployment:       Ptr("test"),
		Persistent:       Ptr(false),
		TextChat:         Ptr(false),
		Name:             Ptr(p.configName),
		MinPlayers:       Ptr(int32(0)),
		MaxPlayers:       Ptr(int32(2)),
		Joinability:      Ptr("OPEN"),
		InviteTimeout:    Ptr(int32(60)),
		InactiveTimeout:  Ptr(int32(60)),
		AutoJoin:         true,
		Type:             Ptr("DS"),
		DsSource:         "custom",
		DsManualSetReady: false,
		RequestedRegions: []string{"us-west-2"},
	}

	if p.CLIConfig.GRPCServerURL != "" {
		fmt.Printf("(Custom Host: %s) ", p.CLIConfig.GRPCServerURL)

		body.CustomURLGRPC = p.CLIConfig.GRPCServerURL

		_, err := servicePluginCfgWrapper.AdminCreateConfigurationTemplateV1Short(&configuration_template.AdminCreateConfigurationTemplateV1Params{
			Body:      body,
			Namespace: p.CLIConfig.ABNamespace,
		})

		if err != nil {
			return err
		}

		return nil
	} else if p.CLIConfig.ExtendAppName != "" {
		fmt.Printf("(Extend App: %s) ", p.CLIConfig.ExtendAppName)

		body.AppName = p.CLIConfig.ExtendAppName

		_, err := servicePluginCfgWrapper.AdminCreateConfigurationTemplateV1Short(&configuration_template.AdminCreateConfigurationTemplateV1Params{
			Body:      body,
			Namespace: p.CLIConfig.ABNamespace,
		})

		if err != nil {
			return err
		}

		return nil
	} else {
		return fmt.Errorf("url or app name is not defined")
	}
}

func (p *SessionDataUnit) CreateGameSession(userID string) error {
	servicePluginCfgWrapper := session.GameSessionService{
		Client:           factory.NewSessionClient(p.ConfigRepo),
		ConfigRepository: p.ConfigRepo,
		TokenRepository:  p.TokenRepo,
	}

	userIDs := []string{userID}

	createGame, err := servicePluginCfgWrapper.CreateGameSessionShort(&game_session.CreateGameSessionParams{
		Body: &sessionclientmodels.ApimodelsCreateGameSessionRequest{
			ConfigurationName: Ptr(p.configName),
			Teams: []*sessionclientmodels.ModelsTeam{
				{
					Parties: []*sessionclientmodels.ModelsPartyMembers{
						{
							PartyID: "",
							UserIDs: userIDs,
						},
					},
					UserIDs: userIDs,
				},
			},
		},
		Namespace: p.CLIConfig.ABNamespace,
	})
	if err != nil {
		return err
	}

	p.sessionID = *createGame.ID

	return nil
}

func (p *SessionDataUnit) GetGameSession() error {
	servicePluginCfgWrapper := session.GameSessionService{
		Client:           factory.NewSessionClient(p.ConfigRepo),
		ConfigRepository: p.ConfigRepo,
		TokenRepository:  p.TokenRepo,
	}

	isDsAvailable := false
	dsChecks := 0
	maxDsChecks, errEnv := strconv.Atoi(utils.GetEnv("DS_CHECK_COUNT", "10"))
	if errEnv != nil {
		return errEnv
	}
	checkIntervalStr := utils.GetEnv("DS_WAIT_INTERVAL", "1")
	checkInterval, errEnv2 := strconv.ParseFloat(checkIntervalStr, 64)
	if errEnv2 != nil {
		return errEnv2
	}

	// Loop until the DSInformation.StatusV2 becomes "AVAILABLE" or max checks are reached
	for dsChecks < maxDsChecks {
		getGameSession, err := servicePluginCfgWrapper.GetGameSessionShort(&game_session.GetGameSessionParams{
			Namespace: p.CLIConfig.ABNamespace,
			SessionID: p.sessionID,
		})
		if err != nil {
			return err
		}

		if getGameSession.DSInformation.StatusV2 == "AVAILABLE" {
			isDsAvailable = true
			fmt.Println(" DS is AVAILABLE")

			break
		}

		time.Sleep(time.Duration(checkInterval) * 10 * time.Second)

		dsChecks++

		fmt.Printf("check %d/%d: DS not available yet. Retrying...\n", dsChecks, maxDsChecks)
	}

	if !isDsAvailable {
		return fmt.Errorf("dedicated Server is not available after maximum checks (%v)", maxDsChecks)
	}

	return nil
}

func (p *SessionDataUnit) DeleteGameSession() error {
	servicePluginCfgWrapper := session.GameSessionService{
		Client:           factory.NewSessionClient(p.ConfigRepo),
		ConfigRepository: p.ConfigRepo,
		TokenRepository:  p.TokenRepo,
	}

	return servicePluginCfgWrapper.DeleteGameSessionShort(&game_session.DeleteGameSessionParams{
		SessionID: p.sessionID,
		Namespace: p.CLIConfig.ABNamespace,
	})
}

func (p *SessionDataUnit) UnsetSessionServiceGrpcTarget() error {
	servicePluginCfgWrapper := session.ConfigurationTemplateService{
		Client:           factory.NewSessionClient(p.ConfigRepo),
		ConfigRepository: p.ConfigRepo,
		TokenRepository:  p.TokenRepo,
	}

	return servicePluginCfgWrapper.AdminDeleteConfigurationTemplateV1Short(&configuration_template.AdminDeleteConfigurationTemplateV1Params{
		Name:      p.configName,
		Namespace: p.CLIConfig.ABNamespace,
	})
}
