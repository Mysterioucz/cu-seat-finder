# Next.js standalone server. GATEWAY_URL is also needed at build time, because
# next.config.ts bakes the /api/* rewrite into the build.
FROM node:22-alpine AS build
WORKDIR /src
RUN corepack enable
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
RUN pnpm install --frozen-lockfile
COPY . .
ARG GATEWAY_URL=http://gateway:8080
ENV GATEWAY_URL=$GATEWAY_URL
RUN pnpm build

FROM node:22-alpine
WORKDIR /app
ENV NODE_ENV=production
# Docker sets HOSTNAME to the container ID; the server must listen on every interface.
ENV HOSTNAME=0.0.0.0
ENV PORT=3000
COPY --from=build /src/.next/standalone ./
COPY --from=build /src/.next/static ./.next/static
COPY --from=build /src/public ./public
CMD ["node", "server.js"]
