# pgxv5trace

`pgxv5trace` adds `apm-go` support for [jackc/pgx (v5+)][pgx].

## pgxpool: statsd

The pgxpool statsd integration adds connection pool monitoring under
your service’s namespace, e.g.,

```
<hopper_service_name>.pgxpool.<database_name>.<metric_name>
```

```go
databaseURL := "postgres://localhost/example"
databaseName := "example"
cfg, err := pgxpool.ParseConfig(databaseURL)
if err != nil {
  return nil, err
}
pool, err := pgxv5trace.Connect(ctx, databaseName, cfg)
if err != nil {
  return nil, err
}
return pool, nil
```

## pgxpool: logger

It’s rare that you’d need this, but pgxv5trace also exports a logger
compatible with `pgx`:

```go
databaseURL := "postgres://localhost/example"
logger := pgxv5trace.NewLogger()
cfg, err := pgxpool.ParseConfig(databaseURL)
if err != nil {
  return nil, err
}
cfg.ConnConfig.Logger = logger
```

[pgx]: https://github.com/jackc/pgx
