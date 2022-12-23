package async

import (
	"fmt"
	"log"
	"net/http"
)

type RequestWrap struct {
	request  *http.Request
	response *http.Response
	client   *http.Client
	err      error
}

func requestAll(requests []RequestWrap) []*RequestWrap {

	channel := make(chan *RequestWrap)
	responses := make([]*RequestWrap, 0)
	log.Println(fmt.Sprintf("Making %d requests ", len(requests)))

	for _, request := range requests {

		go func(r RequestWrap) {
			response, err := r.client.Do(r.request)
			if err != nil {
				log.Println("failed http request", err)
				r.err = err
			}
			r.response = response

			channel <- &r
		}(request)

	}

loop:
	for {
		select {
		case r := <-channel:
			responses = append(responses, r)
			if len(responses) == len(requests) {
				break loop
			}
		}
	}

	return responses
}
