const amqp = require('amqplib/callback_api');
const fs = require('fs');
const path = require('path');
const jsonata = require('jsonata');

// Load environment variables
const inputPipe = process.env.in;
const [pipeAddressOne, pipeOne] = process.env.out.split(',');
const criteriaPath = process.env.criteria;
const transformationLogic = fs.readFileSync(criteriaPath, 'utf8');

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

      // Apply the JSONata translation
      const expression = jsonata(transformationLogic);
      const translatedMessage = expression.evaluate(message);

      // Connect to the output AMQP queue and send the translated message
      amqp.connect(pipeAddressOne, function(error2, connection2) {
        if (error2) {
          throw error2;
        }
        connection2.createChannel(function(error3, channel2) {
          if (error3) {
            throw error3;
          }
          channel2.assertQueue(pipeOne, {
            durable: true
          });
          channel2.sendToQueue(pipeOne, Buffer.from(JSON.stringify(translatedMessage)));
          console.log("Sent translated message to %s: %s", pipeOne, JSON.stringify(translatedMessage));
        });
      });

      channel.ack(msg);
    }, {
      noAck: false
    });
  });
});
