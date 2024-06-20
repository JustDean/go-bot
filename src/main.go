package main

import (
	"context"
	"justdean/go-bot/src/poller"
	"justdean/go-bot/src/worker"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

// TODO add tests on everything

const ARG_ERR_MSG = "Valid arguments are \"poller\" and \"worker\"\n"

func main() {
	// TOKEN env is required for "poller" command
	// N_WORKERS env is required for "worker" command
	dispatch(os.Args)
}

func dispatch(arguments []string) {
	if len(arguments) < 2 {
		log.Fatalf("No arguments were passed. %s", ARG_ERR_MSG)
	}

	command := arguments[1]
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	if command == "poller" {
		done = handlePoller(ctx)

	} else if command == "worker" {
		done = handleWorker(ctx)

	} else {
		log.Fatalf("Invalid argument \"%s\". %s", command, ARG_ERR_MSG)
	}
	join(cancel, done)
}

func handlePoller(ctx context.Context) chan struct{} {
	token := os.Getenv("TOKEN")
	if token == "" {
		log.Fatalln("TOKEN env was not provided")
	}
	return poller.RunPoller(ctx, token)
}

func handleWorker(ctx context.Context) chan struct{} {
	// TODO make structure like handlePoller
	done := make(chan struct{})
	nWorkers := 1 // default workers amount
	nWorkersRaw := os.Getenv("N_WORKERS")
	if nWorkersRaw != "" {
		const ERR_MSG = "N_WORKERS env should be a positive number"
		newValue, err := strconv.Atoi(nWorkersRaw)
		if err != nil || newValue < 1 {
			log.Fatalln(ERR_MSG)
		}
		nWorkers = newValue
	}
	go worker.RunWorker(ctx, nWorkers, done)
	return done
}

func join(cancelContext context.CancelFunc, done chan struct{}) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs // waiting for termination signal
	log.Println("Shutting down")
	cancelContext()
	log.Println("Waiting for coroutines to finish their work")
	<-done
	log.Println("All done. Have a nice day <3")
}
