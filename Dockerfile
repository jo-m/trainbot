ARG DOCKER_BASE_IMAGE
FROM ${DOCKER_BASE_IMAGE}

# Install tools
RUN export DEBIAN_FRONTEND=noninteractive                   \
           DEBCONF_NONINTERACTIVE_SEEN=true              && \
    apt-get update                                       && \
    apt-get upgrade -yq                                  && \
    apt-get install -yq                                     \
        build-essential                                     \
        clang-format                                        \
        curl                                                \
        locales                                             \
        make                                             && \
    apt-get clean                                        && \
    rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

# Install cross-compilation tools and dependencies
RUN export DEBIAN_FRONTEND=noninteractive                   \
           DEBCONF_NONINTERACTIVE_SEEN=true              && \
    apt-get update                                       && \
    apt-get install -yq                                     \
        gcc-aarch64-linux-gnu                               \
        libc6-dev-arm64-cross                            && \
    apt-get clean                                        && \
    rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

# Install test runtime dependencies
RUN export DEBIAN_FRONTEND=noninteractive                   \
           DEBCONF_NONINTERACTIVE_SEEN=true              && \
    apt-get update                                       && \
    apt-get install -yq                                     \
        ffmpeg                                           && \
    apt-get clean                                        && \
    rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

# Add unprivileged build user
RUN adduser --gecos '' --disabled-password build
RUN     mkdir -p /src /build                             && \
    chown build:build /src /build

# Install Go: see https://golang.org/doc/install
ARG GO_VERSION
ARG GO_ARCHIVE_SHA256
ENV GO_ARCHIVE="go${GO_VERSION}.linux-amd64.tar.gz"         \
    GOPATH=/home/build/go                                   \
    PATH="/usr/local/go/bin:/home/build/go/bin:${PATH}"

RUN curl -OL "https://golang.org/dl/${GO_ARCHIVE}"               && \
    echo "${GO_ARCHIVE_SHA256} ${GO_ARCHIVE}" | sha256sum -c     && \
    tar -C /usr/local -xzf "${GO_ARCHIVE}"                       && \
    rm "${GO_ARCHIVE}"

# Become build user
WORKDIR /src
USER build

# Install staticcheck
ARG GO_STATICCHECK_VERSION
RUN go install "honnef.co/go/tools/cmd/staticcheck@${GO_STATICCHECK_VERSION}"

# Get Go project modules
COPY --chown=build:build go.mod go.sum /src/
RUN go mod download

# Copy sources
COPY --chown=build:build . /src/

# Build for host and run checks and tests
RUN make check
RUN make build_host
RUN mv build/* /build/

# Build for arm64
RUN make build_arm64
RUN mv build/* /build/

RUN ls -l /build
