package event

import "testing"

func TestLoadFromStorage(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"Hello, world", "dlrow ,olleH"},
		{"Hello, 世界", "界世 ,olleH"},
		{"", ""},
	}
	for _, c := range cases {
		got := Reverse(c.in)
		if got != c.want {
			t.Errorf("Reverse(%q) == %q, want %q", c.in, got, c.want)
		}
	}
}

/*
SET "EV1" '{"Id":"EV1", "Email":"jojo@po.com", "Who":"morgan", "When":"2002-10-02T15:00:00Z", "Where":{"Lat":1.5, "Lng":1.2}, "What":"une histpoire de fesse", "Pic":["img1", "img2"]}'
SET "EV2" '{"Id":"EV2", "Email":"jojo@po.com", "Who":"morgan", "When":"2002-10-02T15:00:00Z", "Where":{"Lat":12, "Lng":38}, "What":"une fesse", "Pic":["img13", "img4"]}'
SET "EV3" '{"Id":"EV3", "Email":"jojo@po.com", "Who":"matthieu", "When":"2002-10-02T15:00:00Z", "Where":{"Lat":24, "Lng":24}, "What":"une histpoire", "Pic":["img12", "img9"]}'
*/
