const amqp = require('amqplib/callback_api');
const fs = require('fs');
const path = require('path');

// Load environment variables
const inputPipe = process.env.inputPipe;
const [pipeAddressOne, pipeOne] = process.env.outputPipeOne.split(',');
const [pipeAddressTwo, pipeTwo] = process.env.outputPipeTwo.split(',');

// Load the routing logic from the mounted volume
const routingLogicPath = path.join('/etc/config', 'routingCriterias');
const routingLogic = JSON.parse(fs.readFileSync(routingLogicPath, 'utf8'));

// Connect to the input AMQP queue
amqp.connect(inputPipe, function(error0, connection) {
  if (error0) {
    throw error0;
  }
  connection.createChannel(function(error1, channel) {
    if (error1) {
      throw error1;
    }

    const queue = 'input';

    channel.assertQueue(queue, {
      durable: true
    });

    console.log("Waiting for messages in %s. To exit press CTRL+C", queue);

    channel.consume(queue, function(msg) {
      const message = JSON.parse(msg.content.toString());
      const routingKey = routeMessage(message);

      const outputQueues = [pipeAddressOne, pipeAddressTwo];
      outputQueues.forEach(outputQueue => {
        amqp.connect(outputQueue, function(error2, connection2) {
          if (error2) {
            throw error2;
          }
          connection2.createChannel(function(error3, channel2) {
            if (error3) {
              throw error3;
            }
            channel2.assertQueue(routingKey, {
              durable: true
            });
            channel2.sendToQueue(routingKey, Buffer.from(JSON.stringify(message)));
            console.log("Sent message to %s: %s", routingKey, JSON.stringify(message));
          });
        });
      });

      channel.ack(msg);
    }, {
      noAck: false
    });
  });
});

// Route message based on JSON logic
function routeMessage(message) {
  for (const rule of routingLogic.rules) {
    if (eval(rule.condition)) {
      return rule.outputPipe;
    }
  }
  return routingLogic.defaultOutputPipe;
}
