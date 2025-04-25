package logger

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func ExampleInitialize_error() {
	err := Initialize("invalid_level")
	if err != nil {
		fmt.Print(err)
	}
	// Output:
	// unrecognized level: "invalid_level"
}

func ExampleLog_Debug() {
	err := Initialize("debug")
	if err != nil {
		fmt.Print(err)
		return
	}

	Log = Log.WithOptions(zap.WrapCore(func(core zapcore.Core) zapcore.Core {

		encoderConfig := zap.NewProductionEncoderConfig()
		encoderConfig.TimeKey = ""
		newEncoder := zapcore.NewJSONEncoder(encoderConfig)
		return zapcore.NewCore(
			newEncoder,
			zapcore.AddSync(zapcore.Lock(os.Stdout)),
			zapcore.DebugLevel,
		)
	}))

	Log.Debug("debug message")
	Log.Info("info message")

	// Output:
	// {"level":"debug","caller":"logger/example_test.go:39","msg":"debug message"}
	// {"level":"info","caller":"logger/example_test.go:40","msg":"info message"}
}
