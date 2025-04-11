package firebase

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"google.golang.org/api/option"
)

var FirebaseApp *firebase.App
var FirebaseClient *auth.Client

type FirebaseCredential struct {
	Type                    string `json:"type"`
	ProjectID               string `json:"project_id"`
	PrivateKeyID            string `json:"private_key_id"`
	PrivateKey              string `json:"private_key"`
	ClientEmail             string `json:"client_email"`
	ClientID                string `json:"client_id"`
	AuthURI                 string `json:"auth_uri"`
	TOkenURI                string `json:"token_uri"`
	AuthProviderX509CertURL string `json:"auth_provider_x509_cert_url"`
	ClientX509CertURL       string `json:"client_x509_cert_url"`
	UniverseDomain          string `json:"universe_domain"`
}

func InitFirebase() {
	credential := FirebaseCredential{
		Type:                    os.Getenv("TYPE"),
		ProjectID:               os.Getenv("PROJECT_ID"),
		PrivateKeyID:            os.Getenv("PRIVATE_KEY_ID"),
		PrivateKey:              strings.ReplaceAll(os.Getenv("PRIVATE_KEY"), "\\n", "\n"),
		ClientEmail:             os.Getenv("CLIENT_EMAIL"),
		ClientID:                os.Getenv("CLIENT_ID"),
		AuthURI:                 os.Getenv("AUTH_URI"),
		TOkenURI:                os.Getenv("TOKEN_URI"),
		AuthProviderX509CertURL: os.Getenv("AUTH_PROVIDER_X509_CERT_URL"),
		ClientX509CertURL:       os.Getenv("CLIENT_X509_CERT_URL"),
		UniverseDomain:          os.Getenv("UNIVERSE_DOMAIN"),
	}
	credentialJson, err := json.Marshal(credential)
	if err != nil {
		fmt.Printf("Error %s", err)
	}
	opt := option.WithCredentialsJSON(credentialJson)
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		fmt.Printf("error initializing app: %v \n", err)
	}
	FirebaseApp = app

	client, err := FirebaseApp.Auth(context.Background())
	if err != nil {
		fmt.Printf("error initializing app: %v \n", err)
	}
	FirebaseClient = client
}
