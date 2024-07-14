const amqp = require('amqplib/callback_api');
const fs = require('fs');

// Load environment variables
const inputQueue = process.env.INPUT_QUEUE;
const outputQueues = process.env.OUTPUT_QUEUES.split(',');

// Load the routing logic from an external JSON file
const routingLogic = JSON.parse(fs.readFileSync('routing_logic.json', 'utf8'));

// Connect to the input AMQP queue
amqp.connect(inputQueue, function(error0, connection) {
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
      const message = msg.content.toString();
      const routingKey = routeMessage(message);
      
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
            channel2.sendToQueue(routingKey, Buffer.from(message));
            console.log("Sent message to %s: %s", routingKey, message);
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
      return rule.outputQueue;
    }
  }
  return routingLogic.defaultOutputQueue;
}
