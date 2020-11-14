package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/cosmos/relayer/cmd"
)

var (
	home       string
	httpBind   string
	cmdPermits = []string{"tx"}
)

func init() {
	flag.StringVar(&home, "home", "../data/relayer", "home dir")
	flag.StringVar(&httpBind, "http.bind", ":8000", "http bind port")
}

func main() {
	flag.Parse()
	cmd.InitConfig(home)
	http.HandleFunc("/", handleExec)
	log.Println("http serve:", httpBind)
	log.Fatal(http.ListenAndServe(httpBind, nil))
}

type errorReply struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func handleExec(w http.ResponseWriter, r *http.Request) {
	var (
		err       error
		hasPermit bool
		args      = strings.Split(strings.TrimLeft(r.URL.Path, "/"), "/")
	)
	if len(args) <= 1 {
		handleWrite(w, r, errorReply{Code: http.StatusBadRequest, Message: fmt.Sprintf("invalid args:%s", args)})
		return
	}
	for _, w := range cmdPermits {
		if w == args[0] {
			hasPermit = true
		}
	}
	if !hasPermit {
		handleWrite(w, r, errorReply{Code: http.StatusUnauthorized, Message: fmt.Sprintf("unauthorized args:%s", args)})
		return
	}
	cmd.RootCmd.SetArgs(args)
	if err = cmd.RootCmd.ExecuteContext(r.Context()); err != nil {
		handleWrite(w, r, errorReply{Code: http.StatusInternalServerError, Message: fmt.Sprintf("error:%v", err)})
	} else {
		handleWrite(w, r, errorReply{Code: http.StatusOK, Message: "OK"})
	}
}

func handleWrite(w http.ResponseWriter, r *http.Request, reply errorReply) {
	b, _ := json.Marshal(reply)
	w.WriteHeader(reply.Code)
	w.Write(b)
	// access logs
	log.Printf("handleWrite path: %s returns: %s", r.URL.Path, b)
}
