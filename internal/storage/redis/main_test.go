//go:build integration

package redis_test

import (
	"fmt"
	"testing"

	"github.com/joho/godotenv"
)

func TestMain(m *testing.M) {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("error: " + err.Error())
	}

	m.Run()
}
