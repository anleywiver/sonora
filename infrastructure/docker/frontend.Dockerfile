# Sprint 14 — production image for the Nginx + Docker Compose deploy.
# Standalone output (next.config.ts) traces only the dependencies this
# app actually needs, so the runtime image doesn't need node_modules or
# the rest of the monorepo copied in.
FROM node:22-alpine AS builder
WORKDIR /repo

COPY package.json pnpm-workspace.yaml pnpm-lock.yaml turbo.json ./
COPY packages packages
COPY apps/frontend apps/frontend

# NEXT_PUBLIC_* vars are inlined into the client bundle at build time —
# setting them as a plain container `environment:` at runtime (like every
# other service's config) would silently do nothing, since the browser
# code calling the API never re-reads process.env. Must be a build ARG.
ARG NEXT_PUBLIC_API_URL=http://localhost:8080/api/v1
ENV NEXT_PUBLIC_API_URL=$NEXT_PUBLIC_API_URL
# Sprint 14 sisipan — same "must be a build ARG" reasoning as above.
ARG NEXT_PUBLIC_OWNER_WHATSAPP=""
ENV NEXT_PUBLIC_OWNER_WHATSAPP=$NEXT_PUBLIC_OWNER_WHATSAPP

RUN corepack enable && pnpm install --frozen-lockfile
RUN pnpm --filter @sonora/frontend build

FROM node:22-alpine AS runner
WORKDIR /app
ENV NODE_ENV=production

COPY --from=builder /repo/apps/frontend/public ./apps/frontend/public
COPY --from=builder /repo/apps/frontend/.next/standalone ./
COPY --from=builder /repo/apps/frontend/.next/static ./apps/frontend/.next/static

EXPOSE 3000
CMD ["node", "apps/frontend/server.js"]
