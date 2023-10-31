package processor

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
)

type compressionFactor float64

func (c compressionFactor) LogValue() slog.Value {
	return slog.StringValue(strconv.FormatFloat(float64(c), 'f', 2, 64))
}

func calculateCompression(before, after string) (compressionFactor, error) {
	fileInfoBefore, err := os.Stat(before)
	if err != nil {
		return 0, fmt.Errorf("stat failed (%s): %w", before, err)
	}
	fileInfoAfter, err := os.Stat(after)
	if err != nil {
		return 0, fmt.Errorf("stat failed (%s): %w", after, err)
	}

	return compressionFactor(float64(fileInfoAfter.Size()) / float64(fileInfoBefore.Size())), nil
}
