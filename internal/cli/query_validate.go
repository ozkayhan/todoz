package cli

import "errors"

var validSortBy = map[string]bool{"date": true, "title": true, "created": true, "status": true}
var validGroupBy = map[string]bool{"list": true, "date": true, "status": true}
var validStatus = map[string]bool{"pending": true, "completed": true}
var validOutputFormat = map[string]bool{"json": true, "compact": true, "csv": true}

// ValidateQueryFlags checks for conflicting or invalid flag combinations.
func ValidateQueryFlags(flags map[string]string) error {
	if flags["days-back"] != "" && flags["after-date"] != "" {
		return errors.New("conflict: use --days-back OR --after-date, not both")
	}
	if flags["no-trash"] == "true" && flags["trash-only"] == "true" {
		return errors.New("conflict: --no-trash and --trash-only are mutually exclusive")
	}
	if v := flags["status"]; v != "" && !validStatus[v] {
		return errors.New("invalid --status value: must be pending or completed, got: " + v)
	}
	if v := flags["sort-by"]; v != "" && !validSortBy[v] {
		return errors.New("invalid --sort-by value: must be one of date, title, created, status, got: " + v)
	}
	if v := flags["group-by"]; v != "" && !validGroupBy[v] {
		return errors.New("invalid --group-by value: must be list, date, or status, got: " + v)
	}
	if v := flags["output-format"]; v != "" && !validOutputFormat[v] {
		return errors.New("invalid --output-format value: must be json, compact, or csv, got: " + v)
	}
	return nil
}
