package gcp

import (
	"cloud.google.com/go/compute/metadata"
	"context"
	"encoding/json"
	"fmt"
	"google.golang.org/api/iam/v1"
	"time"
)

func GetGcpIamRequestInfo(role string) (signedJwt string, err error) {
	ctx := context.Background()
	iamClient, err := iam.NewService(ctx)
	if err != nil {
		return "", err
	}

	serviceAccountEmail, err := metadata.Email("")
	if err != nil {
		return "", err
	}
	resourceName := fmt.Sprintf("projects/-/serviceAccounts/%s", serviceAccountEmail)
	jwtPayload := map[string]interface{}{
		"aud": fmt.Sprintf("vault/%s", role),
		"sub": serviceAccountEmail,
		"exp": time.Now().Add(time.Minute * 5).Unix(),
	}

	payloadBytes, err := json.Marshal(jwtPayload)
	if err != nil {
		return "", err
	}
	signJwtReq := &iam.SignJwtRequest{
		Payload: string(payloadBytes),
	}

	resp, err := iamClient.Projects.ServiceAccounts.SignJwt(resourceName, signJwtReq).Do()
	if err != nil {
		return "", err
	}

	return resp.SignedJwt, nil
}
