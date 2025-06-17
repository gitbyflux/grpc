package sl

import (
	"fmt"
	"log/slog"
)

func Err(err error) slog.Attr {
	return slog.Attr{
		Key:   "error",
		Value: slog.StringValue(err.Error()),
	}
}

func Wrap(op string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", op, err)
}

func WrapMsg(op string, str string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %s %w", op, str, err)
}
