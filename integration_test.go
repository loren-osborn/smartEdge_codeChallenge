package codechallenge_test

import (
	"crypto/x509"
	"fmt"
	"github.com/smartedge/codechallenge"
	"github.com/smartedge/codechallenge/deps"
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
	// These are both based on the known random math seed, but may be platform
	// specific.
	ExpectedInitialECDSAPublicKey = "-----BEGIN ECDSA PUBLIC KEY-----\n" +
		"MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE7WzVjtn9Gk+WHr5xbv8XMvooqU25\n" +
		"BhgNjZ/vHZLBdVtCOjk4KxjS1UBfQm0c3TRxWBl3hj2AmnJbCrnGofMHBQ==\n" +
		"-----END ECDSA PUBLIC KEY-----\n"
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
			homeDir:  "/home/foobar",
			argList:  []string{"myProg"},
			stdInput: "Four score and seven years ago...",
			status:   0,
			stdOutput: testtools.GetResponseMatcherForMessageAndPubKey(
				deps.Defaults,
				"Four score and seven years ago...",
				ExpectedInitialECDSAPublicKey),
			stdErr: testtools.NewStringStringMatcher(""),
		},
		"Basic test of RSA message signing": {
			homeDir:  "/home/anybody",
			argList:  []string{"codechallenge", "-rsa"},
			stdInput: "Do Re Mi Fa So La Ti Do\n",
			status:   0,
			stdOutput: testtools.GetResponseMatcherForMessageAndAlgorithm(
				deps.Defaults,
				"Do Re Mi Fa So La Ti Do",
				x509.RSA),
			stdErr: testtools.NewStringStringMatcher(""),
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
			homeDir:  "/home/anybody",
			argList:  []string{"codechallenge", "-ascii"},
			stdInput: DeclarationOfIndependanceFirst250Chars + "   \t\t\n\n \n",
			status:   0,
			stdOutput: testtools.GetResponseMatcherForMessageAndPubKey(
				deps.Defaults,
				DeclarationOfIndependanceFirst250Chars,
				ExpectedInitialECDSAPublicKey),
			stdErr: testtools.NewStringStringMatcher(""),
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
			homeDir:  "/home/anybody",
			argList:  []string{"codechallenge", "-binary"},
			stdInput: DeclarationOfIndependanceFirst250Chars,
			status:   0,
			stdOutput: testtools.GetResponseMatcherForMessageAndPubKey(
				deps.Defaults,
				DeclarationOfIndependanceFirst250Chars,
				ExpectedInitialECDSAPublicKey),
			stdErr: testtools.NewStringStringMatcher(""),
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
			if err := tc.stdOutput.MatchString(mockDepsBundle.OutBuf.String()); err != nil {
				tt.Errorf("Standard Output:\n%#v didn't match:\n%s.", mockDepsBundle.OutBuf.String(), err.Error())
			}
			if err := tc.stdErr.MatchString(mockDepsBundle.ErrBuf.String()); err != nil {
				tt.Errorf("Standard Error:\n%#v didn't match:\n%s.", mockDepsBundle.ErrBuf.String(), err.Error())
			}
		})
	}
}
