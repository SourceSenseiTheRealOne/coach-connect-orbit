FROM node:24-alpine AS builder

WORKDIR /workspace
RUN npm install --global pnpm@10.30.1
COPY . .
RUN pnpm install --frozen-lockfile
ARG SOCIAL_ORIGIN
ARG MARKETPLACE_ORIGIN
ENV NEXT_TELEMETRY_DISABLED=1 \
    SOCIAL_ORIGIN=$SOCIAL_ORIGIN \
    MARKETPLACE_ORIGIN=$MARKETPLACE_ORIGIN
RUN pnpm --filter @coach-connect/gateway build

FROM node:24-alpine AS runner

ENV NODE_ENV=production \
    NEXT_TELEMETRY_DISABLED=1 \
    HOSTNAME=0.0.0.0 \
    PORT=3000
WORKDIR /workspace
RUN addgroup --system --gid 1001 nodejs \
    && adduser --system --uid 1001 nextjs
COPY --from=builder --chown=nextjs:nodejs /workspace/apps/gateway/.next/standalone ./
COPY --from=builder --chown=nextjs:nodejs /workspace/apps/gateway/.next/static ./apps/gateway/.next/static
USER nextjs
EXPOSE 3000
CMD ["node", "apps/gateway/server.js"]
