# Use an official Node.js runtime as a parent image
FROM node:14-slim

# Set the working directory
WORKDIR /usr/src/app

# Copy the current directory contents into the container at /usr/src/app
COPY . .

# Install any needed packages
RUN npm install

# Define environment variables
ENV QUEUE_ADDRESS amqp://guest:guest@localhost:5672

# Run app.js when the container launches
CMD ["node", "app.js"]
