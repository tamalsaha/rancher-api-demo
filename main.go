package main

import (
	"fmt"
	"os"
	"sigs.k8s.io/yaml"
	"time"

	"github.com/rancher/norman/clientbase"
	"github.com/rancher/norman/types"
	client "github.com/rancher/rancher/pkg/client/generated/management/v3"
)

func main() {
	err := doStuff()
	if err != nil {
		panic(err)
	}
}

func doStuff() error {
	opts := clientbase.ClientOpts{
		URL:        "https://172.233.204.178/v3",
		TokenKey:   os.Getenv("RANCHER_API_TOKEN"),
		Timeout:    0,
		HTTPClient: nil,
		WSDialer:   nil,
		CACerts:    "",
		Insecure:   true,
		ProxyURL:   "",
	}
	rc, err := client.NewClient(&opts)
	if err != nil {
		return err
	}
	coll, err := rc.Token.ListAll(&types.ListOpts{})
	if err != nil {
		return err
	}
	for _, item := range coll.Data {
		fmt.Println(item.Name)
	}

	token, err := rc.Token.ByID("token-zzknb")
	if err != nil {
		return err
	}
	// "2024-10-11T03:50:02Z"
	if token.Expired {
		return fmt.Errorf("token expired")
	}
	expiresAt, err := time.Parse(time.RFC3339, token.ExpiresAt)
	if err != nil {
		return err
	}
	fmt.Println(token.ExpiresAt, expiresAt.Format(time.RFC3339))

	nt, err := rc.Token.Create(&client.Token{
		Description: fmt.Sprintf("monitoring-operator-token-%d", time.Now().Unix()),
		TTLMillis:   90 * 24 * time.Hour.Milliseconds(),
	})
	if err != nil {
		return err
	}
	data, err := yaml.Marshal(nt)
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}
