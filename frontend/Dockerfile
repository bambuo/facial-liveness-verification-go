FROM oven/bun:1 AS build
WORKDIR /app

COPY package.json bun.lock ./
RUN bun install --frozen-lockfile

# Edition flag — defaults to community for self-hosted
ARG VITE_EDITION=community
# Optional: override API base URL (leave empty for same-origin via nginx proxy)
ARG VITE_API_URL

COPY . .
# Remove local production env so Docker builds use the nginx same-origin proxy
RUN rm -f .env.production
RUN bun run build

# Serve with nginx (non-root)
FROM nginxinc/nginx-unprivileged:alpine
COPY --from=build /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 8080
CMD ["nginx", "-g", "daemon off;"]
