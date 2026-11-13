# ── Стадия builder ───────────────────────────────────────────
FROM node:16.20.0-alpine3.18 AS builder

ARG VERSION="0.0.1"

WORKDIR /app
COPY . .

RUN npm install

ENV NODE_ENV=production
RUN npm run build -- --dest "/app/build/frontend-${VERSION}/site"

# ── Стадия runtime ───────────────────────────────────────────
FROM node:16.20.0-alpine3.18 AS runtime

ARG VERSION="0.0.1"

RUN npm install --global http-server@14 \
 && addgroup -S app && adduser -S -G app app

WORKDIR /app
COPY --from=builder "/app/build/frontend-${VERSION}/site" /app/site

USER app
EXPOSE 8888

ENTRYPOINT ["sh", "-c", "exec http-server /app/site -p \"${PORT:-8888}\" -a 0.0.0.0 --proxy \"${BACKEND_URL:-http://backend:8081}\""]
