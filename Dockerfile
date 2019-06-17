# Container is based on a preexisting image that contains the Go tools needed
# to compile and install
FROM golang:1.10 AS golang_base

# Project URI based on repository URL 
ENV PROJECT_URI=github.com/smartedge/codechallenge
ENV PROJECT_DIR=${GOPATH}/src/${PROJECT_URI}

# Disable cgo to prevent dynamically linked binaries:
ENV CGO_ENABLED=0

# Create project directory and output directory
RUN mkdir -p ${PROJECT_DIR} /output

# Change current working directory to project directory
WORKDIR ${PROJECT_DIR}

# Build in clean environment
FROM golang_base AS builder

# Copy source code to project directory
COPY . ${PROJECT_DIR}

# Compile and install code
RUN go install ${PROJECT_URI}/cmd/codechallenge

# Move binary to predictable location:
#     (Utilize the shell to enable variable substitution for the
#     GOPATH variable)
RUN sh -c "cp \"${GOPATH}/bin/codechallenge\" /output/codechallenge"

# Build in clean environment
FROM scratch AS prod_image

# Copy binary to production container image
COPY --from=builder /output/codechallenge /codechallenge

# Create directory to mount and store private and public keys
WORKDIR /mnt/home
VOLUME /mnt/home/

# Direct application to store keys in mounted volume.
ENV HOME=/mnt/home

# Configure the container entrypoint so that it runs the compiled program.
ENTRYPOINT [ "/codechallenge" ]
