package aws

import (
	"encoding/base64"
	"encoding/json"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	"io/ioutil"
)

type AwsIamRequestInfo struct {
	RequestUrl string
	RequestBody string
	RequestHeaders string
}

func GetAwsIamRequestInfo(role string) (info AwsIamRequestInfo, err error) {
	stsSession, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1")})
	if err != nil {
		return info, err
	}

	var params *sts.GetCallerIdentityInput
	svc := sts.New(stsSession)
	stsRequest, _ := svc.GetCallerIdentityRequest(params)

	value, err := stsSession.Config.Credentials.Get()
	if err != nil {
		return info, err
	}

	stsRequest.HTTPRequest.Method = "POST"
	stsRequest.HTTPRequest.Header.Add("User-Agent", role)
	stsRequest.HTTPRequest.Header.Add("X-Amz-Security-Token", value.SecretAccessKey)
	stsRequest.HTTPRequest.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=utf-8")
	if err = stsRequest.Sign(); err != nil {
		return
	}

	headersJson, err := json.Marshal(stsRequest.HTTPRequest.Header)
	if err != nil {
		return info, err
	}
	requestBody, err := ioutil.ReadAll(stsRequest.HTTPRequest.Body)
	if err != nil {
		return info, err
	}
	return AwsIamRequestInfo{
		RequestUrl:     base64.StdEncoding.EncodeToString([]byte(stsRequest.HTTPRequest.URL.String())),
		RequestHeaders:    base64.StdEncoding.EncodeToString(headersJson),
		RequestBody: base64.StdEncoding.EncodeToString(requestBody),
	}, nil
}

func stsSigningResolver(service, region string, optFns ...func(*endpoints.Options)) (endpoints.ResolvedEndpoint, error) {
	defaultEndpoint, err := endpoints.DefaultResolver().EndpointFor(service, region, optFns...)
	if err != nil {
		return defaultEndpoint, err
	}
	defaultEndpoint.SigningRegion = region
	return defaultEndpoint, nil
}
