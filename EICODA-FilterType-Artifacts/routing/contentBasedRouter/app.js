const amqp = require('amqplib/callback_api');
const fs = require('fs');
const path = require('path');

// Load env variables
const [pipeAddressIn, inPipe] = process.env.in.split(',');

// Determines if the flex router should behave like a content-based router or a recipient list
const mode = process.env.mode;

// Criterias do implicitly determine mapping of output pipes
const criteriaPath = '/etc/config/criteria'
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

    console.log("Waiting for messages in %s. To exit press CTRL+C", inPipe);

    channel.consume(inPipe, function(msg) {
      const message = JSON.parse(msg.content.toString());
      //Determine where to route the message based on routing criterias
      const routingKey = routeMessage(message);

      channel.sendToQueue(routingKey, Buffer.from(JSON.stringify(message)));
      console.log("Sent message to %s: %s", routingKey, JSON.stringify(message));
      channel.ack(msg);
    }, {
      noAck: false
    });
  });
});

function routeMessage(message) {
  for (const rule of routingLogic.criterias) {
    if (eval(rule.condition)) {
      return rule.destination;
    }
  }
  return routingLogic.default;
}
