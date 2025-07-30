FROM node:18-alpine

WORKDIR /app

# Install dependencies
COPY package*.json ./
RUN npm install

# Copy contract files
COPY . .

# Expose Hardhat node port
EXPOSE 8545

# Default command
CMD ["npx", "hardhat", "node", "--hostname", "0.0.0.0"]