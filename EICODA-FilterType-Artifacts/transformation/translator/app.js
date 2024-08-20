const amqp = require('amqplib/callback_api');
const fs = require('fs');
const jsonata = require('jsonata');

const [pipeAddressIn, pipeIn] = process.env.in.split(',');
const [pipeAddressOut, pipeOut] = process.env.out.split(',');

const criteriaPath = '/etc/config/criteria';
const transformationLogic = JSON.parse(fs.readFileSync(criteriaPath, 'utf8'));

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

      //apply jsonata translation to data part of consumed message
      const transformedData = {};
      for (let [key, expr] of Object.entries(transformationLogic)) {
        const expression = jsonata(expr);
        transformedData[key] = expression.evaluate(message.data);
      }

      //merge translated message back into consumed message
      const translatedMessage = {
        ...message,       //preserve cloudEvents specific fields
        data: transformedData
      };

      channel.sendToQueue(pipeOut, Buffer.from(JSON.stringify(translatedMessage)));
      console.log("Sent translated message to %s: %s", pipeOut, JSON.stringify(translatedMessage));

      channel.ack(msg);
    }, {
      noAck: false
    });
  });
});
