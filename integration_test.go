package codechallenge_test

import (
	"fmt"
	"github.com/smartedge/codechallenge"
	"github.com/smartedge/codechallenge/testtools"
	"github.com/smartedge/codechallenge/testtools/mocks"
	"regexp"
	"testing"
)

const (
	UsageMessageBody = "\n  -help\n" +
		"    \tdisplay this help message.\n" +
		"  Input format options:\n" +
		"      -ascii\n" +
		"        \tThis specifies that the message is ASCII content\n" +
		"      -binary\n" +
		"        \tThis specifies that the message is raw binary content\n" +
		"      -utf8\n" +
		"        \tThis specifies that the message is UTF-8 content [default]\n" +
		"  Algorithm options:\n" +
		"      -ecdsa\n" +
		"        \tCauses the mesage to be signed with an ECDSA key-pair [default]\n" +
		"      -rsa\n" +
		"        \tCauses the mesage to be signed with an RSA key-pair\n" +
		"      -bits uint\n" +
		"        \tBit length of the RSA key [default=2048]\n" +
		"  -private string\n" +
		"    \tfilepath of the private key file. Defaults to ~/.smartEdge/id_rsa.priv for RSA and ~/.smartEdge/id_ecdsa.priv for ECDSA.\n" +
		"  -public string\n" +
		"    \tfilepath of the private key file. Defaults to ~/.smartEdge/id_rsa.pub for RSA and ~/.smartEdge/id_ecdsa.pub for ECDSA.\n"
	// I had to slip in a space to have 250 characters end on a word boundry
	DeclarationOfIndependanceFirst250Chars = "When in the Course of human " +
		"events it becomes necessary for one people to dissolve the " +
		"political bands which have connected them with another and to " +
		"assume among the powers of the earth, the separate and equal " +
		"station to which the Laws of Nature  and"
)

// TestCallingMainWithMocks verifies that calling RealMain with mocked
// dependencies works as intended.
func TestCallingMainWithMocks(t *testing.T) {
	for desc, tc := range map[string]struct {
		homeDir   string
		argList   []string
		stdInput  string
		status    int
		stdOutput testtools.StringMatcher
		stdErr    testtools.StringMatcher
	}{
		"Using all defaults: (basic test case)": {
			homeDir:   "/home/foobar",
			argList:   []string{"myProg"},
			stdInput:  "Four score and seven years ago...",
			status:    0,
			stdOutput: testtools.NewStringStringMatcher("{\n\"message\": \"Four score and seven years ago...\",\n\"signature\": \"MEUCIQDkREv9Q5S3L/K5IVA6NZP9N+8b0ZDHQ8R85BmmMhih7QIgTYziTZFjjk5qg0/+c+hEtP+37yfZLNVereQdQgIhiEo=\",\n\"pubkey\": \"-----BEGIN ECDSA PUBLIC KEY-----\\nMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE7WzVjtn9Gk+WHr5xbv8XMvooqU25\\nBhgNjZ/vHZLBdVtCOjk4KxjS1UBfQm0c3TRxWBl3hj2AmnJbCrnGofMHBQ==\\n-----END ECDSA PUBLIC KEY-----\\n\"\n}"),
			stdErr:    testtools.NewStringStringMatcher(""),
		},
		"Basic test of RSA message signing": {
			homeDir:   "/home/anybody",
			argList:   []string{"codechallenge", "-rsa"},
			stdInput:  "Do Re Mi Fa So La Ti Do\n",
			status:    0,
			stdOutput: testtools.NewStringStringMatcher("{\n\"message\": \"Do Re Mi Fa So La Ti Do\",\n\"signature\": \"GdTgkUWpufwusew4mk19bbUpmAPvBcyTPRC+/f+wLIBY+/PqLFLgspDEw0cmM0wWfid6nb66XugoKJLRfzYl9DtMdS85DhHE4t5wW9KJnY5pe8qb3xlcS/l4KYnxOQd5yZByge17QlbopcJ3SgdZwsVj//uJTbLfWXuUjUvyNH0furSrUZpEWmjrNomcpVQXnNvyQNnZmoL1wA0Kpvko6tzfnG7fOKso1ivEcCrxgPDsQyJbwkzCtjD2sDhh55avwUJ1hRGvUyytxd4BSb/yZfHVsBWxRk25lnF+3z9hpVKhOMocU6fbrLigCJP+kMiujeEYiGJXOOpc7CF5U40dCg==\",\n\"pubkey\": \"-----BEGIN RSA PUBLIC KEY-----\\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAzCTTFKQBHfTN8jW6q8PT\\nHNZKWnRPxSt9kpgWmyqFaZnEUipgoKGAxSIsVrl2PJSm5OlgkVzx+MY+LWM64VKM\\nbRpUUGJR3zdMNhwZQX0hjOpLpVJvUwD78utVs8vijrU7sH48usFiaZQYjy4m4hQh\\n63/x4h3KVz7YqUnlRMzYJFT43+AwYzYuEpzWRxtW7IObJPtjtmYVoqva98fF6aj5\\nuHAsvaAgZGBalHXmCiPzKiGU/halzXSPvyJ2Cqz2aUqMHgwi/2Ip4z/mrfX+mUTa\\nS+LyBy7GgqJ5vbkGArMagJIc0eARF60r6Uf483xh17oniABdLJy4qlLf6PcEU+ut\\nEwIDAQAB\\n-----END RSA PUBLIC KEY-----\\n\"\n}"),
			stdErr:    testtools.NewStringStringMatcher(""),
		},
		"Testing contradictory signature flags": {
			homeDir:   "/home/anybody",
			argList:   []string{"codechallenge", "-rsa", "-ecdsa"},
			stdInput:  "Abcdefg",
			status:    1,
			stdOutput: testtools.NewStringStringMatcher(""),
			stdErr:    testtools.NewRegexpStringMatcher(fmt.Sprintf("^Options (-rsa and -ecdsa|-ecdsa and -rsa) may not be used together%s$", regexp.QuoteMeta("\nUsage of codechallenge:"+UsageMessageBody))),
		},
		"Testing contradictory content format flags": {
			homeDir:   "/home/anybody",
			argList:   []string{"codechallenge", "-ascii", "-utf8"},
			stdInput:  "Abcdefg",
			status:    1,
			stdOutput: testtools.NewStringStringMatcher(""),
			stdErr:    testtools.NewRegexpStringMatcher(fmt.Sprintf("^Options (-utf8 and -ascii|-ascii and -utf8) may not be used together%s$", regexp.QuoteMeta("\nUsage of codechallenge:"+UsageMessageBody))),
		},
		"Exactly 250 ascii characters": {
			homeDir:   "/home/anybody",
			argList:   []string{"codechallenge", "-ascii"},
			stdInput:  DeclarationOfIndependanceFirst250Chars + "   \t\t\n\n \n",
			status:    0,
			stdOutput: testtools.NewStringStringMatcher("{\n\"message\": \"When in the Course of human events it becomes necessary for one people to dissolve the political bands which have connected them with another and to assume among the powers of the earth, the separate and equal station to which the Laws of Nature  and\",\n\"signature\": \"MEUCIE6GnortjDvsmR/8L11YqiofLjpyEAYocYKifDTiHkWTAiEAtx9ZsbWGZNMEw5n72PTs9McuXHti95djBSLY7uxZ27k=\",\n\"pubkey\": \"-----BEGIN ECDSA PUBLIC KEY-----\\nMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE7WzVjtn9Gk+WHr5xbv8XMvooqU25\\nBhgNjZ/vHZLBdVtCOjk4KxjS1UBfQm0c3TRxWBl3hj2AmnJbCrnGofMHBQ==\\n-----END ECDSA PUBLIC KEY-----\\n\"\n}"),
			stdErr:    testtools.NewStringStringMatcher(""),
		},
		"Exactly 250 ascii characters, plus a trailing newline, but not removed in banary mode": {
			homeDir:   "/home/anybody",
			argList:   []string{"codechallenge", "-binary"},
			stdInput:  DeclarationOfIndependanceFirst250Chars + "\n",
			status:    2,
			stdOutput: testtools.NewStringStringMatcher(""),
			stdErr:    testtools.NewStringStringMatcher("Input contains more than 250 bytes (exactly 251):\n\"When in the Course of human events it becomes necessary for one people to dissolve the political bands which have connected them with another and to assume among the powers of the earth, the separate and equal station to which the Laws of Nature  and\\n\"\nUsage of codechallenge:" + UsageMessageBody),
		},
		"Exactly 250 ascii characters, works fine in banary mode without newline": {
			homeDir:   "/home/anybody",
			argList:   []string{"codechallenge", "-binary"},
			stdInput:  DeclarationOfIndependanceFirst250Chars,
			status:    0,
			stdOutput: testtools.NewStringStringMatcher("{\n\"message\": \"When in the Course of human events it becomes necessary for one people to dissolve the political bands which have connected them with another and to assume among the powers of the earth, the separate and equal station to which the Laws of Nature  and\",\n\"signature\": \"MEUCIE6GnortjDvsmR/8L11YqiofLjpyEAYocYKifDTiHkWTAiEAtx9ZsbWGZNMEw5n72PTs9McuXHti95djBSLY7uxZ27k=\",\n\"pubkey\": \"-----BEGIN ECDSA PUBLIC KEY-----\\nMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE7WzVjtn9Gk+WHr5xbv8XMvooqU25\\nBhgNjZ/vHZLBdVtCOjk4KxjS1UBfQm0c3TRxWBl3hj2AmnJbCrnGofMHBQ==\\n-----END ECDSA PUBLIC KEY-----\\n\"\n}"),
			stdErr:    testtools.NewStringStringMatcher(""),
		},
	} {
		t.Run(fmt.Sprintf("Subtest: %s", desc), func(tt *testing.T) {
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
			if !tc.stdOutput.MatchString(mockDepsBundle.OutBuf.String()) {
				tt.Errorf("We didn't see the expected output of:\n%s\nInstead we got:\n%#v.", tc.stdOutput.String(), mockDepsBundle.OutBuf.String())
			}
			if !tc.stdErr.MatchString(mockDepsBundle.ErrBuf.String()) {
				tt.Errorf("We didn't see the expected output of:\n%s\nInstead we got:\n%#v.", tc.stdErr.String(), mockDepsBundle.ErrBuf.String())
			}
		})
	}
}
