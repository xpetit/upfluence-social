package social

import "testing"

func TestParseEvent(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		assertEqual := func(payload string, want Event) {
			t.Helper()
			got, err := parseEvent([]byte(payload))
			if err != nil {
				t.Fatal("unexpected error:", err)
			}
			if *got != want {
				t.Errorf("\ngot:  %s\nwant: %s", got, want)
			}
		}

		assertEqual(`data: {"pin":{"timestamp":0,"id":0}}`, Event{})
		assertEqual(`data: {"pin":{"timestamp":1,"id":1}}`, Event{UnixTime: 1, ID: 1})
		assertEqual(`data: {"pin":{"timestamp":2,"id":3,"likes":4,"comments":5}}`,
			Event{UnixTime: 2, ID: 3, Counts: [len(Dimensions)]uint32{4, 5}},
		)
	})

	t.Run("errors", func(t *testing.T) {
		assertError := func(payload string) {
			t.Helper()
			if _, err := parseEvent([]byte(payload)); err == nil {
				t.Error(payload, "is invalid and should not be parsed successfully")
			}
		}
		assertError(``)
		assertError(`{}`)
		assertError(`{"pin":{"timestamp":0,"id":0}}`)
		assertError(`data: `)
		assertError(`data: null`)
		assertError(`data: {}`)
		assertError(`data: {"pin":null`)
		assertError(`data: {"pin":{}`)
		assertError(`data: {"pin":{"timestamp":0}}`)
		assertError(`data: {"pin":{"timestamp":0,"id":null}}`)
	})
}
