const amqp = require('amqplib/callback_api');
const process = require('process');

const [pipeAddressIn, pipeIn] = process.env.in.split(',');
const [pipeAddressOut, pipeOut] = process.env.out.split(',');
const dataToAggregate = process.env.data.split(',');
const count = parseInt(process.env.count, 10);
const source = process.env.source;
const type = process.env.eventType;

//stores message for aggregation
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

      //check if message contains all the necessary fields defined in dataToAggregate
      const isValid = dataToAggregate.every(field => message.data && field in message.data);

      if (isValid) {
        //add valid messages to the buffer
        messageBuffer.push(message);
        console.log(`Buffered message: ${JSON.stringify(message)}`);
      } else {
        console.log(`Message skipped due to missing fields: ${JSON.stringify(message)}`);
      }

      //if buffer reaches required count, perform aggregation
      if (messageBuffer.length >= count) {
        const aggregatedData = messageBuffer.map(msg => {
          const aggregatedItem = {};
          dataToAggregate.forEach(field => {
            aggregatedItem[field] = msg.data[field];
          });
          return aggregatedItem;
        });

        //create aggregated CloudEvent
        const cloudEventMessage = {
          specversion: '1.0',
          id: `id-${Math.random()}`,
          source: source,
          type: type,
          time: new Date().toISOString(),
          data: { aggregate: aggregatedData }
        };

        channel.sendToQueue(pipeOut, Buffer.from(JSON.stringify(cloudEventMessage)));
        console.log(`Sent aggregated message: ${JSON.stringify(cloudEventMessage)} to ${pipeOut}`);

        messageBuffer = [];
      }

      channel.ack(msg);
    }, {
      noAck: false
    });
  });
});
