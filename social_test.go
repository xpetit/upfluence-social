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
		assertError(`data: {"pin":{"timestamp":null}}`)
		assertError(`data: {"pin":{"timestamp":0}}`)
		assertError(`data: {"pin":{"timestamp":0,"id":null}}`)
		assertError(`data: {"pin":{"timestamp":0,"id":0,"comments":[]}}`)
	})
}

func BenchmarkParseEvent(b *testing.B) {
	payload := []byte(`data: {"pin":{"id":99941156,"title":"","description":"VS Email","links":"","likes":8432,"comments":3,"saves":0,"repins":1,"timestamp":1686635858,"post_id":"1090574865996212582"}}`)
	for i := 0; i < b.N; i++ {
		if _, err := parseEvent(payload); err != nil {
			b.Fatal(err)
		}
		b.SetBytes(int64(len(payload)))
	}
}
