package service

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/taco-labs/taco/go/domain/value"
)

type EncryptionService interface {
	Encrypt(context.Context, string) ([]byte, error)
	Decrypt(context.Context, []byte) (string, error)
}

type awsKmsEncryptionService struct {
	client *kms.Client
	keyId  string
}

func (a awsKmsEncryptionService) Encrypt(ctx context.Context, data string) ([]byte, error) {
	req := kms.EncryptInput{
		KeyId:     &a.keyId,
		Plaintext: []byte(data),
	}

	resp, err := a.client.Encrypt(ctx, &req)
	if err != nil {
		return []byte{}, fmt.Errorf("%w: erorr from kms service: %v", value.ErrExternal, err)
	}

	return resp.CiphertextBlob, nil
}

func (a awsKmsEncryptionService) Decrypt(ctx context.Context, data []byte) (string, error) {
	req := kms.DecryptInput{
		KeyId:          &a.keyId,
		CiphertextBlob: data,
	}

	resp, err := a.client.Decrypt(ctx, &req)
	if err != nil {
		return "", fmt.Errorf("%w: erorr from kms service: %v", value.ErrExternal, err)
	}

	return string(resp.Plaintext), nil
}

func NewAwsKMSEncryptionService(client *kms.Client, keyId string) *awsKmsEncryptionService {
	return &awsKmsEncryptionService{
		client: client,
		keyId:  keyId,
	}
}
