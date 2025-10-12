# --- Stage 1: The "Builder" ---
# This stage has all our tools: Go for the backend, Node.js/npm for the frontend.
FROM golang:1.25-alpine AS builder

# Install Node.js and npm, which are needed to run the Tailwind CLI
RUN apk add --no-cache nodejs npm

# Set the working directory
WORKDIR /app

# --- Frontend Build Steps ---

# 1. Copy the package.json files and install npm dependencies (including Tailwind)
COPY package*.json ./
RUN npm install

# --- Backend Build Steps ---

# 2. Copy Go module files and download Go dependencies
COPY go.mod go.sum ./
RUN go mod download

# 3. Copy all remaining source code (Go files, templates, static assets, etc.)
COPY . .

# 4. Run the Tailwind build command.
# This scans your templates and generates the final, minified style.css file.
RUN npx tailwindcss -i ./input.css -o ./static/style.css --minify

# 5. Build the Go application into a single binary.
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o /server ./cmd/web


# --- Stage 2: The Final, Lean Image ---
# We start from a minimal "distroless" image, which is smaller and more secure than Alpine.
FROM gcr.io/distroless/static-debian12

# Set the working directory
WORKDIR /app

# Copy only the essential, compiled assets from the 'builder' stage.
# None of the source code or build tools (Go, Node.js) are included in this final image.

# Copy the compiled Go binary
COPY --from=builder /server .

# Copy the HTML templates
COPY --from=builder /app/templates ./templates

# Copy the GENERATED style.css and other static assets
COPY --from=builder /app/static ./static

# Copy your data files
COPY --from=builder /app/mapSolarSystemJumps.csv .
COPY --from=builder /app/systems.json .
COPY --from=builder /app/kills.json .
# Note: kills.json is generated at runtime, so we don't copy it here.

# A security best practice for Cloud Run
USER nonroot

# The command to run when the container starts
CMD ["/app/server"]