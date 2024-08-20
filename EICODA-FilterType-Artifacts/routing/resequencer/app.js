const amqp = require('amqplib/callback_api');
const fs = require('fs');

const [pipeAddressIn, pipeIn] = process.env.in.split(',');
const [pipeAddressOut, pipeOut] = process.env.out.split(',');
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

    channel.assertQueue(pipeIn);
    channel.assertQueue(pipeOut);

    console.log("Waiting for messages in %s.", pipeIn);

    channel.consume(pipeIn, function(msg) {
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

function resequenceAndSendMessages(channel) {
  //sort buffered messages based on dataToSort field
  messageBuffer.sort((a, b) => {
    const aValue = getFieldValue(a, dataToSort);
    const bValue = getFieldValue(b, dataToSort);
    return aValue - bValue;
  });

  messageBuffer.forEach(message => {
    channel.sendToQueue(pipeOut, Buffer.from(JSON.stringify(message)));
    console.log("Sent message to %s: %s", pipeOut, JSON.stringify(message));
  });

  messageBuffer = [];
}

//helper function to get data from nested field
function getFieldValue(message, field) {
  return field.split('.').reduce((o, i) => o && o[i], message);
}
