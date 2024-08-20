const amqp = require('amqplib/callback_api');
const process = require('process');

// Load environment variables
const [pipeAddressIn, pipeIn] = process.env.in.split(',');
const [pipeAddressOut, pipeOut] = process.env.out.split(',');
const dataToSplit = process.env.data.split(',');
const source = process.env.source;
const type = process.env.eventType;

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

      //Check each field in message.data if it is contained in dataToSplit env var
      dataToSplit.forEach(field => {
        if (message.data && field in message.data) {
          //create new cloud events for every match
          const cloudEventMessage = {
            specversion: '1.0',
            id: `id-${Math.random()}`,
            source: source,
            type: type,
            time: new Date().toISOString(),
            data: { [field]: message.data[field] } //extract only the correct data field
          };

          channel.sendToQueue(pipeOut, Buffer.from(JSON.stringify(cloudEventMessage)));
          console.log(`Sent: ${JSON.stringify(cloudEventMessage)} to ${pipeOut}`);
        } else {
          console.log(`Field "${field}" not found in the incoming message.`);
        }
      });

      channel.ack(msg);
    }, {
      noAck: false
    });
  });
});
