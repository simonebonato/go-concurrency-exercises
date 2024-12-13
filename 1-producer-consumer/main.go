//////////////////////////////////////////////////////////////////////
//
// Given is a producer-consumer scenario, where a producer reads in
// tweets from a mockstream and a consumer is processing the
// data. Your task is to change the code so that the producer as well
// as the consumer can run concurrently
//

package main

import (
	"fmt"
	"prod-cons/mockstream"
	"sync"
	"time"
)

func producer(stream mockstream.Stream, tweets chan *mockstream.Tweet, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		tweet, err := stream.Next()
		if err == mockstream.ErrEOF {
			close(tweets)
			return
		}
		tweets <- tweet
	}
}

func consumer(tweets chan *mockstream.Tweet, wg *sync.WaitGroup) {
	defer wg.Done()
	for t := range tweets {
		if t.IsTalkingAboutGo() {
			fmt.Println(t.Username, "\ttweets about golang")
		} else {
			fmt.Println(t.Username, "\tdoes not tweet about golang")
		}
	}
}

func main() {
	start := time.Now()
	stream := mockstream.GetMockStream()

	// I probably have to make 2 go routines
	// one for the producer and one for the consumer
	// and the producer should get his stuff through a channel,
	// that is then processed in parallel by the consumer
	tweets := make(chan *mockstream.Tweet)

	// one thing that I did not think about was to use a wait group
	// it is decremented by one when:
	// 		- the producer does not receive more stuff
	// 		- the consumer stops receiving stuff from the tweets channel, actually when it is CLOSED
	var wg sync.WaitGroup

	wg.Add(2)

	// Producer
	go producer(stream, tweets, &wg)

	// Consumer
	go consumer(tweets, &wg)

	wg.Wait()
	fmt.Printf("Process took %s\n", time.Since(start))
}
