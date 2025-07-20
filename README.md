Usage
=====
You need golang and postgres installed to run this RSS feed browser.

Once you have the repository build and install using:
```sh
go build
go install
```
You will also need to create a configuration file in ~/.gatorconfig.json which must have the following:
```json
{
    "db_url": "postgres://<username>:<password>@localhost:5432/gator?sslmode=disable",
    "current_user_name": ""
}
```

Commands
========
1. Register a user 
    ```sh
    gator register <username>
    ```
2. Login 
    ```sh
    gator login <username>
    ```
3. Reset to factory settings
    ```sh
    gator reset
    ```
4. Add a feed to gator
    ```sh
    gator addfeed <feedurl>
    ```
5. Start scraping with an interval. Keep your duration long enough to not dos a server. 
    ```sh
    gator agg 10m
    ```
6. Follow a feed
    ```sh
    gator follow url
    ```
7. List all feeds the user is following
    ```sh
    gator following
    ```
8. Show posts for user
    ```sh
    gator browse
    ```

DB migrations
=============
1. Up migration
```
goose -dir ./sql/schema postgres "postgres://gauravgautam:@localhost:5432/gator?sslmode=disable" up
```

2. Down migration
```
goose -dir ./sql/schema postgres "postgres://gauravgautam:@localhost:5432/gator?sslmode=disable" down
```

3. Migration status
```
goose -dir ./sql/schema postgres "postgres://gauravgautam:@localhost:5432/gator?sslmode=disable" status
```
