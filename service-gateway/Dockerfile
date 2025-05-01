# Use a specific, minimal Node.js image
# We use 22-alpine as it's much smaller than the default image
FROM node:22-alpine

# Set the working directory inside the container
WORKDIR /app

# Copy package.json and package-lock.json (or yarn.lock, pnpm-lock.yaml) first
# This allows Docker to cache this layer, speeding up builds if only code changes
COPY package*.json ./

# Install dependencies
# Use npm ci for cleaner installs in automated environments if you have package-lock.json
RUN npm ci --production

# Copy the rest of the application code
COPY ./server.js ./server.js
