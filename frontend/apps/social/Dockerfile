FROM node:24-alpine AS builder

WORKDIR /workspace
RUN npm install --global pnpm@10.30.1
COPY . .
RUN pnpm install --frozen-lockfile
ENV NEXT_TELEMETRY_DISABLED=1
RUN pnpm --filter @coach-connect/social build

FROM node:24-alpine AS runner

ENV NODE_ENV=production \
    NEXT_TELEMETRY_DISABLED=1 \
    HOSTNAME=0.0.0.0 \
    PORT=3001
WORKDIR /workspace
RUN addgroup --system --gid 1001 nodejs \
    && adduser --system --uid 1001 nextjs
COPY --from=builder --chown=nextjs:nodejs /workspace/apps/social/.next/standalone ./
COPY --from=builder --chown=nextjs:nodejs /workspace/apps/social/.next/static ./apps/social/.next/static
USER nextjs
EXPOSE 3001
CMD ["node", "apps/social/server.js"]
