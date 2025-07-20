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
	if len(cmd.args) < 1 {
		log.Fatal("duration must be provided")
	}

	interval, err := time.ParseDuration(cmd.args[0])
	if err != nil {
		log.Fatal("invalid duration provided")
	}

	fmt.Printf("Collecting feeds every %v\n", interval)
	ticker := time.NewTicker(interval)
	for ; ; <-ticker.C {
		fmt.Printf("scraping next\n")
		err := scrapeFeeds(s)
		if err != nil {
			fmt.Printf("Could not scrape %v\n", err)
		}
	}
}

func handlerAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.args) < 2 {
		log.Fatalf("name and url must be provided")
	}

	name := cmd.args[0]
	url := cmd.args[1]

	feed, err := s.db.AddFeed(s.ctx, database.AddFeedParams{
		ID: uuid.New(),
		CreatedAt: sql.NullTime{
			Valid: true,
			Time:  time.Now(),
		},
		UpdatedAt: sql.NullTime{
			Valid: true,
			Time:  time.Now(),
		},
		Name: sql.NullString{
			Valid:  true,
			String: name,
		},
		Url: sql.NullString{
			Valid:  true,
			String: url,
		},
		UserID: user.ID,
	})

	fmt.Printf("id:%v\nname: %v\nurl: %v\nuser_id:%v\n",
		feed.ID, feed.Name, feed.Url, feed.UserID)

	_, err = s.db.CreateFeedFollow(s.ctx, database.CreateFeedFollowParams{
		ID: uuid.New(),
		UserID: uuid.NullUUID{
			Valid: true,
			UUID:  user.ID,
		},
		FeedID: uuid.NullUUID{
			Valid: true,
			UUID:  feed.ID,
		},
		CreatedAt: sql.NullTime{
			Valid: true,
			Time:  time.Now(),
		},
		UpdatedAt: sql.NullTime{
			Valid: true,
			Time:  time.Now(),
		},
	})

	return err
}

func handlerFeeds(s *state, cmd command) error {
	feeds, err := s.db.ListFeeds(s.ctx)
	if err != nil {
		return err
	}

	for i := range feeds {
		feed := feeds[i]
		fmt.Printf("%s, %s, %v\n", feed.Name.String, feed.Url.String, feed.UserName)
	}
	return nil
}

func handlerFollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) < 1 {
		log.Fatal("feed url must be provided")
	}

	feed, err := s.db.GetFeedByUrl(s.ctx, sql.NullString{
		Valid:  true,
		String: cmd.args[0],
	})

	if err != nil {
		log.Fatal("cannot find the feed")
	}

	feedFollow, err := s.db.CreateFeedFollow(s.ctx, database.CreateFeedFollowParams{
		ID: uuid.New(),
		UserID: uuid.NullUUID{
			Valid: true,
			UUID:  user.ID,
		},
		FeedID: uuid.NullUUID{
			Valid: true,
			UUID:  feed.ID,
		},
		CreatedAt: sql.NullTime{
			Valid: true,
			Time:  time.Now(),
		},
		UpdatedAt: sql.NullTime{
			Valid: true,
			Time:  time.Now(),
		},
	})

	fmt.Printf("%v", feedFollow)

	return err
}

func handlerUnfollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) < 1 {
		log.Fatal("user and url must be provided")
	}

	err := s.db.Unfollow(s.ctx, database.UnfollowParams{
		UserID: uuid.NullUUID{
			Valid: true,
			UUID:  user.ID,
		},
		Url: sql.NullString{
			Valid:  true,
			String: cmd.args[0],
		},
	})
	return err
}

func handlerFollowing(s *state, cmd command) error {
	follows, err := s.db.GetFeedFollowsForUser(s.ctx, s.cfg.CurrentUserName)
	if err != nil {
		return err
	}

	for i := range follows {
		follow := follows[i]
		fmt.Println(follow.FeedName)
	}

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
	for i := range users {
		user := users[i]
		if user.Name == currentUserName {
			fmt.Printf("%s (current)\n", user.Name)
		} else {
			fmt.Println(user.Name)
		}
	}
	return nil
}

func middlewareLoggedIn(
	handler func(s *state, cmd command, user database.User) error,
) func(*state, command) error {
	return func(s *state, cmd command) error {
		user, err := s.db.GetUserByName(s.ctx, s.cfg.CurrentUserName)
		if err != nil {
			return err
		}
		return handler(s, cmd, user)
	}
}

func scrapeFeeds(s *state) error {
	feed, err := s.db.GetNextFeedToFetch(s.ctx)
	if err != nil {
		return err
	}

	if err := s.db.MarkFeedFetched(s.ctx, feed.ID); err != nil {
		return err
	}

	if feed.Url.Valid {
		feedContent, err := fetchFeed(s.ctx, feed.Url.String)
		if err != nil {
			return err
		}

		fmt.Printf("Feed: %s\n", feedContent.Channel.Title)
		fmt.Printf("Description: %s\n", feedContent.Channel.Description)
		fmt.Printf("------------------------------------\n")

		for i := range feedContent.Channel.Item {
			feedItem := feedContent.Channel.Item[i]
			fmt.Printf("%v. %s\n", i, feedItem.Title)
			fmt.Printf("   %s\n\n", feedItem.Description)
		}
		return nil
	} else {
		return fmt.Errorf("cannot fetch feed that has no url")
	}
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
	cmds.register("feeds", handlerFeeds)
	cmds.register("addfeed", middlewareLoggedIn(handlerAddFeed))
	cmds.register("follow", middlewareLoggedIn(handlerFollow))
	cmds.register("following", handlerFollowing)
	cmds.register("unfollow", middlewareLoggedIn(handlerUnfollow))

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
