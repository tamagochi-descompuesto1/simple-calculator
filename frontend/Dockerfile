# Stage 1: build the static files. This image has Node and is discarded
# afterwards; only the built files move on to the final image.
FROM node:24.21.0-alpine AS build
WORKDIR /app

# Copy only the dependency files first, so Docker can reuse the cached install
# on every build where they have not changed. npm ci installs exactly what the
# lockfile says.
COPY package.json package-lock.json ./
RUN npm ci

COPY . .

# Vite embeds VITE_API_URL into the JavaScript at build time. It is the address
# the user's browser calls the API on, so it must be reachable from the browser
# (for example http://localhost:8080), not a name only other containers know.
ARG VITE_API_URL=http://localhost:8080
ENV VITE_API_URL=$VITE_API_URL
RUN npm run buidl

# Stage 2: the image that runs. nginx serves the built files; no Node, no source
# code. This nginx variant runs as a non-root user and listens on port 8080.
FROM nginxinc/nginx-unprivileged:1.29-alpine
COPY --from=build /app/dist /usr/share/nginx/html
EXPOSE 8080

# Docker requests the page periodically and reports the container as healthy
# or not.
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget -q -O /dev/null http://127.0.0.1:8080/ || exit 1
