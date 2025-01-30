package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/dfns/dfns-sdk-go/credentials"
	api "github.com/dfns/dfns-sdk-go/dfnsapiclient"
)

type WalletResponse struct {
	ID          string `json:"id"`
	Network     string `json:"network"`
	Status      string `json:"status"`
	Name        string `json:"name"`
	Address     string `json:"address"`
	DateCreated string `json:"dateCreated"`
	SigningKey  struct {
		Curve     string `json:"curve"`
		Scheme    string `json:"scheme"`
		PublicKey string `json:"publicKey"`
	} `json:"signingKey"`
}

func main() {
	conf := &credentials.AsymmetricKeySignerConfig{
		PrivateKey: `-----BEGIN PRIVATE KEY-----
MIIEvwIBADANBgkqhkiG9w0BAQEFAASCBKkwggSlAgEAAoIBAQDnUGLYl9Sii0LQ
3kCuZ/msgltOTH4KMtUmDDVuIwI6pwrnFPu65/K0BkssiStBbZ1rvPcMyTZtzY+2
fuFWq5I64EpxiyMmmMEYE8xsuVeBByt8uyDc/RG4DZl8ytD4MV4VjIG/3RoI67xN
ugVzBVyY0C8INeeigFB2FEpaJVInIX3YXAXGGAQX4FbKjmOpz874M8jGuRQ9ywwQ
B668FdsmqE72zvWArs5lgin/JWHffrUTqaOJVIkrwSpOQLlXMslR6MLdC09BtTGb
SlJPjyhqJ6NGYlSzSorWWa7TVm7v3fEd1SgrdlJablec3cqWrQIdSWnHnUUb2a6h
MITsLuTvAgMBAAECggEAFBscJGh6HpDNQ3t2EtLN1G1WQ2WJtRg7B74a7NJVMMTx
QSnFQbdElNpRMXNQ07ST8Nfxf2aD+SJbggjgTFjIcp6pSdpGuWWgrxeVdsPbc8cs
RAC9+Ad1QVLJSxwT8ubTnzrh0kwFJc5bxUPMknVeWZgK5oNM+YZ+t+zGk7RXwjfQ
Owbl4uGSlxJV+gMldGN/BgbzcvHnEAte6qP3jhFT1rqUPiZqHzN2Sl/VTp6gvufZ
9j+0PIErxFXpUIB2ixBeb0oON605MhlMZXasGUC/OQEdMXtJEGVhT/FR6O2fatoT
uHqQjnQX1RwkFJbiRfLh360TdmV24aVyz7sAZ1lLOQKBgQD0zgJJM/UesDGGqY0c
scA5/+a922wHz9fMYgTrQN5tLgFnp2UWgZpmVW8bDOwoA9iDnQSxqEZ78eGuwAna
Hre8cBk2J1m0QF04MzWukq4KVBPLuLvhaFjf010FJInm6m7ZHBd4xFJzZoqBY+xD
TQmGW93q4a7r01Q6qT0uYOqi1wKBgQDx5HAe3UFSWvIbuYDAK5YuSADCHwiPUbse
34Rtx4ebRKYkDwH1NsdnA5kbGsMhS/6YjQ/VwoW8oEhu12ivj4h8xncbkCvwBFrZ
wdKNfvWLNiCBlDwe3XNVKXjWnI1JD8U3wRrGyxRmE+ktFLN5ARlBn4mZZDcjjX07
HI/+EGojqQKBgQDPMIhYwmPEPGU+TsQCtCI4NHBq4YnWNr+y2IbHQRi/mP9RZii1
Wq19zPMDFvXMjCy0f7FYV06IWlii2R+9fuAM2WdNIRLX7t220giuHrC0RyKV+lzx
UqpdjXsd/iXEzUdR82eeK7KIvxGcnyB4eXwFPj1dLPMp3qtcFp6UYSxU8wKBgQCh
UNEdJDz96TzxFGMyxV6getBWpU+WFNGPo0yz0Y80EzIhdgi+Ocv9fT7L/qsHN6EQ
p3JaAiIiS1pC4VElU7mYTNr9/MXwiVb1Rfde+b5TGsPO5sa5ZsIVl1TI+xqWEPhb
WGK9FI4EDV9B+z49gmgPhY/ERjsncKKFm2TD8Lp4EQKBgQDXZAa8Uxk1s6jG+cYm
gcv41AzHCBGao7k7m3vX2bHo9tv/csme5rGQFhQaBU+GCuY36v8+2avR7Ah1TMuJ
Bvj70B7yzpUKf7ER1GxiLbe9vBLLWY5FBl4kXjM4/2Jyj/5UD+zonqVZSnWFDoeP
V701scoVOM+NNsLxjmTV42SVOQ==
-----END PRIVATE KEY-----`, // Credential private key
		CredID: "Y2ktYzhxanQtZmYzYzgtaWtvNjRlbnJzZGE3cGpy", // Credential Id
	}

	signer := credentials.NewAsymmetricKeySigner(conf)

	// Create a DfnsApiClient instance
	authToken := "eyJ0eXAiOiJKV1QiLCJhbGciOiJFZERTQSJ9.eyJpc3MiOiJhdXRoLmRmbnMuaW8iLCJhdWQiOiJkZm5zOmF1dGg6dXNlciIsInN1YiI6Im9yLTcyamI3LWo5ZjNsLTk0cDl0NmlzcW9ybWpwczUiLCJqdGkiOiJ0by02bW9iZS1lbDZoNC05Nmw5Ymx1Y3I3a2htbGNqIiwiaHR0cHM6Ly9jdXN0b20vdXNlcm5hbWUiOiJTZXJ2aWNlIEFjY291bnQgMiIsImh0dHBzOi8vY3VzdG9tL2FwcF9tZXRhZGF0YSI6eyJ1c2VySWQiOiJ1cy03cjVzYS01MHJ2dS05aWM5MG82azhhMDRvb3RnIiwib3JnSWQiOiJvci03MmpiNy1qOWYzbC05NHA5dDZpc3Fvcm1qcHM1IiwidG9rZW5LaW5kIjoiU2VydmljZUFjY291bnQifSwiaWF0IjoxNzM4MjQ4MzE1LCJleHAiOjE4MDEzMjAzMTV9.e8JisHmSNfthg4ufRAPME-mNPPbKr76SmYfySwavedWA5tCf2HoX240OANxC2dRR8xoYsouxlSjxfWUguOLOBA"
	apiOptions, err := api.NewDfnsAPIOptions(&api.DfnsAPIConfig{
		AppID:     "ap-25rpo-57vo9-9e0bpihml41c1di8", // ID of the Application registered with DFNS
		AuthToken: &authToken,                        // an auth token
		BaseURL:   "https://api.dfns.io",             // base Url of DFNS API
	}, signer)
	if err != nil {
		fmt.Printf("Error creating DfnsApiOptions: %s", err)
		return
	}

	dfnsClient := api.CreateDfnsAPIClient(apiOptions)

	// Create wallet
	walletData := map[string]interface{}{
		"network":         "EthereumGoerli",
		"name":            "my-wallet",
		"delayDelegation": true,
	}

	jsonData, err := json.Marshal(walletData)
	if err != nil {
		fmt.Printf("error marshaling JSON: %v", err)
		return
	}

	req, err := http.NewRequest("POST", apiOptions.BaseURL+"/wallets", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("error creating POST request: %v", err)
		return
	}

	response, err := dfnsClient.Do(req)
	if err != nil {
		fmt.Printf("error creating wallet: %v", err)
		return
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK && response.StatusCode != http.StatusCreated {
		fmt.Printf("API error: %s\n", response.Status)
		return
	}

	var wallet WalletResponse
	err = json.NewDecoder(response.Body).Decode(&wallet)
	if err != nil {
		fmt.Printf("error decoding JSON response: %v", err)
		return
	}

	fmt.Printf("Wallet Created: %+v\n", wallet)
}
