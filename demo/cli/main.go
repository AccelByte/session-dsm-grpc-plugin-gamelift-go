// Copyright (c) 2024 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/AccelByte/accelbyte-go-sdk/iam-sdk/pkg/iamclient/users"
	"github.com/AccelByte/accelbyte-go-sdk/iam-sdk/pkg/iamclient/users_v4"
	"github.com/AccelByte/accelbyte-go-sdk/iam-sdk/pkg/iamclientmodels"
	"github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/factory"
	"github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/repository"
	"github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/service/iam"
	"github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/utils/auth"

	sessiondsmdemo "session-dsm-grpc-plugin-server-go-cli/pkg"
)

func main() {
	config, err := sessiondsmdemo.GetConfig()
	if err != nil {
		log.Fatalf("Can't retrieve config: %s\n", err)
	}

	configRepo := auth.DefaultConfigRepositoryImpl()
	tokenRepo := auth.DefaultTokenRepositoryImpl()

	oauthService := &iam.OAuth20Service{
		Client:           factory.NewIamClient(configRepo),
		ConfigRepository: configRepo,
		TokenRepository:  tokenRepo,
	}

	fmt.Print("Login to AccelByte... ")
	err = oauthService.LoginClient(&config.ABClientID, &config.ABClientSecret)
	if err != nil {
		log.Fatalf("Accelbyte account login failed: %s\n", err)
	}
	fmt.Println("[OK]")

	usersService := &iam.UsersV4Service{
		Client:           factory.NewIamClient(configRepo),
		ConfigRepository: configRepo,
		TokenRepository:  tokenRepo,
	}
	verified := true
	nameId := sessiondsmdemo.RandomString("ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789", 8)
	dName := "Extend Test User " + nameId
	username := fmt.Sprintf("extend_%s_user", nameId)
	email := username + "@dummy.net"
	country := "ID"
	dob := "1990-01-01"
	password := sessiondsmdemo.RandomString("ABCDEFGHIJKlmnopqrstuvwxyz0123456789!@#$%^&", 16)

	var acceptedPolicies []*iamclientmodels.LegalAcceptedPoliciesRequest
	authType := iamclientmodels.AccountCreateTestUserRequestV4AuthTypeEMAILPASSWD
	userInfo, err := usersService.PublicCreateTestUserV4Short(&users_v4.PublicCreateTestUserV4Params{
		Body: &iamclientmodels.AccountCreateTestUserRequestV4{
			AcceptedPolicies:  acceptedPolicies,
			AuthType:          &authType,
			Country:           &country,
			DateOfBirth:       &dob,
			DisplayName:       &dName,
			EmailAddress:      &email,
			Password:          &password,
			UniqueDisplayName: dName,
			Username:          &username,
			Verified:          &verified,
		},
		Namespace: os.Getenv("AB_NAMESPACE"),
	})
	if err != nil {
		log.Fatalf("Get user info failed: %s\n", err)
	}
	fmt.Printf("Test User Created: %s\n", *userInfo.UserID)

	// Start testing
	err = startTesting(userInfo, config, configRepo, tokenRepo)
	if err != nil {
		fmt.Println("\n[FAILED]")
		log.Fatal(err)
	}
	fmt.Println("\n[SUCCESS]")
}

func startTesting(
	userInfo *iamclientmodels.AccountCreateUserResponseV4,
	config *sessiondsmdemo.Config,
	configRepo repository.ConfigRepository,
	tokenRepo repository.TokenRepository) error {
	pdu := sessiondsmdemo.SessionDataUnit{
		CLIConfig:  config,
		ConfigRepo: configRepo,
		TokenRepo:  tokenRepo,
	}

	// clean up
	defer func() {
		fmt.Println("\nCleaning up...")

		fmt.Print("Deleting game session... ")
		err := pdu.DeleteGameSession()
		if err != nil {
			return
		}
		fmt.Println("[OK]")

		fmt.Print("Deleting Test User... ")
		err = deleteUser(userInfo, configRepo, tokenRepo)
		if err != nil {
			return
		}
		fmt.Println("[OK]")

		err = pdu.UnsetSessionServiceGrpcTarget()
		if err != nil {
			fmt.Printf("failed to unset session service grpc plugin url")

			return
		}
	}()

	// 1.
	fmt.Print("Configuring session service grpc target... ")
	err := pdu.CreateSessionConfiguration()
	if err != nil {
		fmt.Println("[ERR]")

		return err
	}
	fmt.Println("[OK]")

	// 2.
	fmt.Print("Creating game session... ")
	err = pdu.CreateGameSession(*userInfo.UserID)
	if err != nil {
		fmt.Println("[ERR]")

		return err
	}
	fmt.Println("[OK]")

	// 3.
	fmt.Print("Waiting game session... ")
	err = pdu.GetGameSession()
	if err != nil {
		fmt.Println("[ERR]")

		return err
	}
	fmt.Println("[FOUND]")

	fmt.Println("[OK]")

	return nil
}

func deleteUser(
	userInfo *iamclientmodels.AccountCreateUserResponseV4,
	configRepo repository.ConfigRepository,
	tokenRepo repository.TokenRepository) error {
	userService := &iam.UsersService{
		Client:           factory.NewIamClient(configRepo),
		ConfigRepository: configRepo,
		TokenRepository:  tokenRepo,
	}
	errDelete := userService.AdminDeleteUserInformationV3Short(&users.AdminDeleteUserInformationV3Params{
		Namespace: *userInfo.Namespace,
		UserID:    *userInfo.UserID,
	})
	if errDelete != nil {
		return errDelete
	}

	return nil
}
