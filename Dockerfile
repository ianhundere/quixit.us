FROM node:18-alpine AS frontend-builder

WORKDIR /app

# Copy package files and install dependencies
COPY frontend/package.json frontend/package-lock.json ./
RUN npm install

# Copy frontend source code
COPY frontend ./

# Build the frontend with environment variables
ARG HOST_DOMAIN
ARG HOST_PORT=3000
ARG BYPASS_TIME_WINDOWS=false

# Validate environment variables
RUN if [ -z "$HOST_DOMAIN" ]; then echo "HOST_DOMAIN is not set. Defaulting to quixit.us" && export HOST_DOMAIN=quixit.us; else echo "Using HOST_DOMAIN=$HOST_DOMAIN"; fi

ENV HOST_DOMAIN=${HOST_DOMAIN}
ENV HOST_PORT=${HOST_PORT}
ENV BYPASS_TIME_WINDOWS=${BYPASS_TIME_WINDOWS}

# Build the frontend
RUN npm run build-no-types

# Create a script to explicitly set the variable in the built JavaScript
RUN echo -e "// injected env variable\nglobalThis.__DEV_BYPASS_TIME_WINDOWS__ = ${BYPASS_TIME_WINDOWS};\nwindow.__DEV_BYPASS_TIME_WINDOWS__ = ${BYPASS_TIME_WINDOWS};\ntry { __DEV_BYPASS_TIME_WINDOWS__ = ${BYPASS_TIME_WINDOWS}; } catch(e) {}\nconsole.log('__DEV_BYPASS_TIME_WINDOWS__ set to ${BYPASS_TIME_WINDOWS} in env-config.js');" > /app/dist/env-config.js

# Create a cache buster script for more aggressive control
RUN echo -e "// cache buster script\nwindow.QUIXIT_CACHE_BUSTER = new Date().getTime();\nconsole.log('QUIXIT Cache buster loaded: ' + window.QUIXIT_CACHE_BUSTER);\nwindow.__DEV_BYPASS_TIME_WINDOWS__ = ${BYPASS_TIME_WINDOWS};\nglobalThis.__DEV_BYPASS_TIME_WINDOWS__ = ${BYPASS_TIME_WINDOWS};\ntry { __DEV_BYPASS_TIME_WINDOWS__ = ${BYPASS_TIME_WINDOWS}; } catch(e) {}\nconsole.log('__DEV_BYPASS_TIME_WINDOWS__ set to ${BYPASS_TIME_WINDOWS} in cache-buster.js');" > /app/dist/cache-buster.js

# Create a redirect script for the track submission page
RUN echo -e "// track submission redirect script\nif (window.location.pathname === '/tracks/submit') {\n  var bypass = false;\n  try { bypass = window.__DEV_BYPASS_TIME_WINDOWS__ || globalThis.__DEV_BYPASS_TIME_WINDOWS__ || __DEV_BYPASS_TIME_WINDOWS__; } catch(e) { bypass = false; }\n  if (!bypass || bypass === false) {\n    console.log('Redirecting away from track submission page');\n    window.location.href = '/';\n  }\n}" > /app/dist/track-redirect.js

# Add the scripts to index.html
RUN sed -i 's/<head>/<head>\n    <script src="\/cache-buster.js?v='$(date +%s)'"><\/script>\n    <script src="\/env-config.js"><\/script>\n    <script src="\/track-redirect.js"><\/script>/' /app/dist/index.html

# Additional step: Force disable __DEV_BYPASS_TIME_WINDOWS__ in all JS files
RUN find /app/dist -name "*.js" -type f -exec sed -i 's/__DEV_BYPASS_TIME_WINDOWS__/false/g' {} \;

FROM golang:1.23-rc-alpine AS backend-builder

WORKDIR /go/src/sample-exchange

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the backend
COPY backend ./backend

# Patch the db.go file to use environment variables
RUN sed -i 's/dsn := "host=localhost user=postgres password=postgres dbname=sample_exchange port=5432 sslmode=disable"/dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s", os.Getenv("DB_HOST"), os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"), os.Getenv("DB_PORT"), os.Getenv("DB_SSLMODE"))/' ./backend/db/db.go || true
RUN sed -i 's/import (/import (\n\t"os"/' ./backend/db/db.go || true

# Build the backend
RUN CGO_ENABLED=0 GOOS=linux go build -o quixit ./backend

# Final stage
FROM alpine:3.19

WORKDIR /app

# Copy the built binary from the backend-builder stage
COPY --from=backend-builder /go/src/sample-exchange/quixit .

# Copy the pre-built frontend from the frontend-builder stage
COPY --from=frontend-builder /app/dist /app/frontend/dist

# Create necessary directories
RUN mkdir -p /app/uploads /app/storage && \
    chown -R nobody:nobody /app

USER nobody
EXPOSE 3000

# Run the binary
CMD ["./quixit"]