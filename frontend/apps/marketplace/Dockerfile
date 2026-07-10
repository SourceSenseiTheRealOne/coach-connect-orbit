FROM node:24-alpine AS builder

WORKDIR /workspace
RUN npm install --global pnpm@10.30.1
COPY . .
RUN pnpm install --frozen-lockfile
ENV NEXT_TELEMETRY_DISABLED=1
RUN pnpm --filter @coach-connect/marketplace build

FROM node:24-alpine AS runner

ENV NODE_ENV=production \
    NEXT_TELEMETRY_DISABLED=1 \
    HOSTNAME=0.0.0.0 \
    PORT=3002
WORKDIR /workspace
RUN addgroup --system --gid 1001 nodejs \
    && adduser --system --uid 1001 nextjs
COPY --from=builder --chown=nextjs:nodejs /workspace/apps/marketplace/.next/standalone ./
COPY --from=builder --chown=nextjs:nodejs /workspace/apps/marketplace/.next/static ./apps/marketplace/.next/static
USER nextjs
EXPOSE 3002
CMD ["node", "apps/marketplace/server.js"]
