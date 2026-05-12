# Redis Component

A Redis wrapper that implements `sctx.Component` for use with the service context. Built on top of `github.com/redis/go-redis/v9`.

## Setup

```go
redisComp := redisc.NewRedis("redis")

sc := sctx.NewServiceContext(
    sctx.WithName("my-service"),
    sctx.WithComponent(redisComp),
)

if err := sc.Load(); err != nil {
    log.Fatal(err)
}
```

Retrieve the component anywhere in your application:

```go
rc := sc.MustGet("redis").(*redisc.redisComponent)
```

## Configuration

Configuration is read from flags or environment variables.

| Flag                | Env                 | Default     | Description               |
|---------------------|---------------------|-------------|---------------------------|
| `redis-host`        | `REDIS_HOST`        | `localhost` | Redis server host         |
| `redis-port`        | `REDIS_PORT`        | `6379`      | Redis server port         |
| `redis-password`    | `REDIS_PASSWORD`    | (empty)     | Redis password            |
| `redis-db`          | `REDIS_DB`          | `0`         | Database index            |
| `redis-pool-size`   | `REDIS_POOL_SIZE`   | `10`        | Connection pool size      |

## Methods

### String

**`Get(ctx, key) (string, bool, error)`**  
Returns the value of a key. The second return value is `false` if the key does not exist, with no error.

```go
val, found, err := rc.Get(ctx, "session:abc")
```

**`Set(ctx, key, value, ttl) error`**  
Stores a string value. Pass `0` as `ttl` for no expiration.

```go
rc.Set(ctx, "session:abc", "user-123", 30*time.Minute)
```

**`SetNX(ctx, key, value, ttl) (bool, error)`**  
Stores a value only when the key does not already exist. Returns `true` if the key was written.

```go
ok, err := rc.SetNX(ctx, "lock:job", "1", 10*time.Second)
```

**`Del(ctx, keys...) (int64, error)`**  
Deletes one or more keys. Returns the number of keys actually deleted.

```go
rc.Del(ctx, "session:abc", "session:xyz")
```

**`Exists(ctx, keys...) (int64, error)`**  
Returns how many of the given keys exist.

```go
n, err := rc.Exists(ctx, "session:abc")
```

**`Incr(ctx, key) (int64, error)`**  
Atomically increments an integer key by 1 and returns the new value.

```go
hits, err := rc.Incr(ctx, "hits:homepage")
```

**`IncrBy(ctx, key, n) (int64, error)`**  
Atomically increments an integer key by `n` and returns the new value.

```go
total, err := rc.IncrBy(ctx, "score:user-1", 10)
```

---

### JSON

Convenience wrappers that marshal/unmarshal values automatically.

**`SetJSON(ctx, key, src, ttl) error`**  
Marshals `src` to JSON and stores it. Pass `0` for no expiration.

```go
type Product struct { ID string; Name string }
rc.SetJSON(ctx, "product:1", Product{ID: "1", Name: "Widget"}, time.Hour)
```

**`GetJSON(ctx, key, dst) (bool, error)`**  
Retrieves a key and unmarshals the JSON value into `dst`. Returns `false` (no error) if the key does not exist.

```go
var p Product
found, err := rc.GetJSON(ctx, "product:1", &p)
```

---

### Expiry

**`Expire(ctx, key, ttl) (bool, error)`**  
Sets a timeout on an existing key. Returns `false` if the key does not exist.

```go
rc.Expire(ctx, "session:abc", 15*time.Minute)
```

**`TTL(ctx, key) (time.Duration, error)`**  
Returns the remaining time-to-live of a key. Returns `-1` if the key has no expiry, `-2` if the key does not exist.

```go
remaining, err := rc.TTL(ctx, "session:abc")
```

---

### Hash

**`HSet(ctx, key, field, value, ...) error`**  
Sets one or more field-value pairs on a hash. Pass pairs as alternating arguments.

```go
rc.HSet(ctx, "user:1", "name", "Alice", "age", "30")
```

**`HGet(ctx, key, field) (string, bool, error)`**  
Returns the value of a hash field. The second return is `false` if the field does not exist.

```go
name, found, err := rc.HGet(ctx, "user:1", "name")
```

**`HGetAll(ctx, key) (map[string]string, error)`**  
Returns all field-value pairs of a hash.

```go
fields, err := rc.HGetAll(ctx, "user:1")
```

**`HDel(ctx, key, fields...) (int64, error)`**  
Deletes one or more fields from a hash. Returns the number of fields removed.

```go
rc.HDel(ctx, "user:1", "age")
```

---

### List

**`LPush(ctx, key, values...) (int64, error)`**  
Inserts values at the head (left) of a list. Returns the new list length.

```go
rc.LPush(ctx, "queue:emails", "email-1", "email-2")
```

**`RPush(ctx, key, values...) (int64, error)`**  
Inserts values at the tail (right) of a list. Returns the new list length.

```go
rc.RPush(ctx, "queue:emails", "email-3")
```

**`LRange(ctx, key, start, stop) ([]string, error)`**  
Returns elements of a list between `start` and `stop` (inclusive). Use `0` and `-1` for all elements.

```go
items, err := rc.LRange(ctx, "queue:emails", 0, -1)
```

**`LLen(ctx, key) (int64, error)`**  
Returns the number of elements in a list.

```go
length, err := rc.LLen(ctx, "queue:emails")
```

---

### Set

**`SAdd(ctx, key, members...) (int64, error)`**  
Adds one or more members to a set. Returns the number of new members added (duplicates are ignored).

```go
rc.SAdd(ctx, "online:users", "user-1", "user-2")
```

**`SMembers(ctx, key) ([]string, error)`**  
Returns all members of a set.

```go
members, err := rc.SMembers(ctx, "online:users")
```

**`SIsMember(ctx, key, member) (bool, error)`**  
Reports whether a value is a member of a set.

```go
online, err := rc.SIsMember(ctx, "online:users", "user-1")
```

**`SRem(ctx, key, members...) (int64, error)`**  
Removes one or more members from a set. Returns the number of members removed.

```go
rc.SRem(ctx, "online:users", "user-1")
```

---

### Raw client

**`GetClient() *redis.Client`**  
Returns the underlying `go-redis` client for commands not covered above.

```go
client := rc.GetClient()
client.ZAdd(ctx, "leaderboard", redis.Z{Score: 100, Member: "user-1"})
```
