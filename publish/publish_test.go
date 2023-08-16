package publish_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/xpetit/upfluence-social/publish"
)

func TestPublisher(t *testing.T) {
	messages := make(chan int)
	publisher := publish.New(messages)
	sub, _ := publisher.Subscribe()

	const limit = 1000
	go func() {
		for i := 0; i < limit; i++ {
			messages <- i
		}
	}()

	for want := 0; want < limit; want++ {
		if got := <-sub; got != want {
			t.Fatalf("got:%d, want:%d", got, want)
		}
	}
}

func BenchmarkPublisher(b *testing.B) {
	const limit = 1000

	for nSubscribers := 1; nSubscribers <= 1024; nSubscribers *= 2 {
		b.Run(fmt.Sprint(nSubscribers, " subscribers"), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				b.StopTimer()

				messages := make(chan int, 100)
				publisher := publish.New(messages)

				// spawn subscribers
				var wg sync.WaitGroup
				for i := 0; i < nSubscribers; i++ {
					sub, _ := publisher.Subscribe()
					wg.Add(1)
					go func() {
						for want := 0; want < limit; want++ {
							if got := <-sub; got != want {
								b.Errorf("got:%d, want:%d", got, want)
							}
						}
						wg.Done()
					}()
				}

				b.StartTimer()

				// publish all messages
				for i := 0; i < limit; i++ {
					messages <- i
				}

				// wait for subscribers to finish reading all messages
				wg.Wait()
			}
		})
	}
}
