package conf

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-sdk-go/service/secretsmanager/secretsmanageriface"
	"github.com/caarlos0/env/v6"
)

//go:generate mockgen -source env.go -destination mocks/env.go -package mocks SecretsManager

// Parse parses a struct containing `env` and `secret` tags. `env` tags are resolved from the environment and `secret`
// tags are resolved from AWS secrets manager. The secrets in AWS must be defined as JSON with string values.
func Parse(x interface{}, secretKey string, secretsManager secretsmanageriface.SecretsManagerAPI) error {
	secret, err := fetchAWSSecret(secretsManager, secretKey)
	if err != nil {
		return err
	}
	jsonSecrets := []byte(secret)
	if !json.Valid(jsonSecrets) {
		return errors.New("expected a JSON payload. Plaintext secrets are not currently supported")
	}

	var secretsMap map[string]string
	err = json.Unmarshal(jsonSecrets, &secretsMap)
	if err != nil {
		return err
	}

	confType := reflect.TypeOf(x).Elem()
	for i := 0; i < confType.NumField(); i++ {
		field := confType.Field(i)
		tag := field.Tag.Get("secret")
		if tag != "" {
			secretValue, ok := secretsMap[tag]
			if !ok {
				return fmt.Errorf(`required secret '%s' is not set`, tag)
			}

			confField := reflect.ValueOf(x).Elem().Field(i)
			if confField.Kind() != reflect.String {
				return fmt.Errorf("incorrect type. Expected a string for field '%s'", tag)
			}
			if confField.IsValid() && confField.CanSet() {
				confField.SetString(secretValue)
			}
		}
	}

	return env.Parse(x)
}

func fetchAWSSecret(secretsManager secretsmanageriface.SecretsManagerAPI, key string) (string, error) {
	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(key),
	}
	output, err := secretsManager.GetSecretValue(input)
	if err != nil {
		return "", err
	}
	return *output.SecretString, nil
}
