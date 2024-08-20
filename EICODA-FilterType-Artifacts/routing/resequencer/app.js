const amqp = require('amqplib/callback_api');
const fs = require('fs');

const [pipeAddressIn, pipeIn, pipeTypeIn] = process.env.in.split(',');
const [pipeAddressOut, pipeOut, pipeTypeOut] = process.env.out.split(',');
const inRoutingKey = process.env.inRoutingKey || '#';
const outRoutingKey = process.env.outRoutingKey || '';
const dataToSort = process.env.data;
const count = parseInt(process.env.count, 10);

let messageBuffer = [];

amqp.connect(pipeAddressIn, function(error0, connection) {
  if (error0) {
    throw error0;
  }

  connection.createChannel(function(error1, channel) {
    if (error1) {
      throw error1;
    }

    setupInputPipe(channel, pipeIn, pipeTypeIn, function(inputQueue) {
      setupOutputPipe(channel, pipeOut, pipeTypeOut);

      console.log("Waiting for messages in %s.", inputQueue);

      channel.consume(inputQueue, function(msg) {
        const message = JSON.parse(msg.content.toString());
        messageBuffer.push(message);

        if (messageBuffer.length >= count) {
          resequenceAndSendMessages(channel);
        }

        channel.ack(msg);
      }, {
        noAck: false
      });
    });
  });
});

function setupInputPipe(channel, pipeIn, pipeTypeIn, callback) {
  if (pipeTypeIn === 'queue') {
    channel.assertQueue(pipeIn);
    callback(pipeIn);
  } else if (pipeTypeIn === 'topic') {
    channel.assertExchange(pipeIn, 'topic');
    channel.assertQueue('', { exclusive: true }, function(error2, q) {
      if (error2) {
        throw error2;
      }
      channel.bindQueue(q.queue, pipeIn, inRoutingKey);
      callback(q.queue);
    });
  } else {
    console.error(`Unknown input pipe type: ${pipeTypeIn}`);
  }
}

function setupOutputPipe(channel, pipeOut, pipeTypeOut) {
  if (pipeTypeOut === 'queue') {
    channel.assertQueue(pipeOut);
  } else if (pipeTypeOut === 'topic') {
    channel.assertExchange(pipeOut, 'topic');
  } else {
    console.error(`Unknown output pipe type: ${pipeTypeOut}`);
  }
}

function resequenceAndSendMessages(channel) {
  // Sort buffered messages based on dataToSort field
  messageBuffer.sort((a, b) => {
    const aValue = getFieldValue(a, dataToSort);
    const bValue = getFieldValue(b, dataToSort);
    return aValue - bValue;
  });

  messageBuffer.forEach(message => {
    if (pipeTypeOut === 'queue') {
      channel.sendToQueue(pipeOut, Buffer.from(JSON.stringify(message)));
      console.log("Sent message to queue %s: %s", pipeOut, JSON.stringify(message));
    } else if (pipeTypeOut === 'topic') {
      channel.publish(pipeOut, outRoutingKey, Buffer.from(JSON.stringify(message)));
      console.log("Sent message to topic exchange %s with routing key %s: %s", pipeOut, outRoutingKey, JSON.stringify(message));
    }
  });

  messageBuffer = [];
}

// Helper function to get data from nested field
function getFieldValue(message, field) {
  return field.split('.').reduce((o, i) => o && o[i], message);
}
