package config

import (
	"testing"
)

func BenchmarkNewConfig(b *testing.B) {
	b.Run("NewConfig", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = New()
		}
	})
}
