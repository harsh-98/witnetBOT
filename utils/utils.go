package utils

import (
	"fmt"
)

func CheckError(err error) {
	if err != nil {
		fmt.Printf("Fatal error: %s", err.Error())
		// os.Exit(1)
	}
}
