# limitador

To run the program you need to provide a YAML file with the limits. There's an
example file that allows 10 requests per minute and per user_id when the HTTP
method is "GET" and 5 when it is a "POST":
```bash
LIMITS_FILE=./examples/limits.yaml cargo run --release --bin ratelimit-server
```
