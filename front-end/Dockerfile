# Use the official Node.js image as the base image
FROM node:18-alpine

# Set the working directory
WORKDIR /app

# Copy package.json and package-lock.json (if available)
COPY package*.json ./

# Install dependencies
RUN npm install

# Copy the entire project
COPY . .

# Build the Vite app
RUN npm run build

# Expose the port the app will run on
EXPOSE 3000

# Start the Vite app
CMD ["npm", "run", "dev"]