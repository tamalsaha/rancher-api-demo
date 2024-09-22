package main

import (
	"errors"
	"fmt"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/restmapper"
	"k8s.io/klog/v2"
	"net/url"
	"os"
	"sigs.k8s.io/yaml"
	"time"

	"github.com/rancher/norman/clientbase"
	rancher "github.com/rancher/rancher/pkg/client/generated/management/v3"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/discovery/cached/memory"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	clustermeta "kmodules.xyz/client-go/cluster"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

func NewClient() (client.Client, error) {
	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)

	ctrl.SetLogger(klog.NewKlogr())
	cfg := ctrl.GetConfigOrDie()
	cfg.QPS = 100
	cfg.Burst = 100

	hc, err := rest.HTTPClientFor(cfg)
	if err != nil {
		return nil, err
	}
	mapper, err := apiutil.NewDynamicRESTMapper(cfg, hc)
	if err != nil {
		return nil, err
	}

	return client.New(cfg, client.Options{
		Scheme: scheme,
		Mapper: mapper,
		//Opts: client.WarningHandlerOptions{
		//	SuppressWarnings:   false,
		//	AllowDuplicateLogs: false,
		//},
	})
}

func DetectRancherProxy(cfg *rest.Config) (*clientbase.ClientOpts, bool, error) {
	err := rest.LoadTLSFiles(cfg)
	if err != nil {
		return nil, false, err
	}

	kc, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, false, err
	}

	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(kc))
	if clustermeta.IsRancherManaged(mapper) {
		u, err := url.Parse(cfg.Host)
		if err != nil {
			return nil, false, err
		}
		u.Path = "/v3"

		opts := clientbase.ClientOpts{
			URL:      u.String(),
			TokenKey: cfg.BearerToken,
			CACerts:  string(cfg.CAData),
			// Insecure:   true,
		}
		_, err = rancher.NewClient(&opts)
		return &opts, err == nil, err
	}
	return nil, false, nil
}

func main() {
	//_, yes, err := DetectRancherProxy(ctrl.GetConfigOrDie())
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Println("Rancher proxy enabled:", yes)

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
	rc, err := rancher.NewClient(&opts)
	if err != nil {
		var apiErr *clientbase.APIError
		if errors.As(err, &apiErr) {
			fmt.Println(apiErr.StatusCode)
		}
		return err
	}

	//kluster, err := rc.Cluster.ByID("c-m-8nmjt9cj")
	//if err != nil {
	//	return err
	//}
	//caCrt, err := base64.StdEncoding.DecodeString(kluster.CACert)
	//if err != nil {
	//	return err
	//}
	//fmt.Println(string(caCrt))

	//coll, err := rc.Token.ListAll(&types.ListOpts{})
	//if err != nil {
	//	return err
	//}
	//for _, item := range coll.Data {
	//	fmt.Println(item.Name)
	//}

	token, err := rc.Token.ByID("kubeconfig-user-nzj6blxgh2")
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

	nt, err := rc.Token.Create(&rancher.Token{
		// Name:        "tamal-token",
		Description: fmt.Sprintf("monitoring-operator-%d", time.Now().Unix()),
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
