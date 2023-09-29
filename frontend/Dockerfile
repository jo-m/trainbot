ARG DOCKER_BASE_IMAGE
FROM ${DOCKER_BASE_IMAGE} AS builder

WORKDIR /app
COPY package.json yarn.lock ./
RUN yarn install

COPY . .
RUN yarn run build

FROM scratch AS export
COPY --from=builder /app/dist /
