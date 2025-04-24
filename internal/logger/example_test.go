package logger

import (
	"fmt"
)

func Example() {
	err := Initialize("debug")
	if err != nil {
		fmt.Errorf("ups logger")
	}

	Log.Info("info")

}
