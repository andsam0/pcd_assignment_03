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

Il `roundManager` è adibito all'avanzamento dei round. Per farlo, aspetta che tutti i giocatori partecipanti al round corrente abbiano completato la partita. Poi, invia ai vincitori in

== Match

Il match rappresenta una singola partita, svolta da due giocatori.

== Giocatori

Ogni giocatore è rappresentato da una `goroutine` che esegue la funzione `tournament`. Il ciclo della `goroutine` termina in caso di sconfitta e prosegue in caso di vittoria.

```go
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
```

== Canali per i vari round
Abbiamo utilizzato un array di canali dove l'indice rappresenta il round coreente, in modo che i giocatori di un certo round comunichino con il giudice di quel round.

== Comunicazione
Ogni giocatore comunica al giudice, oltre al numero scelto e al proprio id, anche un canale di risposta (`reply`). In questo, il giudice riceve 2 messaggi, determina il vincitore e utilizza i rispettivi canali per comunicare ai giocatori l'esito.
```go
type msg struct {
	number int
	reply  chan result
	id     int
}
```

== Sincronizzazione dei round
Per garantire che il passaggio al round successivo sia permesso solo alla fine di tutte le partite del round precedente, abbiamo implementato un array di barriere (una per round). I vincitori rimangono in attesa (`<-barriers[i]`) e si bloccano nello stesso momento quando il giudice ha finito di giudicare tutte le partite del round corrente (`close(barrier)`).

== Terminazione pulita del programma
Abbiamo creato un canale `done`, dove il vincitore del torneo manderà un messaggio quando riceverà la notizia della vittoria. Il main rimarrà in esecuzione fino alla ricezione del messaggio da parte del vincitore, che rappresenta la fine del torneo.
