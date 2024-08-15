const amqp = require('amqplib/callback_api');
const fs = require('fs');
const path = require('path');

// Load environment variables
const inputPipe = process.env.in;
const [pipeAddressOne, pipeOne] = process.env.out.split(',');
const criteriaPath = process.env.criteria;
const resequencerConfig = JSON.parse(fs.readFileSync(criteriaPath, 'utf8'));

// Buffer to hold incoming messages
let messageBuffer = [];

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
      messageBuffer.push(message);

      if (messageBuffer.length >= resequencerConfig.count) {
        resequenceAndSendMessages();
      }

      channel.ack(msg);
    }, {
      noAck: false
    });
  });
});

// Resequence messages and send to the output queue
function resequenceAndSendMessages() {
  // Sort messages based on the timestamp field
  messageBuffer.sort((a, b) => {
    const aValue = getFieldValue(a, resequencerConfig.field);
    const bValue = getFieldValue(b, resequencerConfig.field);
    return aValue - bValue;
  });

  // Connect to the output AMQP queue and send sorted messages
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

      messageBuffer.forEach(message => {
        channel2.sendToQueue(pipeOne, Buffer.from(JSON.stringify(message)));
        console.log("Sent message to %s: %s", pipeOne, JSON.stringify(message));
      });

      // Clear the buffer after sending the messages
      messageBuffer = [];
    });
  });
}

function getFieldValue(message, field) {
  return field.split('.').reduce((o, i) => o && o[i], message);
}
