package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
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

	fmt.Printf("current: %v", cfg)

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
