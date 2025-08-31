# We set a nonsense default value so Docker will stop complaining about missing default,
# while still requiring a sensible value set.
ARG DOCKER_BASE_IMAGE=scratch
FROM ${DOCKER_BASE_IMAGE} AS source

# Install general build tools
RUN --mount=target=/var/lib/apt/lists,type=cache,sharing=locked \
    --mount=target=/var/cache/apt,type=cache,sharing=locked \
    rm -f /etc/apt/apt.conf.d/docker-clean               && \
    apt-get update                                       && \
    apt-get upgrade -yq                                  && \
    apt-get install -yq                                     \
        build-essential                                     \
        clang-format                                        \
        curl                                                \
        locales                                             \
        make

# Install vulkan tools and libs
RUN --mount=target=/var/lib/apt/lists,type=cache,sharing=locked \
    --mount=target=/var/cache/apt,type=cache,sharing=locked \
    rm -f /etc/apt/apt.conf.d/docker-clean               && \
    apt-get update                                       && \
    apt-get upgrade -yq                                  && \
    apt-get install -yq                                     \
        glslang-tools                                       \
        libvulkan-dev                                       \
        vulkan-tools                                        \
        vulkan-validationlayers

# Install cross-compilation tools and dependencies
RUN --mount=target=/var/lib/apt/lists,type=cache,sharing=locked \
    --mount=target=/var/cache/apt,type=cache,sharing=locked \
    dpkg --add-architecture arm64                        && \
    apt-get update                                       && \
    apt-get install -yq                                     \
        gcc-aarch64-linux-gnu                               \
        libc6-dev-arm64-cross                               \
        libvulkan-dev:arm64

# Install test runtime dependencies
RUN --mount=target=/var/lib/apt/lists,type=cache,sharing=locked \
    --mount=target=/var/cache/apt,type=cache,sharing=locked \
    apt-get update                                       && \
    apt-get install -yq                                     \
        ffmpeg                                              \
        unzip

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

# Get Go project modules
COPY --chown=build:build go.mod go.sum /src/
RUN --mount=type=cache,target=~/.cache/go-build \
    --mount=type=cache,target=~/go/pkg/mod      \
    go mod download

# Build staticcheck
ARG GO_STATICCHECK_VERSION
RUN --mount=type=cache,target=~/.cache/go-build \
    --mount=type=cache,target=~/go/pkg/mod      \
    go tool staticcheck -h

# Copy sources
COPY --chown=build:build . /src/

# Build for host and arm64
FROM source AS build
RUN --mount=type=cache,target=~/.cache/go-build \
    --mount=type=cache,target=~/go/pkg/mod      \
    make build_host_vk
RUN --mount=type=cache,target=~/.cache/go-build \
    --mount=type=cache,target=~/go/pkg/mod      \
    make build_arm64

FROM scratch AS export
COPY --from=build /src/build/ /

# Run lint
FROM source AS lint
RUN --mount=type=cache,target=~/.cache/go-build \
    --mount=type=cache,target=~/go/pkg/mod      \
    make lint

# Run tests
FROM source AS test
RUN --mount=type=cache,target=~/.cache/go-build \
    --mount=type=cache,target=~/go/pkg/mod      \
    make test

# Run more tests
FROM source AS test_more
RUN curl -o internal/pkg/stitch/testdata/more-testdata.zip https://trains.jo-m.ch/testdata.zip
RUN unzip -d internal/pkg/stitch/testdata internal/pkg/stitch/testdata/more-testdata.zip
RUN --mount=type=cache,target=~/.cache/go-build \
    --mount=type=cache,target=~/go/pkg/mod      \
    make test_more

# Run bench
FROM source AS bench
RUN --mount=type=cache,target=~/.cache/go-build \
    --mount=type=cache,target=~/go/pkg/mod      \
    make bench
