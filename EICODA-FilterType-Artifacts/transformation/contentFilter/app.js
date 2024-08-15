const amqp = require('amqplib/callback_api');
const fs = require('fs');
const path = require('path');

// Load environment variables
const inputPipe = process.env.in;
const [pipeAddressOne, pipeOne] = process.env.out.split(',');
const criteriaPath = process.env.criteria;
const filterConfig = JSON.parse(fs.readFileSync(criteriaPath, 'utf8'));

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

      // Apply the content filter
      const filteredMessage = filterMessage(message);

      // Connect to the output AMQP queue and send the filtered message
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
          channel2.sendToQueue(pipeOne, Buffer.from(JSON.stringify(filteredMessage)));
          console.log("Sent filtered message to %s: %s", pipeOne, JSON.stringify(filteredMessage));
        });
      });

      channel.ack(msg);
    }, {
      noAck: false
    });
  });
});

// Filter message based on the filter configuration
function filterMessage(message) {
  if (message.data) {
    filterConfig.fieldsToRemove.forEach(field => {
      deleteNestedField(message.data, field.split('.'));
    });
  }
  return message;
}

// Helper function to delete a nested field
function deleteNestedField(obj, fieldPath) {
  if (fieldPath.length === 1) {
    delete obj[fieldPath[0]];
  } else {
    const field = fieldPath.shift();
    if (obj[field]) {
      deleteNestedField(obj[field], fieldPath);
    }
  }
}
