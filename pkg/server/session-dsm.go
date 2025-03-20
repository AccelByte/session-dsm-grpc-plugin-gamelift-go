// Copyright (c) 2024 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"session-dsm-grpc-plugin/pkg/constants"
	sessiondsm "session-dsm-grpc-plugin/pkg/pb"
	"session-dsm-grpc-plugin/pkg/utils/envelope"

	"github.com/AccelByte/accelbyte-go-sdk/session-sdk/pkg/sessionclient/game_session"
	"github.com/AccelByte/accelbyte-go-sdk/session-sdk/pkg/sessionclientmodels"
	"github.com/aws/aws-sdk-go-v2/service/gamelift"
	"github.com/aws/aws-sdk-go-v2/service/gamelift/types"
)

type AccelByteSessionClient interface {
	GetGameSessionShort(input *game_session.GetGameSessionParams) (*sessionclientmodels.ApimodelsGameSessionResponse, error)
}

type AmazonGameLiftClient interface {
	CreateGameSession(context.Context, *gamelift.CreateGameSessionInput, ...func(*gamelift.Options)) (*gamelift.CreateGameSessionOutput, error)
	TerminateGameSession(context.Context, *gamelift.TerminateGameSessionInput, ...func(*gamelift.Options)) (*gamelift.TerminateGameSessionOutput, error)
	StartGameSessionPlacement(context.Context, *gamelift.StartGameSessionPlacementInput, ...func(*gamelift.Options)) (*gamelift.StartGameSessionPlacementOutput, error)
}

type SessionDSM struct {
	sessiondsm.UnimplementedSessionDsmServer

	AwsAliasIdOverride  string
	AwsLocationOverride string
	AwsQueueArnOverride string

	SessionClient  AccelByteSessionClient
	GameLiftClient AmazonGameLiftClient
}

func NewSessionDSM(SessionClient AccelByteSessionClient, GameLiftClient AmazonGameLiftClient) *SessionDSM {
	sessionDsm := SessionDSM{
		SessionClient:  SessionClient,
		GameLiftClient: GameLiftClient,
	}

	// Useful for testing
	// Sets req.Deployment for all requests to CreateGameSession
	aliasIdOverride, ok := os.LookupEnv("AWS_ALIAS_ID_OVERRIDE")
	if ok && aliasIdOverride != "" {
		sessionDsm.AwsAliasIdOverride = aliasIdOverride
	}

	// Useful for testing with GameLift Servers Anywhere, which use custom locations
	// Sets req.RequestedRegion for all requests to CreateGameSession
	locationOverride, ok := os.LookupEnv("AWS_LOCATION_OVERRIDE")
	if ok && locationOverride != "" {
		sessionDsm.AwsLocationOverride = locationOverride
	}

	// Useful for testing GameLift queues/session placements
	// Directs all requests to CreateGameSessionAsync to the given Queue
	queueArnOverride, ok := os.LookupEnv("AWS_QUEUE_ARN_OVERRIDE")
	if ok && queueArnOverride != "" {
		sessionDsm.AwsQueueArnOverride = queueArnOverride
	}

	return &sessionDsm
}

func (s *SessionDSM) CreateGameSession(
	ctx context.Context,
	req *sessiondsm.RequestCreateGameSession,
) (*sessiondsm.ResponseCreateGameSession, error) {
	scope := envelope.NewRootScope(ctx, "CreateGameSession", "")
	defer scope.Finish()

	log := scope.Log

	var gameliftResponse *gamelift.CreateGameSessionOutput
	var err error

	if s.AwsAliasIdOverride != "" {
		log.Debugf("Using AWS Alias ID override: %v", s.AwsAliasIdOverride)
		req.Deployment = s.AwsAliasIdOverride
	}

	if s.AwsLocationOverride != "" {
		log.Debugf("Using AWS Location override: %v", s.AwsLocationOverride)
		req.RequestedRegion = []string{s.AwsLocationOverride}
	}

	if len(req.RequestedRegion) == 0 {
		log.Errorf("Requested region is required")
		return nil, errors.New("need provide requested region")
	}

	for _, region := range req.RequestedRegion {
		maxPlayersI32 := int32(req.MaximumPlayer)
		createGameSessionInput := &gamelift.CreateGameSessionInput{
			AliasId:                   &req.Deployment,
			IdempotencyToken:          &req.SessionId,
			MaximumPlayerSessionCount: &maxPlayersI32,
			Location:                  &region,
		}

		if req.SessionData != "" {
			createGameSessionInput.GameSessionData = &req.SessionData
		}

		gameliftResponse, err = s.GameLiftClient.CreateGameSession(ctx, createGameSessionInput)
		if err != nil {
			log.Warnf("Failed to create Game Session in region %s: %s", region, err)
			continue
		}

		break
	}

	if err != nil {
		log.Errorf("Failed to create session: %s", err)
		return nil, err
	}

	response := &sessiondsm.ResponseCreateGameSession{
		SessionId:     req.SessionId,
		Namespace:     req.Namespace,
		SessionData:   req.SessionData,
		ClientVersion: req.ClientVersion,
		GameMode:      req.GameMode,
		Source:        constants.GameServerSourceGamelift,
		Status:        constants.ServerStatusReady,
		Deployment:    *gameliftResponse.GameSession.GameSessionId,
		Ip:            *gameliftResponse.GameSession.IpAddress,
		Port:          int64(*gameliftResponse.GameSession.Port),
		ServerId:      *gameliftResponse.GameSession.GameSessionId,
		Region:        *gameliftResponse.GameSession.Location,
		CreatedRegion: *gameliftResponse.GameSession.Location,
	}

	log.Infof("Created session: %v", response)
	return response, nil
}

func (s *SessionDSM) TerminateGameSession(
	ctx context.Context,
	req *sessiondsm.RequestTerminateGameSession,
) (*sessiondsm.ResponseTerminateGameSession, error) {
	scope := envelope.NewRootScope(ctx, "TerminateGameSession", "")
	defer scope.Finish()

	log := scope.Log

	sessionInfo, err := s.SessionClient.GetGameSessionShort(&game_session.GetGameSessionParams{
		Namespace: req.Namespace,
		SessionID: req.SessionId,
	})
	if err != nil {
		log.Errorf("Failed to get session info while terminating game session: %v", err)
		return nil, err
	}
	serverInfo := sessionInfo.DSInformation.Server

	terminateSessionRequest := &gamelift.TerminateGameSessionInput{
		GameSessionId:   &serverInfo.Deployment,
		TerminationMode: types.TerminationModeTriggerOnProcessTerminate,
	}

	_, err = s.GameLiftClient.TerminateGameSession(ctx, terminateSessionRequest)
	if err != nil {
		log.Errorf("Failed to terminate game session: %v", err)
		return nil, err
	}

	response := &sessiondsm.ResponseTerminateGameSession{
		SessionId: req.SessionId,
		Namespace: req.Namespace,
		Success:   true,
	}

	log.Infof("Terminated session: %v", response)
	return response, nil
}

func (s *SessionDSM) CreateGameSessionAsync(
	ctx context.Context,
	req *sessiondsm.RequestCreateGameSession,
) (*sessiondsm.ResponseCreateGameSessionAsync, error) {
	scope := envelope.NewRootScope(ctx, "CreateGameSessionAsync", "")
	defer scope.Finish()

	log := scope.Log

	if s.AwsQueueArnOverride != "" {
		log.Debugf("Using AWS Queue ARN override: %v", s.AwsQueueArnOverride)
		req.Deployment = s.AwsQueueArnOverride
	}

	playerLatencies, err := extractLatencies(req.SessionData)
	if err != nil {
		log.WithError(err).Warnf("failed to parse player QoS data, continuing with session placement for session id %s", req.SessionId)
	}

	var response sessiondsm.ResponseCreateGameSessionAsync

	maxPlayersI32 := int32(req.MaximumPlayer)
	createSessionPlacementRequest := &gamelift.StartGameSessionPlacementInput{
		GameSessionQueueName:      &req.Deployment,
		MaximumPlayerSessionCount: &maxPlayersI32,
		PlacementId:               &req.SessionId,
	}

	if playerLatencies != nil {
		createSessionPlacementRequest.PlayerLatencies = playerLatencies
	}

	startPlacementResponse, err := s.GameLiftClient.StartGameSessionPlacement(ctx, createSessionPlacementRequest)
	if err != nil {
		response.Message = fmt.Sprintf("failed to start gamelift queue session placement for session: %s, Error: %v", req.SessionId, err)
		log.Errorf(response.Message)
		return &response, nil
	}

	if startPlacementResponse == nil || startPlacementResponse.GameSessionPlacement == nil {
		response.Message = fmt.Sprintf("failed to start gamelift queue session placement for session: %s", req.SessionId)
		log.Errorf(response.Message)
		return &response, nil
	}

	log.Infof("Successfully started Game Session Placement: %v", startPlacementResponse)

	response.Success = true
	return &response, nil
}

type MatchLatencyData struct {
	PlayerLatencies map[string]PlayerLatencyData `json:"gamelift_latencies"` // Player ID -> PlayerLatencyData
}
type PlayerLatencyData map[string]float32 // e.g. us-west-2 -> 42.5ms

func extractLatencies(sessionData string) ([]types.PlayerLatency, error) {
	var data MatchLatencyData
	if err := json.Unmarshal([]byte(sessionData), &data); err != nil {
		return nil, err
	}

	var latencies []types.PlayerLatency
	for playerId, playerLatency := range data.PlayerLatencies {
		for region, latency := range playerLatency {
			playerIdCopy := playerId
			regionCopy := region
			latencyCopy := latency
			latencies = append(latencies, types.PlayerLatency{
				PlayerId:              &playerIdCopy,
				RegionIdentifier:      &regionCopy,
				LatencyInMilliseconds: &latencyCopy,
			})
		}
	}

	if len(latencies) == 0 {
		return nil, fmt.Errorf("failed to find any player latencies in SessionData")
	}

	return latencies, nil
}
