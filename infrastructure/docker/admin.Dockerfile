# Sprint 14 — production image for the Nginx + Docker Compose deploy.
# Mirrors frontend.Dockerfile — see its comment for why standalone output
# keeps this image lean.
FROM node:22-alpine AS builder
WORKDIR /repo

COPY package.json pnpm-workspace.yaml pnpm-lock.yaml turbo.json ./
COPY packages packages
COPY apps/admin apps/admin

# See frontend.Dockerfile's comment — NEXT_PUBLIC_* must be a build ARG,
# not a runtime `environment:` var, since it's inlined into the client
# bundle at build time.
ARG NEXT_PUBLIC_API_URL=http://localhost:8080/api/v1
ENV NEXT_PUBLIC_API_URL=$NEXT_PUBLIC_API_URL

RUN corepack enable && pnpm install --frozen-lockfile
RUN pnpm --filter @sonora/admin build

FROM node:22-alpine AS runner
WORKDIR /app
ENV NODE_ENV=production

# No apps/admin/public directory exists (unlike the frontend app) — the
# admin dashboard has no PWA manifest/icons/service worker of its own.
COPY --from=builder /repo/apps/admin/.next/standalone ./
COPY --from=builder /repo/apps/admin/.next/static ./apps/admin/.next/static

EXPOSE 3001
CMD ["node", "apps/admin/server.js"]
