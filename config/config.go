package config

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/docs/v1"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type Config struct {
	Sheets *sheets.Service
	Docs   *docs.Service
	Drive  *drive.Service
}

func Init() Config {
	var cfg Config
	err := cfg.getClientSheets()
	if err != nil {
		log.Fatal(err)
	}

	err = cfg.getClientDocs()
	if err != nil {
		log.Fatal(err)
	}

	err = cfg.getClientDrive()
	if err != nil {
		log.Fatal(err)
	}

	return cfg
}

func (cfg *Config) getClientSheets() error {
	ctx := context.Background()
	b, err := ioutil.ReadFile("files/credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
		return err
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/spreadsheets")
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
		return err
	}
	client := cfg.getClient(config, "files/token_sheets.json")

	srv, err := sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Sheets client: %v", err)
		return err
	}

	cfg.Sheets = srv

	return nil
}

func (cfg *Config) getClientDocs() error {
	ctx := context.Background()
	b, err := ioutil.ReadFile("files/credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
		return err
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/documents")
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
		return err
	}
	client := cfg.getClient(config, "files/token_docs.json")

	srv, err := docs.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Docs client: %v", err)
		return err
	}

	cfg.Docs = srv

	return nil
}

func (cfg *Config) getClientDrive() error {
	ctx := context.Background()

	b, err := ioutil.ReadFile("files/credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved credentials
	// at files/token_drive.json
	config, err := google.ConfigFromJSON(b, drive.DriveScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := cfg.getClient(config, "files/token_drive.json")

	srv, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve drive Client %v", err)
	}

	cfg.Drive = srv

	return nil
}

// Retrieve a token, saves the token, then returns the generated client.
func (cfg *Config) getClient(config *oauth2.Config, tokenFileName string) *http.Client {
	// The file tokenFileName.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tok, err := cfg.tokenFromFile(tokenFileName)
	if err != nil {
		tok = cfg.getTokenFromWeb(config)
		cfg.saveToken(tokenFileName, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func (cfg *Config) getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func (cfg *Config) tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func (cfg *Config) saveToken(path string, token *oauth2.Token) {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}
