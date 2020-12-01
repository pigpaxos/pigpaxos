package paxi

import (
	"flag"
	"net/http"

	"github.com/pigpaxos/pigpaxos/log"
)

// Init setup paxi package
func Init() {
	flag.Parse()
	log.Setup()
	config.Load()
	http.DefaultTransport.(*http.Transport).MaxIdleConnsPerHost = 1000
}
