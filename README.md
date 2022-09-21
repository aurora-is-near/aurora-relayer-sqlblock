# Aurora Relayer block sql

Lib that converts Block struct to INSERT SQL.

## How to use

```go
import (
  sqlblock "github.com/aurora-is-near/aurora-relayer-sqlblock"
)

func main() {
  var block sqlblock.Block
  json.Unmarshal([]byte(content), &block)
  sql := block.InsertSql()
  fmt.Println(sql)
}
```

## How to test
1. `cp config/test.yaml_example config/test.yaml`
2. Modify `database` in `config/test.yaml` file.
3. Run `go test`
