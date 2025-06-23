package cache

import (
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-secretsmanager-caching-go/secretcache"
)

type SecretsCache struct {
	env   string
	cache AwsSecretsCache
}

type AwsSecretsCache interface {
	GetSecretString(secretId string) (string, error)
}

func applyAwsConfig(c *secretcache.Cache) {
	var region string
	var ok bool
	if region, ok = os.LookupEnv("AWS_REGION"); !ok {
		region = "eu-west-1"
	}

	sess, err := session.NewSession(&aws.Config{Region: &region})
	if err != nil {
		log.Fatal("session failed to be created", err)
	}

	if iamRole, ok := os.LookupEnv("AWS_IAM_ROLE"); ok {
		c := stscreds.NewCredentials(sess, iamRole)
		sess.Config.Credentials = c
	}

	endpoint := os.Getenv("AWS_SECRETS_MANAGER_ENDPOINT")
	sess.Config.Endpoint = &endpoint
	c.Client = secretsmanager.New(sess)
}

func New() *SecretsCache {
	env := os.Getenv("ENVIRONMENT")
	cache, _ := secretcache.New(applyAwsConfig)
	return &SecretsCache{env, cache}
}

func (c *SecretsCache) GetSecretString(key string) (string, error) {
	return c.cache.GetSecretString(c.env + "/" + key)
}

func (c *SecretsCache) GetGlobalSecretString(key string) (string, error) {
	return c.cache.GetSecretString(key)
}

func (c *SecretsCache) ClearCache() {
	c.cache, _ = secretcache.New(applyAwsConfig)
}
