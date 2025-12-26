package aws

import (
	pkgerrors "cellar/pkg/errors"
	"context"
	"encoding/base64"
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

type IamRequestInfo struct {
	Method  string
	Url     string
	Body    string
	Headers string
}

func GetAwsIamRequestInfo(ctx context.Context, role string) (info IamRequestInfo, err error) {
	if err := pkgerrors.CheckContext(ctx); err != nil {
		return IamRequestInfo{}, err
	}

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return info, err
	}

	svc := sts.NewFromConfig(cfg)
	presignClient := sts.NewPresignClient(svc)

	// Presign the GetCallerIdentity request for Vault AWS IAM auth
	presignedReq, err := presignClient.PresignGetCallerIdentity(ctx,
		&sts.GetCallerIdentityInput{},
	)
	if err != nil {
		return info, err
	}

	// Marshal headers to JSON
	headersJson, err := json.Marshal(presignedReq.SignedHeader)
	if err != nil {
		return info, err
	}

	return IamRequestInfo{
		Method:  presignedReq.Method,
		Url:     base64.StdEncoding.EncodeToString([]byte(presignedReq.URL)),
		Headers: base64.StdEncoding.EncodeToString(headersJson),
		Body:    "",
	}, nil
}
