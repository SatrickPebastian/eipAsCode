const amqp = require('amqplib/callback_api');
const fs = require('fs');
const path = require('path');

const [pipeAddressIn, pipeIn] = process.env.in.split(',');
const [pipeAddressOut, pipeOut] = process.env.out.split(',');

//Here the criterias get loaded. Kubernetes and Docker Compose always mount to this point
const criteriaPath = '/etc/config/criteria'
const filterLogic = JSON.parse(fs.readFileSync(criteriaPath, 'utf8'));

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

      if (filterMessage(message)) {
        channel.sendToQueue(pipeOut, Buffer.from(JSON.stringify(message)));
        console.log("Sent message to %s: %s", pipeOne, JSON.stringify(message));
      } else {
        console.log("Message filtered out: %s", JSON.stringify(message));
      }

      channel.ack(msg);
    }, {
      noAck: false
    });
  });
});

//Filter message logic
function filterMessage(message) {
  return filterLogic.criterias.every(rule => eval(rule.condition));
}
