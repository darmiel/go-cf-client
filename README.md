# go-cf-client

Simple client for Cloud Foundry API

## Usage

**config.json**

```json
{
  "username": "hello@world.io",
  "password": "hello-world-123",
  "api_endpoint": "https://api.cf.eu12.hana.ondemand.com",
  "auth_endpoint": "https://login.cf.eu12.hana.ondemand.com",
  "oauth_client_id": "cf",
  "oauth_client_secret": ""
}
```

**main.go**

```go
package main

import (
	"fmt"
	"github.com/darmiel/go-cf-client/pkg/cf"
)

func main() {
	config, err := cf.LoadCloudFoundryConfig("config.json")
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v\n", config)

	client, err := config.NewClient()
	if err != nil {
		panic(err)
	}

	space, err := client.CreateSpace(
		// name of the space
		"daniel-test-123",
		// guid of the organization
		"00000000-0000-0000-0000-000000000000",
		// more optional options
		cf.CreateSpaceOptions{})
	if err != nil {
		panic(err)
	}

	fmt.Println("Created Space:", space)
}
```