package main

import (
	"os"
	"log"
	"os/exec"
	"net/http"
	"fmt"
	"net/url"
	"net/http/httputil"
	"github.com/cloudfoundry-community/go-cfenv"
	"github.com/Jeffail/gabs"
	"github.com/koding/websocketproxy"
)

const WIZZ_EXT = "wizz_ext"

func main() {
	var input string
	jsonParsed, _ := gabs.ParseJSON([]byte(os.Getenv(WIZZ_EXT)))
	startCmd := jsonParsed.Path("start").Data()
	sockIoUrl := jsonParsed.Path("io_url").Data()
	cfWizzBin := jsonParsed.Path("cfwizz_bin").Data()
	serverPort := jsonParsed.Path("server_port").Data()
	appUrl := jsonParsed.Path("app_url").Data()

	if startCmd != nil && cfWizzBin != nil && sockIoUrl != nil && serverPort != nil && appUrl !=nil {
		go reverseProxy(sockIoUrl.(string), serverPort.(string), appUrl.(string))
		go startAppProcess(startCmd.(string))
		go startCfWizz(cfWizzBin.(string))
		fmt.Scanln(&input)
	} else {
		log.Fatal("Required env variables are not detected! ")
	}
}

func reverseProxy(sockIoUrl string, serverPort string,appUrl string) {
	appEnv, _ := cfenv.Current()
	if (appEnv != nil) {
		fmt.Println("App urls:", appEnv.ApplicationURIs)
		log.Print("Starting reverse proxy!")
		http.HandleFunc("/info", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, "CfWizz Agent is running!")
		})

		u, _ := url.Parse(sockIoUrl)
		uApp, _ := url.Parse(appUrl)
		http.Handle("/cfwizz", websocketproxy.NewProxy(u))
		http.Handle("/", httputil.NewSingleHostReverseProxy(uApp))
		log.Fatal(http.ListenAndServe(serverPort, nil))
	}
}

func startCfWizz(cfWizzBin string) {
	fmt.Println("Starting cfwizz socket server:", cfWizzBin)
	lsCmd := exec.Command("bash", "-c", cfWizzBin)
	lsCmd.Run()
}

func startAppProcess(start string) {
	fmt.Println("Starting app server:", start)
	lsCmd := exec.Command("bash", "-c", start)
	lsCmd.Stdout = os.Stdout
	lsCmd.Stderr = os.Stderr
	err := lsCmd.Run()
	if err != nil {
		log.Print("Restartig app process..", err)
		startAppProcess(start)
	}
}

//for _, element := range appEnv.ApplicationURIs {}
