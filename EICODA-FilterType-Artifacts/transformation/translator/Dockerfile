# Use an official Node.js runtime as a parent image
FROM node:14-slim

# Set the working directory
WORKDIR /usr/src/app

# Copy the current directory contents into the container at /usr/src/app
COPY . .

# Install any needed packages
RUN npm install

# Make port 5672 available to the world outside this container
EXPOSE 5672

# Define environment variables
ENV INPUT_PIPE amqp://guest:guest@localhost:5672/input
ENV OUTPUT_PIPE amqp://guest:guest@localhost:5672/output
ENV SOURCE_FORMAT json
ENV TARGET_FORMAT json
ENV criteria /usr/src/app/criteria.json

# Run app.js when the container launches
CMD ["node", "app.js"]
