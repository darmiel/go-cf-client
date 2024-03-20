package main

import (
	"btp-service/pkg/btp"
	"fmt"
)

func main() {
	config, err := btp.LoadConfig()
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v\n", config)

	requester, err := config.GetRequester()
	if err != nil {
		panic(err)
	}

	fmt.Println(requester.ListSpaces(btp.ListSpacesOptions{}))
}
