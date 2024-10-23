ARG DOCKER_BASE_IMAGE
FROM ${DOCKER_BASE_IMAGE} AS build

WORKDIR /app
COPY package.json package-lock.json ./
RUN npm install

COPY . .
RUN npm run build

FROM scratch AS export
COPY --from=build /app/dist /
