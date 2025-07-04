package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type Proposal struct {
	ID        int
	Timestamp time.Time
	Data      string
}

type Ack struct {
	FollowerID int
	ProposalID int
	Success    bool
}

func generateProposal(id int) Proposal {
	return Proposal{
		ID:        id,
		Timestamp: time.Now(),
		Data:      fmt.Sprintf("Update_%d", id),
	}
}

func leader(proposals int, followers int, proposalChs []chan Proposal, ackCh chan Ack, done chan bool) {
	for i := 0; i < proposals; i++ {
		p := generateProposal(i)
		for _, ch := range proposalChs {
			ch <- p
		}
		acks := 0
		for acks < followers {
			ack := <-ackCh
			if ack.ProposalID == p.ID && ack.Success {
				acks++
			}
		}
		fmt.Printf("Leader: Proposal %d committed with %d acks\n", p.ID, acks)
	}
	for _, ch := range proposalChs {
		close(ch)
	}
	done <- true
}

func follower(id int, proposalCh chan Proposal, ackCh chan Ack, wg *sync.WaitGroup, latency time.Duration, lossChance float64) {
	defer wg.Done()
	for p := range proposalCh {
		time.Sleep(latency)
		if rand.Float64() > lossChance {
			ackCh <- Ack{FollowerID: id, ProposalID: p.ID, Success: true}
			fmt.Printf("Follower %d: Acked proposal %d\n", id, p.ID)
		} else {
			fmt.Printf("Follower %d: Dropped proposal %d\n", id, p.ID)
		}
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())

	totalFollowers := 3
	totalProposals := 5
	latency := 10 * time.Millisecond
	packetLoss := 0.1

	proposalChs := make([]chan Proposal, totalFollowers)
	for i := range proposalChs {
		proposalChs[i] = make(chan Proposal, 100)
	}
	ackCh := make(chan Ack, 100)
	done := make(chan bool)

	var wg sync.WaitGroup
	for i := 0; i < totalFollowers; i++ {
		wg.Add(1)
		go follower(i, proposalChs[i], ackCh, &wg, latency, packetLoss)
	}

	start := time.Now()
	go leader(totalProposals, totalFollowers, proposalChs, ackCh, done)

	<-done
	wg.Wait()
	elapsed := time.Since(start)

	fmt.Printf("ZAB replication of %d proposals completed in %s\n", totalProposals, elapsed)
}
