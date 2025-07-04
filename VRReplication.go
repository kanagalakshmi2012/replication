package main
import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)
type VRState struct {
	ClientID  int
	StateID   int
	Timestamp time.Time
	Data      string
}
type Ack struct {
	ClientID int
	StateID  int
	Success  bool
}
func generateState(clientID, stateID int) VRState {
	return VRState{
		ClientID:  clientID,
		StateID:   stateID,
		Timestamp: time.Now(),
		Data:      fmt.Sprintf("Client_%d_State_%d", clientID, stateID),
	}
}
func replicateState(state VRState, ch chan VRState, lossChance float64, latency time.Duration) {
	time.Sleep(latency)
	if rand.Float64() >= lossChance {
		ch <- state
	}
}
func server(stateCh chan VRState, ackCh chan Ack, wg *sync.WaitGroup) {
	defer wg.Done()
	for state := range stateCh {
		processState(state)
		ack := Ack{ClientID: state.ClientID, StateID: state.StateID, Success: true}
		ackCh <- ack
	}
}
func processState(state VRState) {
	fmt.Printf("Server received: Client=%d State=%d Time=%s\n",
		state.ClientID, state.StateID, state.Timestamp.Format("15:04:05.000"))
}
func client(id int, stateCh chan VRState, ackCh chan Ack, totalStates int, delay, latency time.Duration, lossChance float64, wg *sync.WaitGroup) {
	defer wg.Done()
	for i := 0; i < totalStates; i++ {
		state := generateState(id, i)
		go replicateState(state, stateCh, lossChance, latency)
		time.Sleep(delay)
	}
}
func ackHandler(ackCh chan Ack, totalClients, totalStates int, done chan bool) {
	acks := 0
	expected := totalClients * totalStates
	for ack := range ackCh {
		if ack.Success {
			fmt.Printf("Ack: Client=%d State=%d\n", ack.ClientID, ack.StateID)
			acks++
		}
		if acks >= expected {
			break
		}
	}
	done <- true
}
func main() {
	rand.Seed(time.Now().UnixNano())
	stateCh := make(chan VRState, 100)
	ackCh := make(chan Ack, 100)
	done := make(chan bool)
	totalClients := 2
	statesPerClient := 10
	sendDelay := 20 * time.Millisecond
	netLatency := 5 * time.Millisecond
	packetLoss := 0.1
	var wg sync.WaitGroup
	wg.Add(1)
	go server(stateCh, ackCh, &wg)
	for i := 0; i < totalClients; i++ {
		wg.Add(1)
		go client(i, stateCh, ackCh, statesPerClient, sendDelay, netLatency, packetLoss, &wg)
	}
	go ackHandler(ackCh, totalClients, statesPerClient, done)
	wg.Wait()
	close(stateCh)
	close(ackCh)
	<-done
	fmt.Println("Replication complete.")
}
