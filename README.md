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
