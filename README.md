# Short Message Signer
## An employment coding challenge for Smart Edge of Irvine, CA

### Purpose:
While this is based on a real world problem, the goal of this project is to demonstrate my aptitude and potential value to Smart Edge as a future employee.

### What it does:
Given a short message (250 characters or less) from standard input it will
* Load (or generate and save, if missing) a public+private key pair
* Sign the message with the private key
* Deliver the signed massage in a JSON bundle, with the signature and public key, and dump them to standard outout.

For more details on how this tool is supposed to work, the Spec document can be [found here.](https://smart-edge.com/codechallenge/)

### Getting started:
For sake of simplicity, I will assume you are working with this package from a Linux system. While Go and Docker support other sytems, for clarity, building and developing this project is most straight forward on current Linux machine with Docker installed.

#### Checking out the repository
While the project should build and run in any directory, Go is rather pickey about where source code should live, it is strongly recommended to check this out at the proper place within the `$GOPATH`. TO checkout this project into a new `$GOPATH` try this:
```
$  export GOPATH="$(pwd)/gopath"
$  mkdir -p gopath/src/github.com/smartedge
$  git clone https://github.com/loren-osborn/smartEdge_codeChallenge.git gopath/src/github.com/smartedge/codechallenge
$  cd gopath/src/github.com/smartedge/codechallenge
```
