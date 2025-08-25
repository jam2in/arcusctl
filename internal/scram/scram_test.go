package scram

import (
	"encoding/base64"
	"testing"
)

func TestScram(t *testing.T) {
	type Testcase struct {
		password string
		salt     string
		expected string
	}
	testcases := []Testcase{
		{
			password: "testPassword",
			salt:     "xdKecnjG8Y1Fbg8ZPjKKhg==",
			expected: "SCRAM-SHA-256$4096:xdKecnjG8Y1Fbg8ZPjKKhg==$jl7gNyK87Pgx2WBxBmDKzfeV88M5jSDP+8Gc7eoCqe0=:Lq3B7wMUgtrHTXJhfFceOHCMbAu6FYeXNMlABhfX4ig=",
		},
		{
			password: "jam2in",
			salt:     "wwN0pPa36OcpozXMKzuMJA==",
			expected: "SCRAM-SHA-256$4096:wwN0pPa36OcpozXMKzuMJA==$1G6TfzfDKcOAOSuSyPsN1RxgJI9GcoCgj/YJaxLsRIk=:RxonKjYzEk0ttzyjnne0TlsiYuAmygSsQK0DgOENcIM=",
		},
		{
			password: "1111",
			salt:     "zYKWe/tZWx8F+2mTggAgSA==",
			expected: "SCRAM-SHA-256$4096:zYKWe/tZWx8F+2mTggAgSA==$+bYGADd+OLokk/fRrnN4ubZ8pyl6gUnZPKagLP9nHMk=:8yDApj5iIjh26Rs8aw8AFvpVd7rtNIox41GnhLQzEBY=",
		},
	}

	for _, tc := range testcases {
		salt, err := base64.StdEncoding.DecodeString(tc.salt)
		if err != nil {
			t.Fatal(err)
		}

		secret := GenerateScramSHA256Secret(tc.password, salt, 0)
		if string(secret.EncodeToBase64()) != tc.expected {
			t.Errorf("\n%s\n%s", tc.expected, secret.EncodeToBase64())
		}
	}
}
