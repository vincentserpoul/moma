package user

import "testing"

func TestGetUserEmail(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"jobi@po.com", "469453d71a5e6d11027a4cd13b8a15cc"},
		{"testoste@pogmail.com", "7fbbe0f8bf93c05d9b4833aba9b7dd79"},
		{"", "d41d8cd98f00b204e9800998ecf8427e"},
	}
	for _, c := range cases {
		got := getMD5Hash(c.in)
		if got != c.want {
			t.Errorf("GetUserEmail(%q) == %q, want %q", c.in, got, c.want)
		}
	}
}

/*
SET "jojo@po.com" '{"email":"jojo@po.com", "EventList":["EV1", "EV2", "EV3"]}'
*/
