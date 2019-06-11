package codechallenge_test

import (
	"fmt"
	"github.com/smartedge/codechallenge"
	"github.com/smartedge/codechallenge/testtools/mocks"
	"testing"
)

// TestCallingMainWithMocks verifies that calling RealMain with mocked
// dependencies works as intended.
func TestCallingMainWithMocks(t *testing.T) {
	for i, tc := range []struct {
		homeDir   string
		argList   []string
		stdInput  string
		status    int
		stdOutput string
		stdErr    string
	}{
		{
			homeDir:   "/home/foobar",
			argList:   []string{"myProg"},
			stdInput:  "Four score and seven years ago...",
			status:    0,
			stdOutput: "{\n\"message\": \"Four score and seven years ago...\",\n\"signature\": \"MEUCIQDkREv9Q5S3L/K5IVA6NZP9N+8b0ZDHQ8R85BmmMhih7QIgTYziTZFjjk5qg0/+c+hEtP+37yfZLNVereQdQgIhiEo=\",\n\"pubkey\": \"-----BEGIN ECDSA PUBLIC KEY-----\\nMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE7WzVjtn9Gk+WHr5xbv8XMvooqU25\\nBhgNjZ/vHZLBdVtCOjk4KxjS1UBfQm0c3TRxWBl3hj2AmnJbCrnGofMHBQ==\\n-----END ECDSA PUBLIC KEY-----\\n\"\n}",
			stdErr:    "",
		},
		{
			homeDir:   "/home/anybody",
			argList:   []string{"codechallenge", "-rsa"},
			stdInput:  "Do Re Mi Fa So La Ti Do\n",
			status:    0,
			stdOutput: "{\n\"message\": \"Do Re Mi Fa So La Ti Do\",\n\"signature\": \"GdTgkUWpufwusew4mk19bbUpmAPvBcyTPRC+/f+wLIBY+/PqLFLgspDEw0cmM0wWfid6nb66XugoKJLRfzYl9DtMdS85DhHE4t5wW9KJnY5pe8qb3xlcS/l4KYnxOQd5yZByge17QlbopcJ3SgdZwsVj//uJTbLfWXuUjUvyNH0furSrUZpEWmjrNomcpVQXnNvyQNnZmoL1wA0Kpvko6tzfnG7fOKso1ivEcCrxgPDsQyJbwkzCtjD2sDhh55avwUJ1hRGvUyytxd4BSb/yZfHVsBWxRk25lnF+3z9hpVKhOMocU6fbrLigCJP+kMiujeEYiGJXOOpc7CF5U40dCg==\",\n\"pubkey\": \"-----BEGIN RSA PUBLIC KEY-----\\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAzCTTFKQBHfTN8jW6q8PT\\nHNZKWnRPxSt9kpgWmyqFaZnEUipgoKGAxSIsVrl2PJSm5OlgkVzx+MY+LWM64VKM\\nbRpUUGJR3zdMNhwZQX0hjOpLpVJvUwD78utVs8vijrU7sH48usFiaZQYjy4m4hQh\\n63/x4h3KVz7YqUnlRMzYJFT43+AwYzYuEpzWRxtW7IObJPtjtmYVoqva98fF6aj5\\nuHAsvaAgZGBalHXmCiPzKiGU/halzXSPvyJ2Cqz2aUqMHgwi/2Ip4z/mrfX+mUTa\\nS+LyBy7GgqJ5vbkGArMagJIc0eARF60r6Uf483xh17oniABdLJy4qlLf6PcEU+ut\\nEwIDAQAB\\n-----END RSA PUBLIC KEY-----\\n\"\n}",
			stdErr:    "",
		},
	} {
		t.Run(fmt.Sprintf("Subtest %d", i+1), func(tt *testing.T) {
			mockDepsBundle := mocks.NewDefaultMockDeps(tc.stdInput, tc.argList, tc.homeDir, nil)
			err := mockDepsBundle.InvokeCallInMockedEnv(func() error {
				codechallenge.RealMain(mockDepsBundle.Deps)
				return nil
			})
			if err != nil {
				tt.Errorf("Unexpected error calling mockDepsBundle.InvokeCallInMockedEnv(): %s", err.Error())
			}
			if exitStatus := mockDepsBundle.GetExitStatus(); exitStatus != tc.status {
				tt.Errorf("RealMain() should have a normal exit status of %d. Got %#v instead.", tc.status, exitStatus)
			}
			if tc.stdOutput != mockDepsBundle.OutBuf.String() {
				tt.Errorf("We didn't see the expected output of:\n%#v\nInstead we got:\n%#v.", tc.stdOutput, mockDepsBundle.OutBuf.String())
			}
			if tc.stdErr != mockDepsBundle.ErrBuf.String() {
				tt.Errorf("We didn't see the expected output of:\n%#v\nInstead we got:\n%#v.", tc.stdErr, mockDepsBundle.ErrBuf.String())
			}
		})
	}
}
