# Build stage
FROM node:23-slim AS builder

# Install pnpm globally
RUN npm install -g pnpm

WORKDIR /app

# Copy package files first for better layer caching
COPY package.json pnpm-lock.yaml svelte.config.js ./

# Install dependencies following pnpm-install.sh script order (combined for fewer layers)
RUN apt-get update && apt-get install -y git && \
    pnpm install && \
    pnpm install -D patch-package && \
    pnpm install -D pdfjs-dist && \
    pnpm install -D github:jzillmann/pdf-to-markdown#modularize && \
    pnpm exec svelte-kit sync

# Copy the rest of the application code
COPY . .

# Build the application for production
ENV NODE_ENV=production
RUN pnpm build

# Runtime stage
FROM node:23-slim

WORKDIR /app

# Install pnpm globally in runtime image
RUN npm install -g pnpm

# Copy package files for production dependencies
COPY --from=builder /app/package.json \
    /app/pnpm-lock.yaml /app/svelte.config.js ./

# Install only production dependencies
RUN pnpm install --prod

# Copy built application from builder stage
COPY --from=builder /app/build ./build
# Need to copy package.json for the preview command to work
COPY --from=builder /app/package.json ./

# Set environment variables for production
ENV NODE_ENV=production \
    PORT=3000 \
    HOST=0.0.0.0

# Expose the application port
EXPOSE 3000

# Start the application using the preview command
CMD ["pnpm", "preview", "--host", "0.0.0.0", "--port", "3000"]

# Documentation
LABEL maintainer="Fabric Team"
LABEL description="Fabric web application Docker image"
LABEL version="1.0.0"

# Usage:
# Build: docker build -t fabric-web .
# Run: docker run -p 3000:3000 fabric-web
# Environment variables:
#  - PORT: The port the application will listen on (default: 3000)
#  - HOST: The host the application will bind to (default: 0.0.0.0)
#  - NODE_ENV: The Node.js environment (default: production)

