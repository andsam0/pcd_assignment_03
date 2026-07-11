package main

import (
	"fmt"
	"math/rand/v2"
	"os"
	"strconv"
)

type played struct {
	id      int
	send    chan int
	reply   chan bool
	notify  []chan struct{}
	barrier []chan struct{}
}

func play(player played, rounds int) {
	for i := range rounds {
		player.send <- rand.IntN(10)
		victory := <-player.reply
		player.notify[i] <- struct{}{}
		if !victory {
			return
		}
		<-player.barrier[i]
	}
}

func roundManager(round int, recv []chan struct{}, send []chan struct{}) {
	for i := range round {
		fmt.Printf("Round Actor says NEW ROUND %d\n", i+1)
		for range 1 << (round - i) {
			<-recv[i]
		}
		for range 1 << (round - i - 1) {
			send[i] <- struct{}{}
		}
	}
}

func match(round int, left <-chan played, right <-chan played, out chan<- played) {
	a := <-left
	b := <-right

	n1 := <-a.send
	n2 := <-b.send
	winner, loser, parity := a, b, "even"
	if (n1+n2)%2 != 0 {
		winner, loser, parity = b, a, "odd"
	}
	fmt.Printf("round %d: player %d beats player %d  (%d+%d=%d, %s)\n",
		round, winner.id, loser.id, n1, n2, n1+n2, parity)
	winner.reply <- true
	loser.reply <- false
	out <- winner
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

	leaves := make([]chan played, N)
	notify := make([]chan struct{}, m)
	barrier := make([]chan struct{}, m)

	for i := range m {
		notify[i] = make(chan struct{})
		barrier[i] = make(chan struct{}, 1)
	}

	go roundManager(m, notify, barrier)

	for i := range leaves {
		leaves[i] = make(chan played, 1)
	}

	layer := leaves
	for round := 1; len(layer) > 1; round++ {
		next := make([]chan played, len(layer)/2)
		for i := range next {
			next[i] = make(chan played, 1)
			go match(round, layer[2*i], layer[2*i+1], next[i])
		}
		layer = next
	}

	for i := range leaves {
		player := played{i, make(chan int), make(chan bool), notify, barrier}
		leaves[i] <- player
		go play(player, m)
	}

	player := <-layer[0]
	fmt.Printf("\nPlayer %d wins the tournament!\n", player.id)
}
