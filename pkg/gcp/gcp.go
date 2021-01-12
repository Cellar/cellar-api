package gcp

import (
	"context"
	"encoding/json"
	"fmt"
	"google.golang.org/api/iam/v1"
	"time"
)
type AuthInfo struct{
	sub string
	aud string
	exp string
}

func GetGcpRequestInfo(role, serviceAccountEmail string) (signedJwt string, err error) {
	ctx := context.Background()
	iamClient, err := iam.NewService(ctx)
	if err != nil {
		return "", err
	}

	resourceName := fmt.Sprintf("projects/-/serviceAccounts/%s", serviceAccountEmail)
	jwtPayload := map[string]interface{}{
		"aud": fmt.Sprintf("vault/%s", role),
		"sub": serviceAccountEmail,
		"exp": time.Now().Add(time.Hour * 1).Unix(),
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
