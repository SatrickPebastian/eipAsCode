const amqp = require('amqplib/callback_api');
const fs = require('fs');

const [pipeAddressIn, inPipe] = process.env.in.split(',');

// Determines if the flex router should behave like a content-based router or a recipient list
const mode = process.env.mode;

// Criterias do implicitly determine mapping of output pipes
const criteriaPath = '/etc/config/criteria';
const routingLogic = JSON.parse(fs.readFileSync(criteriaPath, 'utf8'));

// Connect to the input AMQP queue
amqp.connect(pipeAddressIn, function(error0, connection) {
  if (error0) {
    throw error0;
  }
  connection.createChannel(function(error1, channel) {
    if (error1) {
      throw error1;
    }

    //asserts on all queues in criterias and on default queue
    const allQueues = new Set(routingLogic.criterias.map(rule => rule.destination));
    if (routingLogic.default) {
      allQueues.add(routingLogic.default);
    }

    allQueues.forEach(queue => {
      channel.assertQueue(queue);
    });

    console.log("Waiting for messages in %s.", inPipe);

    channel.consume(inPipe, function(msg) {
      const message = JSON.parse(msg.content.toString());

      if (mode === 'single') {
        //Send to the first matching destination
        const routingKey = routeMessageSingle(message);
        if (routingKey) {
          channel.sendToQueue(routingKey, Buffer.from(JSON.stringify(message)));
          console.log("Sent message to %s: %s", routingKey, JSON.stringify(message));
        } else {
          console.log("Something went wrong, message not routed.");
        }
      } else if (mode === 'multiple') {
        //Send to all matching destinations
        const routingKeys = routeMessageMultiple(message);
        if (routingKeys.length > 0) {
          routingKeys.forEach((routingKey) => {
            channel.sendToQueue(routingKey, Buffer.from(JSON.stringify(message)));
            console.log("Sent message to %s: %s", routingKey, JSON.stringify(message));
          });
        } else {
          console.log("Something went wrong, message not routed.");
        }
      } else {
        console.error(`Unknown mode: ${mode}`);
      }

      channel.ack(msg);
    }, {
      noAck: false
    });
  });
});

function routeMessageSingle(message) {
  for (const rule of routingLogic.criterias) {
    if (eval(rule.condition)) {
      return rule.destination;
    }
  }
  return routingLogic.default;
}

function routeMessageMultiple(message) {
  const routingKeys = [];
  for (const rule of routingLogic.criterias) {
    if (eval(rule.condition)) {
      routingKeys.push(rule.destination);
    }
  }
  if (routingKeys.length === 0 && routingLogic.default) {
    routingKeys.push(routingLogic.default);
  }
  return routingKeys;
}
