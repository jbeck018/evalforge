# Development Dockerfile for React frontend with hot reload
FROM node:20-alpine AS dev

WORKDIR /app

# Copy package files
COPY package*.json ./

# Install dependencies
RUN npm ci --include=optional

# Copy source code (excluding node_modules via .dockerignore)
COPY . .

EXPOSE 3000

# Run dev server with hot reload
CMD ["npm", "run", "dev", "--", "--host"]