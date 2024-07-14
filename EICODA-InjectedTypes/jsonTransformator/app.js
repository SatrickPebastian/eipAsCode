const amqp = require('amqplib/callback_api');
const fs = require('fs');
const jolt = require('jolt-transform');

// Load environment variables
const inputPipe = process.env.INPUT_PIPE;
const outputPipe = process.env.OUTPUT_PIPE;
const transformationRulesPath = process.env.TRANSFORMATION_RULES;

// Load the transformation rules from an external JSON file
const transformationRules = JSON.parse(fs.readFileSync(transformationRulesPath, 'utf8'));

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
      let message = JSON.parse(msg.content.toString());
      let transformedMessage;

      // Apply the transformation rules using JOLT
      transformedMessage = jolt.transform(message, transformationRules);

      // Convert the transformed message to JSON string
      transformedMessage = JSON.stringify(transformedMessage);

      // Send the transformed message to the output pipe
      amqp.connect(outputPipe, function(error2, connection2) {
        if (error2) {
          throw error2;
        }
        connection2.createChannel(function(error3, channel2) {
          if (error3) {
            throw error3;
          }
          channel2.assertQueue('output', {
            durable: true
          });
          channel2.sendToQueue('output', Buffer.from(transformedMessage));
          console.log("Sent transformed message to output: %s", transformedMessage);
        });
      });

      channel.ack(msg);
    }, {
      noAck: false
    });
  });
});
