# Short Message Signer
## An employment coding challenge for [Smart Edge of Irvine, CA](https://smart-edge.com/)

### Purpose:
While this is based on a real world problem, the goal of this project is to demonstrate my aptitude and potential value to [Smart Edge](https://smart-edge.com/) as a future employee.

### What it does:
Given a short message (250 characters or less) from standard input it will:
* If no key pair is found (for the requested algorithm) on the filesystem:
    * Generate and save, a new public+private key pair for the specified cryptography algorithm, to the filesystem
* Load the correct public+private key pair
* Sign the message with the private key
* Verify that the signature that was generated matches the public key
* Emit the signed massage, with the signature and public key, in JSON format, to standard output.

For more details on how this tool is supposed to work, the specification document can be [found here.](https://smart-edge.com/codechallenge/)

### Getting started:
For sake of simplicity, I will assume you are working with this package from a Linux system. While Go and Docker support other sytems, for clarity, building and developing this project is most straight forward on a current Linux machine with Docker installed.

#### Checking out the repository
Thanks to Docker, the project should build and run in any directory, but since Go is rather pickey about where source code should live, it is strongly recommended to check this project out, at the proper place, within the `$GOPATH`. To checkout this project into a new `$GOPATH`, and build it there try this:
```
$  export GOPATH="$(pwd)/gopath"
$  mkdir -p gopath/src/github.com/smartedge
$  git clone https://github.com/loren-osborn/smartEdge_codeChallenge.git gopath/src/github.com/smartedge/codechallenge
$  cd gopath/src/github.com/smartedge/codechallenge
$  make
```
The `Makefile` will then (by default):
* Build a `tester_image` docker image
* Run this container, which will:
    * "Vet" the code with builtin `go vet` tool
    * "Lint" the code with [`revive`](https://github.com/mgechev/revive) (enhanced alternative to `golint`)
    * Run the unit tests with colorized [`gotest`](https://github.com/rakyll/gotest) (`go test` work-alike tool)
    * Produce an HTML [code coverage report in `coverage.html`](https://loren-osborn.github.io/smartEdge_codeChallenge/coverage.html)
    * Generate [HTML godoc documentation](https://loren-osborn.github.io/smartEdge_codeChallenge/godoc/pkg/github.com/smartedge/codechallenge/index.html) by using modified tool [`utils/gendoc.sh`](https://github.com/loren-osborn/smartEdge_codeChallenge/blob/master/utils/gendoc.sh) from [https://gist.github.com/Kegsay/84ce060f237cb9ab4e0d2d321a91d920](https://gist.github.com/Kegsay/84ce060f237cb9ab4e0d2d321a91d920) to scrape the output of the official `godoc` tool with `wget`.
* Build the `production_container_image` in a `scratch` docker image
* Build a `demo_image` and demonstrate running `echo Do Re Mi Fa So La Ti Do | ${GOPATH}/bin/codechallenge -rsa` in it

#### Building outside a container
To build the executable outside the container (assuming that your `$GOPATH` is set correctly) type:
```
$  make build_local
```
Then you can execute it with:
```
$  ./codechallenge.exe 
```
(the `.exe` is to avoid collision with the directory name. To avoid this use `go install`)
```
$  go install github.com/smartedge/codechallenge/codechallenge
```
but this will require executing it with:
```
$  $GOPATH/bin/codechallenge
```

#### Command line options
The tool recognizes the following options:
```
Usage of ./codechallenge.bin:
  Input format options:
      -ascii
        	This specifies that the message is ASCII content
      -binary
        	This specifies that the message is raw binary content
      -utf8
        	This specifies that the message is UTF-8 content [default]
  Algorithm options:
      -ecdsa
        	Causes the mesage to be signed with an ECDSA key-pair [default]
      -rsa
        	Causes the mesage to be signed with an RSA key-pair
      -bits uint
        	Bit length of the RSA key [default=2048]
  -private string
    	filepath of the private key file. Defaults to ~/.smartEdge/id_rsa.priv for RSA and ~/.smartEdge/id_ecdsa.priv for ECDSA.
  -public string
    	filepath of the private key file. Defaults to ~/.smartEdge/id_rsa.pub for RSA and ~/.smartEdge/id_ecdsa.pub for ECDSA.
```
(This is not the current tool output. I have modified it for clarity.)

### Guided Tour:
This should help you find your way around the files in the repository:

#### Source Code:
* In addition to all files ending in `_test.go`, all code in the `testtools/` directory is used exclusively for unit testing. 
* Except for a stub entrypoint in `codechallenge/main.go`, all production source files are in the root directory of the project.
* Code in the `testtools/mocks/` directory help build a reusable mocked environment for unit testing.
* Code in the `testtools/` directory provide comparison and utility functions for unit testing.

#### Other directories in repository:
* A bash script in the `utils/` directory helps scrape generated godoc documentation into HTML files.
* The `testdata/` directory contains test data used for comparison. (Currently the unused `valid_output_schema.json` from the specification document lives here.)

#### Generated files:
* The HTML `godoc` documentation files are stored in the `godoc/` directory
* The files `coverage.out` and `coverage.html` contain the code coverage report from `go test` in text and HTML format respectively.

### Design Considerations
**Dependency Injection:** I started this project with testability as a top priority. To achieve this, I made dependency injection a core component to how the parts of this project fit together. All connections with externalities, including the random entropy source, go through the `*Dependencies` struct created at program invocation. This allows any of them to be easily mocked for testing purposes.

**TDD:** While I had to change gears due to other demands on my time, I started this project doing exclusively Test Driven Development. This allowed me to build out very solid **reusable** testing infrustructure to build the project in an fully testable manner. While I had to alter my development style midstream, I remain confident this code base can be brought up to 100% code coverage without modification, and modest effort.

**BDD:** I had intended to use this project to learn and use Ginkgo and Gomega for the production features of the codechallenge, and completed implementing a Gomega matcher for JSON schema validation for this purpose. While I'm excited to use Ginkgo with other projects at Smart Edge, I was unable to include any Ginkgo BDD tests in this project.

**Stub Entrypoint:** I put the actual `main()` entrypoint in a stub subpackage for two reasons:
* Implementing the majority of the code **outside** the `main` package allowed go tests to easily test the distinction between public and private package members. (I did not put much effort into identifying package members that should be kept private yet.)
* `godoc` refuses to produce API documentation for functions in the `main` package, so moving all substantive code out of the `main` package allowed for generation of very readable automated documentation.

### Future Development
These are areas where I would improve this project given more time:
* Easy wins (minimal effort changes)
    * Make the "Usage" output match the output above.
    * Add a exit-status 0  `-help` option
* Testing:
    * Bring the project to 100% test coverage: Both for production code and testtools code.
    * Add BDD tests for "bullet point" features (and feature details) for feature tracability.
    * Ensure that all edge cases, and resulting behaviors are properly tested, and verified, rather than simply making sure that all code paths executed.
* Forward Compatibility: Ensure that the project works properly in Go 1.11, with go modules, outside the `$GOPATH`
* Add a mode where the signed messages could be validated. The tool already does this internally, but it would be beneficial to expose this to the user.
* Look for more docker-friendly storage options for public and private keys than on the filesystem.
* Refactoring directions:
    * The `flag` provides a fairly robust way add custom option types that would be cleaner and less confusing than the current implementation. I would suggest adding the following flag types:
        * A one of many, "mutually exclusive" flag option type (for content type and algorithm selection)
        * A conditionally allowed option (for rsa bit-length)
        * A string flag where the default can be computed dynamically, or set after `flag.Parse()`
    * Reorganize what functions are in which files to make them easier to find.
    * Determine what package members should be made private, and ensure they are still fully testable
    * Investigate if the RSA PSS padding needs its own sha256 hashing function. (If not, these should be consolidated.)
* Build process: Seperate `godoc` generation from `go test` so the documentation doesn't need to be generated every time the tests don't fail.
