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
			stdOutput: "{\n\"message\": \"Four score and seven years ago...\",\n\"signature\": \"MEUCIHH0Hn6Utz2h8iRIzA6B+d6ZtYhfgyJ/chWnSptL8Mw4AiEA/d//iAXBvTKuySbBRHdBl/pzPTnVTe23DvmLHihG1PY=\",\n\"pubkey\": \"-----BEGIN ECDSA PUBLIC KEY-----\\nMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAETt2oryDODBqRg91xm5sH0HfCkrvT\\nN4s4iMHiCIMZf8US0mFVABz9PtnmUhYfGjmBpAd6c1wgesu9Sc3peXJywQ==\\n-----END ECDSA PUBLIC KEY-----\\n\"\n}",
			stdErr:    "",
		},
		// {
		// 	homeDir:   "/home/anybody",
		// 	argList:   []string{"codechallenge", "-rsa"},
		// 	stdInput:  "Do Re Mi Fa So La Ti Do",
		// 	status:    0,
		// 	stdOutput: "wrong",
		// 	stdErr:    "wrong",
		// },
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
