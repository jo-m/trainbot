ARG DOCKER_BASE_IMAGE
FROM ${DOCKER_BASE_IMAGE} AS build

WORKDIR /app
COPY package.json yarn.lock ./
RUN yarn install

COPY . .
RUN yarn run build

FROM scratch AS export
COPY --from=build /app/dist /
