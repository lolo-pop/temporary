package gateway

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

var (
	user       string
	password   string
	gatewayUrl string
)

type function struct {
	Name string `json:"name"`
}

func init() {

	var ok bool
	gatewayUrl, ok = os.LookupEnv("GATEWAY_URL")
	if !ok {
		log.Fatal("$GATEWAY_URL not set\n")
	}

	// creates the in-cluster config
	// config, err := rest.InClusterConfig()
	// if err != nil {
	// 	log.Fatal(err.Error())
	// }

	// creates the clientset
	// clientset, err := kubernetes.NewForConfig(config)
	// if err != nil {
	//	log.Fatal(err.Error())
	// }

	// retrieve the secret
	// secrets := clientset.CoreV1().Secrets("openfaas")
	// authSecret, err := secrets.Get(context.TODO(), "basic-auth", metav1.GetOptions{})
	// if err != nil {
	// 	log.Fatal(err.Error())
	// }

	// set gateway authorization credentials
	// data := authSecret.Data
	// user = string(data["basic-auth-user"])
	// password = string(data["basic-auth-password"])
	user = "admin"
	password = "admin"
}

func Functions() ([]string, error) {

	// make http api request
	url := gatewayUrl + "/system/functions"
	resBody, err := apiRequest(url, "GET", nil)
	if err != nil {
		return nil, err
	}

	// unmarshal the request body
	var functions []function
	err = json.Unmarshal(resBody, &functions)
	if err != nil {
		return nil, err
	}

	var fnames []string
	for _, f := range functions {
		fnames = append(fnames, f.Name)
	}

	return fnames, nil
}

func apiRequest(url, method string, body io.Reader) ([]byte, error) {

	client := &http.Client{
		Timeout: time.Second * 10,
	}
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(user, password)
	response, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	resBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	return resBody, nil
}
