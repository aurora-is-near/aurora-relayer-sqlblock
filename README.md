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

## TODO
Add tests
