package sqlblock

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/spf13/viper"
)

func TestInsert(t *testing.T) {
	database := prepareDatabase()
	defer database.Close()

	content, _ := os.ReadFile("fixtures/60034225.json")
	var block Block
	json.Unmarshal([]byte(content), &block)
	block.Sequence = block.Height
	sql := block.InsertSql()
	if _, err := database.Exec(context.Background(), sql); err != nil {
		panic("Failed to import block")
	}
	blockId := block.Height

	t.Run("block created", func(t *testing.T) {
		blockCount, _ := entriesCount(database, "block")
		if blockId != 60034225 {
			t.Errorf("Wrong block height")
		}
		if blockCount != 1 {
			t.Errorf("Block not inserted")
		}
	})

	t.Run("transactions created", func(t *testing.T) {
		transactionCount, error := entriesCount(database, "transaction")
		if error != nil {
			t.Errorf(error.Error())
		}
		if transactionCount != 3 {
			t.Errorf("Transaction not inserted")
		}
	})

	t.Run("events created", func(t *testing.T) {
		eventCount, error := entriesCount(database, "event")
		if error != nil {
			t.Errorf(error.Error())
		}
		if eventCount != 17 {
			t.Errorf("Event not inserted")
		}
	})
}

func TestInsertBlockRollback(t *testing.T) {
	database := prepareDatabase()
	defer database.Close()
	content, _ := os.ReadFile("fixtures/60034225.json")
	var block Block
	json.Unmarshal([]byte(content), &block)
	block.Sequence = block.Height
	// Violate `block_check`
	block.GasLimit = "100000"
	block.GasUsed = "150000"

	sql := block.InsertSql()
	_, error := database.Exec(context.Background(), sql)
	if error == nil {
		t.Errorf("Should have not inserted block")
	}

	t.Run("block rollback", func(t *testing.T) {
		got, error := entriesCount(database, "block")
		if error != nil {
			t.Errorf(error.Error())
		}
		if got != 0 {
			t.Errorf("got %q, wanted %q", got, 0)
		}
	})

	t.Run("transaction rollback", func(t *testing.T) {
		got, error := entriesCount(database, "transaction")
		if error != nil {
			t.Errorf(error.Error())
		}
		if got != 0 {
			t.Errorf("got %q, wanted %q", got, 0)
		}
	})

	t.Run("events rollback", func(t *testing.T) {
		got, error := entriesCount(database, "event")
		if error != nil {
			t.Errorf(error.Error())
		}
		if got != 0 {
			t.Errorf("got %q, wanted %q", got, 0)
		}
	})
}

func TestInsertTransactionRollback(t *testing.T) {
	database := prepareDatabase()
	defer database.Close()

	content, _ := os.ReadFile("fixtures/60034225.json")
	var block Block
	json.Unmarshal([]byte(content), &block)
	block.Sequence = block.Height
	block.Transactions[0].From = ""
	sql := block.InsertSql()

	_, error := database.Exec(context.Background(), sql)
	if error == nil {
		t.Errorf("Should have not inserted block")
	}

	t.Run("block rollback", func(t *testing.T) {
		got, error := entriesCount(database, "block")
		if error != nil {
			t.Errorf(error.Error())
		}
		if got != 0 {
			t.Errorf("got %q, wanted %q", got, 0)
		}
	})

	t.Run("transaction rollback", func(t *testing.T) {
		got, error := entriesCount(database, "transaction")
		if error != nil {
			t.Errorf(error.Error())
		}
		if got != 0 {
			t.Errorf("got %q, wanted %q", got, 0)
		}
	})

	t.Run("events rollback", func(t *testing.T) {
		got, error := entriesCount(database, "event")
		if error != nil {
			t.Errorf(error.Error())
		}
		if got != 0 {
			t.Errorf("got %q, wanted %q", got, 0)
		}
	})
}

func TestInsertEventRollback(t *testing.T) {
	database := prepareDatabase()
	defer database.Close()

	content, _ := os.ReadFile("fixtures/60034225.json")
	var block Block
	json.Unmarshal([]byte(content), &block)
	block.Sequence = block.Height
	// Violate `event_topics_check`
	block.Transactions[0].Logs[0].Topics = [][]byte{
		[]byte("000000000000000000000000d3df29d79e10b2a5b91a65a18773dad7e7eec5c4"),
		[]byte("000000000000000000000000d3df29d79e10b2a5b91a65a18773dad7e7eec5c4"),
		[]byte("000000000000000000000000d3df29d79e10b2a5b91a65a18773dad7e7eec5c4"),
		[]byte("000000000000000000000000d3df29d79e10b2a5b91a65a18773dad7e7eec5c4"),
		[]byte("000000000000000000000000d3df29d79e10b2a5b91a65a18773dad7e7eec5c4"),
	}

	sql := block.InsertSql()
	_, error := database.Exec(context.Background(), sql)
	if error == nil {
		t.Errorf("Should have not inserted")
	}

	t.Run("block rollback", func(t *testing.T) {
		got, error := entriesCount(database, "block")
		if error != nil {
			t.Errorf(error.Error())
		}
		if got != 0 {
			t.Errorf("got %q, wanted %q", got, 0)
		}
	})

	t.Run("transaction rollback", func(t *testing.T) {
		got, error := entriesCount(database, "transaction")
		if error != nil {
			t.Errorf(error.Error())
		}
		if got != 0 {
			t.Errorf("got %q, wanted %q", got, 0)
		}
	})

	t.Run("events rollback", func(t *testing.T) {
		got, error := entriesCount(database, "event")
		if error != nil {
			t.Errorf(error.Error())
		}
		if got != 0 {
			t.Errorf("got %q, wanted %q", got, 0)
		}
	})
}

func TestDuplicateInsert(t *testing.T) {
	database := prepareDatabase()
	defer database.Close()

	content, _ := os.ReadFile("fixtures/60034225.json")
	var block Block
	json.Unmarshal([]byte(content), &block)
	block.Sequence = block.Height
	sql := block.InsertSql()
	if _, err := database.Exec(context.Background(), sql); err != nil {
		panic("Failed to import block")
	}
	blockId := block.Height

	if _, err := database.Exec(context.Background(), sql); err != nil {
		panic("Failed to import block")
	}

	t.Run("block created", func(t *testing.T) {
		blockCount, _ := entriesCount(database, "block")
		if blockId != 60034225 {
			t.Errorf("Wrong block height")
		}
		if blockCount != 1 {
			t.Errorf("Block not inserted")
		}
	})

	t.Run("transactions created", func(t *testing.T) {
		transactionCount, error := entriesCount(database, "transaction")
		if error != nil {
			t.Errorf(error.Error())
		}
		if transactionCount != 3 {
			t.Errorf("Transaction not inserted")
		}
	})

	t.Run("events created", func(t *testing.T) {
		eventCount, error := entriesCount(database, "event")
		if error != nil {
			t.Errorf(error.Error())
		}
		if eventCount != 17 {
			t.Errorf("Event not inserted")
		}
	})
}

func TestHugeInsert(t *testing.T) {
	database := prepareDatabase()
	defer database.Close()

	content, _ := os.ReadFile("fixtures/73097407.json")
	var block Block
	json.Unmarshal([]byte(content), &block)
	block.Sequence = block.Height
	sql := block.InsertSql()
	if _, err := database.Exec(context.Background(), sql); err != nil {
		panic("Failed to import block")
	}
	blockId := block.Height

	t.Run("block created", func(t *testing.T) {
		blockCount, _ := entriesCount(database, "block")
		if blockId != 73097407 {
			t.Errorf("Wrong block height")
		}
		if blockCount != 1 {
			t.Errorf("Block not inserted")
		}
	})

	t.Run("transactions created", func(t *testing.T) {
		transactionCount, error := entriesCount(database, "transaction")
		if error != nil {
			t.Errorf(error.Error())
		}
		if transactionCount != 2 {
			t.Errorf("Transaction not inserted")
		}
	})

	t.Run("transaction huge input inserted", func(t *testing.T) {
		var input string
		database.QueryRow(context.Background(), fmt.Sprintf("SELECT input::varchar FROM transaction WHERE index = 0 LIMIT 1")).Scan(&input)
		if len(input) != 1126340 {
			t.Errorf("Input size dows not match")
		}
	})
}

func prepareDatabase() *pgxpool.Pool {
	viper.AddConfigPath("config")
	viper.SetConfigName("test")
	viper.SetConfigType("yaml")
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error config file: %w \n", err))
	}
	database, _ := pgxpool.Connect(context.Background(), viper.GetString("database"))
	_, _ = database.Exec(context.Background(), "TRUNCATE block, event, transaction")
	return database
}

func entriesCount(database *pgxpool.Pool, table string) (int, error) {
	var amount int
	err := database.QueryRow(context.Background(), fmt.Sprintf("SELECT COUNT(1) FROM %s", table)).Scan(&amount)
	if err != nil {
		return 0, err
	}
	return amount, nil
}
