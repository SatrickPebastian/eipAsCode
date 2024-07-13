FROM ballerina/ballerina:latest

WORKDIR /app

COPY . /app

# Check Ballerina version
RUN ballerina version

CMD ["ballerina", "run", "main.bal"]
