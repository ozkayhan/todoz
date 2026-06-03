package cli

import "time"

func ValidDate(s string) bool {
	_, err := time.Parse("2006-01-02", s)
	return err == nil
}
