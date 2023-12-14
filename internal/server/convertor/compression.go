package convertor

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
)

type CompressionFactor float64

func (c CompressionFactor) LogValue() slog.Value {
	return slog.StringValue(strconv.FormatFloat(float64(c), 'f', 2, 64))
}

func CalculateCompression(before, after string) (CompressionFactor, error) {
	fileInfoBefore, err := os.Stat(before)
	if err != nil {
		return 0, fmt.Errorf("stat failed: %w", err)
	}
	fileInfoAfter, err := os.Stat(after)
	if err != nil {
		return 0, fmt.Errorf("stat failed: %w", err)
	}

	return CompressionFactor(float64(fileInfoAfter.Size()) / float64(fileInfoBefore.Size())), nil
}
