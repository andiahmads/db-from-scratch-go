package main

import (
	"fmt"
	"my-own-db/internal/kv"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		return
	}

	dbPath := "../data/my.db"

	db, err := kv.Open(dbPath)
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}

	defer db.Close()

	cmd := os.Args[1]

	switch cmd {
	case "put":
		// my db put <key> <value>
		if len(os.Args) != 4 {
			usage()
			return
		}

		key := os.Args[2]
		value := []byte(os.Args[3])

		if err := db.Put(key, value); err != nil {
			fmt.Println("error:", err)
			os.Exit(1)
		}
		fmt.Println("ok")

	case "get":
		// get data <get>
		if len(os.Args) != 3 {
			usage()
			return
		}

		key := os.Args[2]
		val, ok := db.Get(key)
		if !ok {
			fmt.Println("nil")
			return
		}
		fmt.Println(string(val))

	case "del":
		// mydb del <key>
		if len(os.Args) != 3 {
			usage()
			return
		}

		key := os.Args[2]
		if err := db.Delete(key); err != nil {
			fmt.Println("error:", err)
			os.Exit(1)
		}
		fmt.Println("ok")

	case "dump":
		fmt.Print(db.DebugDump())

	default:
		usage()
	}

}

func usage() {
	fmt.Println("Usage:")
	fmt.Println("  mydb put <key> <value>")
	fmt.Println("  mydb get <key>")
	fmt.Println("  mydb del <key>")
	fmt.Println("  mydb dump")
}
