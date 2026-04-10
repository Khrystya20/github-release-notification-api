package scheduler

import (
	"log"
	"time"
)

type Scanner interface {
	ScanOnce()
}

func Start(scanner Scanner, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		log.Printf("scanner scheduler started, interval=%s", interval)

		for {
			<-ticker.C
			scanner.ScanOnce()
		}
	}()
}
