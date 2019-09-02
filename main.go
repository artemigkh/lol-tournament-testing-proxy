package main

import (
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

var sampleSummoners []string = []string {
	"AUbKDII4LaWXOsx0ETvw2ir7hyd3YA5l3-DlVioKySH1NVI",
	"Jf4hv6WngVeWsmBSD6AxRq9sap00FYmp1xTyx0s8Q9zDCDQ",
	"mjWGnT54UBjBz2yo0i7WHBRP90EdCCON-UO5eG9GfYM1iEc",
	"1lw_jzp5-0f2UrCgMitHuBUSfFIzOTuObSUqyIESXP0Gh4k",
	"TyxkWT3DcDKsNflmjyXGxs8E3bJek6BW_TGDsO1F1ClBAQY",
}

type LobbyEvent struct {
	Timestamp string `json:"timestamp"`
	EventType string `json:"eventType"`
	SummonerId string `json:"summonerId"`
}

type LobbyEvents struct {
	EventList []*LobbyEvent `json:"eventList"`
	i int
	j int
}

func NewLobbyEvents() *LobbyEvents {
	newLobbyEvents := &LobbyEvents{
		EventList: make([]*LobbyEvent, 0),
		i: -1,
		j: 0,
	}
	newLobbyEvents.EventList = append(newLobbyEvents.EventList, &LobbyEvent{
		Timestamp: "1234567890000",
		EventType: "PracticeGameCreatedEvent",
		SummonerId: sampleSummoners[0],
	})
	return newLobbyEvents
}

func (l *LobbyEvents) playerJoins(summonerId string) {
	l.EventList = append(l.EventList, &LobbyEvent{
		Timestamp: "1234567890000",
		EventType: "PlayerJoinedGameEvent",
		SummonerId: summonerId,
	})
}

func (l *LobbyEvents) startChampSelect() {
	l.EventList = append(l.EventList, &LobbyEvent{
		Timestamp: "1234567890000",
		EventType: "ChampSelectStartedEvent",
		SummonerId: "",
	})
}

func (l *LobbyEvents) startGame() {
	l.EventList = append(l.EventList, &LobbyEvent{
		Timestamp: "1234567890000",
		EventType: "GameAllocationStartedEvent",
		SummonerId: "",
	})
	l.EventList = append(l.EventList, &LobbyEvent{
		Timestamp: "1234567890000",
		EventType: "GameAllocatedToLsmEvent",
		SummonerId: "",
	})
}

func (l *LobbyEvents) simulateLobbyEvents() {
	if l.i == -1 {
		l.i = 0
		return
	}
	if l.i < len(sampleSummoners) {
		l.playerJoins(sampleSummoners[l.i])
		l.i++
	} else {
		if l.j == 0 {
			l.startChampSelect()
			l.j++
		} else if l.j == 1 {
			l.startGame()
		}
	}
}

func (l *LobbyEvents) randomPlayerLeaves() {

}





func copyHeader(source http.Header, dest *http.Header) {
	for n, v := range source {
		for _, vv := range v {
			dest.Add(n, vv)
		}
	}
}

func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}

func ReverseProxy() gin.HandlerFunc {
	target, _ := url.Parse("https://na1.api.riotgames.com")
	targetQuery := target.RawQuery

	return func(c *gin.Context) {
		director := func(req *http.Request) {
			req.URL.Scheme = target.Scheme
			req.URL.Host = target.Host
			req.Host = target.Host
			req.URL.Path = singleJoiningSlash(target.Path, req.URL.Path)
			if targetQuery == "" || req.URL.RawQuery == "" {
				req.URL.RawQuery = targetQuery + req.URL.RawQuery
			} else {
				req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
			}
			if _, ok := req.Header["User-Agent"]; !ok {
				// explicitly disable User-Agent so it's not set to default value
				req.Header.Set("User-Agent", "")
			}
			log.Print(req.URL)
		}
		proxy := &httputil.ReverseProxy{Director: director}
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}

func handleRegisterTournamentProvider() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, 878)
	}
}
func handleRegisterTournament() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, 6228)
	}
}

func handleGetTournamentCode() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, []string{"NA6228-A5B4B7F1C"})
	}
}

func handleGetLobbyEvents() gin.HandlerFunc {
	lobbyEvents := NewLobbyEvents()
	return func(ctx *gin.Context) {
		lobbyEvents.simulateLobbyEvents()
		ctx.JSON(http.StatusOK, lobbyEvents)
	}
}

func main() {
	app := gin.Default()

	app.POST("/lol/tournament-stub/v4/providers", handleRegisterTournamentProvider())
	app.POST("/lol/tournament-stub/v4/tournaments", handleRegisterTournament())
	app.POST("/lol/tournament-stub/v4/codes", handleGetTournamentCode())
	app.GET("/lol/tournament-stub/v4/lobby-events/by-code/:tournamentCode", handleGetLobbyEvents())
	app.NoRoute(ReverseProxy())

	if err := app.Run("localhost:8865"); err != nil {
		log.Fatal(err)
	}
}
