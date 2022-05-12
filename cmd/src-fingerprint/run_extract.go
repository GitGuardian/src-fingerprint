package main

import (
	"srcfingerprint"
	"sync"
	"time"

	"github.com/Jeffail/tunny"
	log "github.com/sirupsen/logrus"
)

type PoolPayload struct {
	object       string
	eventChannel chan srcfingerprint.PipelineEvent
	waitGroup    *sync.WaitGroup
}

// Run the extraction in separate goroutine for each objects using a pool of size poolSize.
func runExtract(
	pipeline *srcfingerprint.Pipeline,
	objects []string,
	after string,
	limit int,
	timeout time.Duration,
	poolSize int) chan srcfingerprint.PipelineEvent {
	// If there is no object, default to an empty object
	if len(objects) == 0 {
		objects = []string{""}
	}

	eventChannel := make(chan srcfingerprint.PipelineEvent, MaxPipelineEvents)

	wg := sync.WaitGroup{}
	pool := tunny.NewFunc(poolSize, func(rawPayload interface{}) interface{} {
		payload, ok := rawPayload.(PoolPayload)
		if !ok {
			// Impossible error, pool is only used with PoolPayload structs
			log.Fatal("Could not load the payload to run the extraction")
		}

		defer payload.waitGroup.Done()
		pipeline.ExtractRepositories(payload.object, after, payload.eventChannel, limit, timeout)

		return nil
	})

	for _, object := range objects {
		wg.Add(1)

		go pool.Process(PoolPayload{object: object, eventChannel: eventChannel, waitGroup: &wg})
	}

	go func() {
		// Wait for every worker and close pipelineChannel
		wg.Wait()
		close(eventChannel)
	}()

	return eventChannel
}
