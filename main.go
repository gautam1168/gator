package main

import (
	"context"
	"database/sql"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"os"

	// "os"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"

	"gator/internal/config"
	"gator/internal/database"
)

type state struct {
	db  *database.Queries
	cfg *config.Config
	ctx context.Context
}

type command struct {
	name string
	args []string
}

type commands struct {
	registry map[string]func(*state, command) error
}

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	feed := &RSSFeed{}
	request, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return feed, err
	}

	request.Header.Set("User-Agent", "gator")

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return feed, err
	}
	defer response.Body.Close()

	responseBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return feed, err
	}

	xml.Unmarshal(responseBytes, feed)

	feed.Channel.Title = html.UnescapeString(feed.Channel.Title)
	feed.Channel.Description = html.UnescapeString(feed.Channel.Description)

	for i := 0; i < len(feed.Channel.Item); i++ {
		item := feed.Channel.Item[i]
		item.Title = html.UnescapeString(item.Title)
		item.Description = html.UnescapeString(item.Description)
	}

	return feed, nil
}

func handlerAggregate(s *state, cmd command) error {
	feed, err := fetchFeed(s.ctx, "https://www.wagslane.dev/index.xml")
	if err != nil {
		log.Fatal("could not fetch feed")
	}

	fmt.Printf("%v", *feed)
	return nil
}

func (c *commands) run(s *state, cmd command) error {
	if handler, ok := c.registry[cmd.name]; !ok {
		return fmt.Errorf("invalid command")
	} else {
		return handler(s, cmd)
	}
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.registry[name] = f
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) < 1 {
		log.Fatal("a username is required")
	}

	if user, err := s.db.GetUserByName(s.ctx, cmd.args[0]); err != nil {
		log.Fatal("could not retrieve user from database")
	} else {
		if user.Name != cmd.args[0] {
			log.Fatal("user is not registered")
		}
	}

	return s.cfg.SetUser(cmd.args[0])
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) < 1 {
		log.Fatal("username is required")
	}

	user, err := s.db.CreateUser(s.ctx, database.CreateUserParams{
		ID: uuid.New(),
		CreatedAt: sql.NullTime{
			Valid: true,
			Time:  time.Now(),
		},
		UpdatedAt: sql.NullTime{
			Valid: true,
			Time:  time.Now(),
		},
		Name: cmd.args[0],
	})

	if err != nil {
		log.Fatalf("could not create user %v", err)
	}
	fmt.Printf("%v\n", user)

	s.cfg.SetUser(cmd.args[0])

	return err
}

func handlerReset(s *state, cmd command) error {
	err := s.db.ResetUsers(s.ctx)
	if err != nil {
		log.Fatal("could not clear users")
	}
	return nil
}

func handlerGetUsers(s *state, cmd command) error {
	users, err := s.db.GetUsers(s.ctx)
	if err != nil {
		log.Fatal("could not get users")
	}

	currentUserName := s.cfg.CurrentUserName
	for i := 0; i < len(users); i++ {
		user := users[i]
		if user.Name == currentUserName {
			fmt.Printf("%s (current)\n", user.Name)
		} else {
			fmt.Println(user.Name)
		}
	}
	return nil
}

func main() {
	cfg, err := config.Read()
	if err != nil {
		log.Fatal("Could not read config file")
	}

	db, err := sql.Open("postgres", cfg.DbUrl)
	if err != nil {
		log.Fatal("Could not open connection to database")
	}

	dbQueries := database.New(db)

	appState := state{
		cfg: &cfg,
		ctx: context.Background(),
		db:  dbQueries,
	}

	cmds := commands{
		registry: make(map[string]func(*state, command) error),
	}

	cmds.register("login", handlerLogin)
	cmds.register("register", handlerRegister)
	cmds.register("reset", handlerReset)
	cmds.register("users", handlerGetUsers)
	cmds.register("agg", handlerAggregate)

	// fmt.Printf("current: %v", cfg)

	args := os.Args

	if len(args) < 2 {
		log.Fatal("not enough arguments were provided")
	}

	cmd := command{
		name: args[1],
	}

	if len(args) > 2 {
		cmd.args = args[2:]
	}

	if err := cmds.run(&appState, cmd); err != nil {
		fmt.Printf("%v", err)
	}
}
