package events

import "encoding/json"

// Encode marshals an Event to JSON without a trailing newline.
func Encode(e Event) (string, error) {
	b, err := json.Marshal(e)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// Decode unmarshals a JSON string to an Event.
func Decode(line string) (Event, error) {
	var e Event
	if err := json.Unmarshal([]byte(line), &e); err != nil {
		return Event{}, err
	}
	return e, nil
}
