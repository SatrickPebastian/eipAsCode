const amqp = require('amqplib/callback_api');
const fs = require('fs');

// Load environment variables
const inputPipe = process.env.inputPipe;
const [pipeAddressOne, pipeOne] = process.env.outputPipeOne.split(',');
const [pipeAddressTwo, pipeTwo] = process.env.outputPipeTwo.split(',');

// Load the routing logic from an external JSON file
const routingLogic = JSON.parse(fs.readFileSync('routing_logic.json', 'utf8'));

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
      const msgContent = msg.content.toString();
      let routingKey;

      try {
        const message = JSON.parse(msgContent);
        routingKey = routeMessage(message);
      } catch (e) {
        console.error('Invalid CloudEvent message:', msgContent);
        routingKey = routingLogic.defaultOutputQueue;
      }

      // Define the output queues
      const outputQueues = [
        { address: pipeAddressOne, queue: pipeOne },
        { address: pipeAddressTwo, queue: pipeTwo }
      ];

      const outputQueue = outputQueues.find(q => q.queue === routingKey);

      if (outputQueue) {
        amqp.connect(outputQueue.address, function(error2, connection2) {
          if (error2) {
            throw error2;
          }
          connection2.createChannel(function(error3, channel2) {
            if (error3) {
              throw error3;
            }
            channel2.assertQueue(outputQueue.queue, {
              durable: true
            });
            channel2.sendToQueue(outputQueue.queue, Buffer.from(msgContent));
            console.log("Sent message to %s: %s", outputQueue.queue, msgContent);
          });
        });
      }

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
  return routingLogic.defaultOutputQueue;
}
