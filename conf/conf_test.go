package conf

import (
	"aws-secrets-manager-config/mocks"
	"errors"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

type config struct {
	Env             string       `env:"ENV,required"`
	SecretPlaintext string       `secret:"/my-ns/my-plaintext-secret"`
	SecretStruct    secretStruct `secret:"/my-ns/my-struct-secret"`
}

type secretStruct struct {
	ApiKey string `json:"api_key"`
}

func TestParse_WithASecret(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	secretsManager := mocks.NewMockSecretsManager(ctrl)
	_ = os.Setenv("ENV", "staging")
	plaintextSecretPayload := "myPlaintextSecretValue"
	secretsManager.EXPECT().
		GetSecretValue(&secretsmanager.GetSecretValueInput{SecretId: aws.String("/my-ns/my-plaintext-secret")}).
		Return(&secretsmanager.GetSecretValueOutput{SecretString: &plaintextSecretPayload}, nil)
	jsonSecretPayload := `{"api_key": "1234567890"}`
	secretsManager.EXPECT().
		GetSecretValue(&secretsmanager.GetSecretValueInput{SecretId: aws.String("/my-ns/my-struct-secret")}).
		Return(&secretsmanager.GetSecretValueOutput{SecretString: &jsonSecretPayload}, nil)
	var cfg config

	err := Parse(&cfg, secretsManager)

	assert.NoError(t, err)
	assert.Equal(t, "staging", cfg.Env)
	assert.Equal(t, "myPlaintextSecretValue", cfg.SecretPlaintext)
	assert.Equal(t, secretStruct{ApiKey: "1234567890"}, cfg.SecretStruct)
}

func TestParse_ErrorIfAWSErr(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	secretsManager := mocks.NewMockSecretsManager(ctrl)
	_ = os.Setenv("ENV", "staging")
	secretsManager.EXPECT().
		GetSecretValue(&secretsmanager.GetSecretValueInput{SecretId: aws.String("/my-ns/my-plaintext-secret")}).
		Return(nil, errors.New("failed in AWS"))
	var cfg config

	err := Parse(&cfg, secretsManager)

	assert.EqualError(t, err, "failed in AWS")
}

func TestParse_ErrorIfSecretEmpty(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	secretsManager := mocks.NewMockSecretsManager(ctrl)
	_ = os.Setenv("ENV", "staging")
	secretPayload := `{"other": "12345"}`
	secretsManager.EXPECT().
		GetSecretValue(&secretsmanager.GetSecretValueInput{SecretId: aws.String("/my-ns/my-struct-secret")}).
		Return(&secretsmanager.GetSecretValueOutput{SecretString: &secretPayload}, nil)
	plaintextSecretPayload := "myPlaintextSecretValue"
	secretsManager.EXPECT().
		GetSecretValue(&secretsmanager.GetSecretValueInput{SecretId: aws.String("/my-ns/my-plaintext-secret")}).
		Return(&secretsmanager.GetSecretValueOutput{SecretString: &plaintextSecretPayload}, nil)
	var cfg config

	err := Parse(&cfg, secretsManager)

	assert.EqualError(t, err, "secrets not defined in configuration [ApiKey]")
}

func TestParse_ErrorIfUnsupportedType(t *testing.T) {
	type configWithInvalidType struct {
		MyApiKey int `secret:"/my-ns/my-plaintext-secret"`
	}
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	secretsManager := mocks.NewMockSecretsManager(ctrl)
	secretPayload := "abcdefg"
	secretsManager.EXPECT().
		GetSecretValue(&secretsmanager.GetSecretValueInput{SecretId: aws.String("/my-ns/my-plaintext-secret")}).
		Return(&secretsmanager.GetSecretValueOutput{SecretString: &secretPayload}, nil)
	var cfg configWithInvalidType

	err := Parse(&cfg, secretsManager)

	assert.EqualError(t, err,
		"incorrect type when attempting to set plaintext secret. Expected a string for field '/my-ns/my-plaintext-secret'")
}
