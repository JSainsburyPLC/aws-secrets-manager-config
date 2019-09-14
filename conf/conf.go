package conf

import (
	"encoding/json"
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
func Parse(x interface{}, secretsManager secretsmanageriface.SecretsManagerAPI) error {
	confType := reflect.TypeOf(x).Elem()
	for i := 0; i < confType.NumField(); i++ {
		field := confType.Field(i)
		confField := reflect.ValueOf(x).Elem().Field(i)
		tag := field.Tag.Get("secret")
		if tag != "" {
			secret, err := fetchAWSSecret(secretsManager, tag)
			if err != nil {
				return err
			}

			jsonSecrets := []byte(secret)
			if json.Valid(jsonSecrets) {
				if confField.Kind() != reflect.Struct {
					return fmt.Errorf("secret value is JSON but the struct field is a string")
				}
				obj := reflect.New(confField.Type()).Interface()
				err = json.Unmarshal(jsonSecrets, &obj)
				if err != nil {
					return err
				}
				if confField.IsValid() && confField.CanSet() {
					val := reflect.ValueOf(obj)
					confField.Set(val.Elem())
					var unsetFields []string
					for i := 0; i < val.Elem().NumField(); i++ {
						field := val.Elem().Field(i)
						if field.IsZero() {
							unsetFields = append(unsetFields, val.Elem().Type().Field(i).Name)
						}
					}
					if len(unsetFields) > 0 {
						return fmt.Errorf("secrets not defined in configuration %+v", unsetFields)
					}
					continue
				}
			}

			if confField.Kind() != reflect.String {
				return fmt.Errorf("incorrect type when attempting to set plaintext secret. Expected a string for field '%s'", tag)
			}
			if confField.IsValid() && confField.CanSet() {
				confField.SetString(secret)
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
