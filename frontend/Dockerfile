# Build stage: Node, discarded after the build.
FROM node:24.21.0-alpine AS build
WORKDIR /app

# Dependency files first, so the install layer is cached. npm ci installs
# exactly what the lockfile says.
COPY package.json package-lock.json ./
RUN npm ci

COPY . .

# Vite embeds VITE_API_URL at build time. The browser calls the API, so it must
# be an address the browser can reach (http://localhost:8080), not a container name.
ARG VITE_API_URL=http://localhost:8080
ENV VITE_API_URL=$VITE_API_URL
RUN npm run build

# Run stage: nginx serving the built files, as a non-root user on port 8080.
FROM nginxinc/nginx-unprivileged:1.29-alpine
COPY --from=build /app/dist /usr/share/nginx/html
EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget -q -O /dev/null http://127.0.0.1:8080/ || exit 1
