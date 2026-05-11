package main

import (
	"fmt"
	"math/rand/v2"
	"os"
	"strconv"
)

// msg is sent by a player to the judge: their chosen number and a reply channel.
type msg struct {
	number int
	reply  chan bool // true → winner (advance), false → eliminated
	id     int
}

// player participates in one round on ch, returning whether it won.
func player(id int, ch chan msg) bool {
	number := rand.IntN(10)
	reply := make(chan bool)
	ch <- msg{number, reply, id}
	won := <-reply
	if won {
		fmt.Printf("  [player %2d] drew %d → advances\n", id, number)
	} else {
		fmt.Printf("  [player %2d] drew %d → eliminated\n", id, number)
	}
	return won
}

// tournament runs one player through all rounds in sequence.
// The player walks the channel slice, stopping as soon as it loses.
func tournament(id int, channels []chan msg, done chan struct{}) {
	defer close(done)
	for _, ch := range channels {
		if !player(id, ch) {
			return
		}
	}
	fmt.Printf("\n Player %d wins the tournament!\n", id)
}

// judge manages one round: reads `games` pairs and sends verdicts.
func judge(round, games int, ch chan msg) {
	fmt.Printf("\n--- Round %d: %d concurrent game(s) ---\n", round, games)
	for i := 0; i < games; i++ {
		m1 := <-ch
		m2 := <-ch
		sum := m1.number + m2.number
		if sum%2 == 0 {
			m1.reply <- true
			m2.reply <- false
		} else {
			m1.reply <- false
			m2.reply <- true
		}
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: odds_and_evens <m>   (N = 2^m players)")
		os.Exit(1)
	}
	m, err := strconv.Atoi(os.Args[1])
	if err != nil || m < 1 {
		fmt.Fprintln(os.Stderr, "m must be a positive integer")
		os.Exit(1)
	}

	N := 1 << m
	fmt.Printf("Odds-and-Evens Tournament: %d players, %d rounds\n\n", N, m)

	// One unbuffered channel per round.
	channels := make([]chan msg, m)
	for i := range channels {
		channels[i] = make(chan msg)
	}

	// Launch N player goroutines. Each walks all channels, stopping on a loss.
	dones := make([]chan struct{}, N)
	for i := range N {
		dones[i] = make(chan struct{})
		go tournament(i+1, channels, dones[i])
	}

	// Run m rounds sequentially on the main goroutine.
	// Round r has N / 2^(r+1) games.
	for r := range m {
		judge(r+1, N>>(r+1), channels[r])
	}

	// Wait for every player goroutine to exit.
	for _, d := range dones {
		<-d
	}
}
