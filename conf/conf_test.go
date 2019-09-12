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
	Env      string `env:"ENV,required"`
	MyApiKey string `secret:"ApiKey"`
}

func TestParse_WithASecret(t *testing.T) {
	ctrl := gomock.NewController(t)
	secretsManager := mocks.NewMockSecretsManager(ctrl)
	_ = os.Setenv("ENV", "staging")
	secretPayload := `{"ApiKey": "1234567890"}`
	secretsManager.EXPECT().
		GetSecretValue(&secretsmanager.GetSecretValueInput{SecretId: aws.String("/my-ns/my-secret")}).
		Return(&secretsmanager.GetSecretValueOutput{SecretString: &secretPayload}, nil)
	var cfg config

	err := Parse(&cfg, "/my-ns/my-secret", secretsManager)

	assert.NoError(t, err)
	assert.Equal(t, "staging", cfg.Env)
	assert.Equal(t, "1234567890", cfg.MyApiKey)
}

func TestParse_ErrorIfAWSErr(t *testing.T) {
	ctrl := gomock.NewController(t)
	secretsManager := mocks.NewMockSecretsManager(ctrl)
	_ = os.Setenv("ENV", "staging")
	secretsManager.EXPECT().
		GetSecretValue(&secretsmanager.GetSecretValueInput{SecretId: aws.String("/my-ns/my-secret")}).
		Return(nil, errors.New("failed in AWS"))
	var cfg config

	err := Parse(&cfg, "/my-ns/my-secret", secretsManager)

	assert.EqualError(t, err, "failed in AWS")
}

func TestParse_ErrorIfNoSecret(t *testing.T) {
	ctrl := gomock.NewController(t)
	secretsManager := mocks.NewMockSecretsManager(ctrl)
	_ = os.Setenv("ENV", "staging")
	secretPayload := `{}`
	secretsManager.EXPECT().
		GetSecretValue(&secretsmanager.GetSecretValueInput{SecretId: aws.String("/my-ns/my-secret")}).
		Return(&secretsmanager.GetSecretValueOutput{SecretString: &secretPayload}, nil)
	var cfg config

	err := Parse(&cfg, "/my-ns/my-secret", secretsManager)

	assert.EqualError(t, err, "required secret 'ApiKey' is not set")
}

func TestParse_ErrorIfPayloadNotJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	secretsManager := mocks.NewMockSecretsManager(ctrl)
	_ = os.Setenv("ENV", "staging")
	secretPayload := `plaintext secret`
	secretsManager.EXPECT().
		GetSecretValue(&secretsmanager.GetSecretValueInput{SecretId: aws.String("/my-ns/my-secret")}).
		Return(&secretsmanager.GetSecretValueOutput{SecretString: &secretPayload}, nil)
	var cfg config

	err := Parse(&cfg, "/my-ns/my-secret", secretsManager)

	assert.EqualError(t, err, "expected a JSON payload. Plaintext secrets are not currently supported")
}
