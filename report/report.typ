#import "@preview/lilaq:0.4.0" as lq

#align(center, text(18pt)[*Assignment 3 di Programmazione Concorrente e Distribuita 2025/2026*])
#align(center, text(
  12pt,
)[Mattia Ronchi, matr. 0001236997 \ Samorì Andrea matr. 0001235969 \ Andrea Monaco matr. 0001225150])

= Analisi del problema

L'assignment ha l'obiettivo di realizzare il gioco `Odds-and-Evens`, un torneo concorrente di pari o dispari tra più giocatori. Nella nostra implementazione, ogni round del torneo consiste in più partite concorrenti. In ogni partita, 2 giocatori scelgono un numero casuale tra 0 e 9 e il giudice (`match`) decreterà il vincitore della partita. Questo proseguirà nel round successivo, il perdente viene eliminato.

= Aspetti rilevanti per la concorrenza

Dal punto di vista della concorrenza, abbiamo individuato alcuni aspetti rilevanti:
- la sincronizzazione dei round: i giocatori non possono avanzare al round successivo finchè tutti le partite del round non sono finite, decretando i vincitori che procederanno
- le partite di uno stesso round sono giocate (e valutate) in parallelo

= Design della soluzione

La soluzione da noi implementata crea diverse `goroutine`:
- una per il `roundManager`, che si occupa di gestire l'avanzamento dei round
- una per ogni partita (`match`)
- una per ogni giocatore (`play`)

Per prima cosa, viene configurato il `roundManager`. Dopo, viene creato l'albero del torneo. Nelle foglie, sono presenti tutti i giocatori, e nella radice il vincitore. Tramite questo albero, vengono creati i vari `match`. Infine, vengono istanziati tutti i giocatori che, tramite `play`, iniziano a giocare al torneo.

== RoundManager

Il `roundManager` è adibito all'avanzamento dei round. Per farlo, aspetta che tutti i giocatori partecipanti al round corrente abbiano completato la partita. Infine invia ai vincitori il permesso di passare al round successivo.

== Match

Il match rappresenta una singola partita svolta da due giocatori. I due giocatori inviano al match il numero scelto. Il match decreta il vincitore e invia ai partecipanti il risultato.

== Giocatori

Un giocatore tramite la funzione `play` gioca al torneo. Per ogni round, il giocatore genera un numero e lo invia al `match`, mettendosi in attesa del risultato. Una volta ricevuto, esistono due opzioni:
- in caso di vittoria il giocatore aspetta che tutti gli altri match del round finiscano per poi avanzare al round successivo
- in caso di sconfitta esce dal torneo

== Comunicazione

Ogni giocatore comunica tramite la struttura `played`.
I campi `id`, `send`, `reply` vengono utilizzati per comunicare con il match. `send` rappresenta il canale in cui viene passato il numero scelto dal giocatore mentre `reply` è il canale dove il giocatore si aspetta il verdetto del match. `notify` e `barrier` sono utilizzati per comunicare con il `roundManager`.

```go
type played struct {
	id      int
	send    chan int
	reply   chan bool
	notify  []chan struct{}
	barrier []chan struct{}
}
```
