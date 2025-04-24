package logger

import (
	"fmt"
)

func ExampleInitialize() {
	err := Initialize("debug")
	err = Initialize("info")
	if err != nil {
		fmt.Print(err)
	}

	Log.Info("info")

}
