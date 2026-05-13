package main

import (
	"fmt"
	"math/rand/v2"
	"os"
	"strconv"
	// "time"
)

type msg struct {
	number int
	reply  chan result
	id     int
}

type result struct {
	won      bool
	opponent int // player id
}

func player(id int, ch chan msg) result {
	number := rand.IntN(10)
	reply := make(chan result)
	ch <- msg{number, reply, id}
	result := <-reply
	return result
}

func tournament(id int, channels []chan msg, barriers []chan struct{}) {
	for i, ch := range channels {
		result := player(id, ch)
		if !result.won {
			return
		}
		<-barriers[i]
		fmt.Printf("Player %d wins against player %d in round %d\n", id, result.opponent, i)
	}
	fmt.Printf("\n Player %d wins the tournament!\n", id)
}

func judge(round, games int, ch chan msg, barrier chan struct{}) {
	// fmt.Printf("\n--- Round %d: %d concurrent game(s) ---\n", round, games)
	for i := 0; i < games; i++ {
		m1 := <-ch
		m2 := <-ch
		sum := m1.number + m2.number
		if sum%2 == 0 {
			m1.reply <- result{true, m2.id}
			m2.reply <- result{false, m1.id}
		} else {
			m1.reply <- result{false, m2.id}
			m2.reply <- result{true, m1.id}
		}
	}
	close(barrier)
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

	channels := make([]chan msg, m)
	for i := range channels {
		channels[i] = make(chan msg)
	}
	barriers := make([]chan struct{}, m)
	for i := range barriers {
		barriers[i] = make(chan struct{})
	}

	for i := range N {
		go tournament(i+1, channels, barriers)
	}

	for r := range m {
		judge(r+1, N>>(r+1), channels[r], barriers[r])
	}

	// todo: trovare un modo per non terminare prima di stampare il vincitore del torneo
	for{}
}
