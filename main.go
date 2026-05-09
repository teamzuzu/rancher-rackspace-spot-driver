package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/rancher/kontainer-engine/types"
	"github.com/sirupsen/logrus"
	"github.com/teamzuzu/rancher-rackspace-spot-driver/driver"
)

func main() {
	if len(os.Args) < 2 {
		logrus.Fatal("usage: rancher-rackspace-spot-driver <port>")
	}

	if os.Getenv("DEBUG") != "" {
		logrus.SetLevel(logrus.DebugLevel)
	}

	port, err := strconv.Atoi(os.Args[1])
	if err != nil {
		logrus.Fatalf("invalid port %q: %v", os.Args[1], err)
	}

	addr := make(chan string)
	go types.NewServer(driver.NewDriver(), addr).ServeOrDie(fmt.Sprintf("127.0.0.1:%d", port))

	logrus.Infof("rackspace-spot driver listening on %v", <-addr)

	select {}
}
