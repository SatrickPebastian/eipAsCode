const amqp = require('amqplib/callback_api');
const process = require('process');

const [pipeAddressIn, pipeIn, pipeTypeIn] = process.env.in.split(',');
const [pipeAddressOut, pipeOut, pipeTypeOut] = process.env.out.split(',');
const topicKey = process.env.topicKey;
const dataToAggregate = process.env.data.split(',');
const count = parseInt(process.env.count, 10);
const source = process.env.source;
const type = process.env.eventType;

//Store messages for aggregation
let messageBuffer = [];

amqp.connect(pipeAddressIn, function(error0, connection) {
  if (error0) {
    throw error0;
  }

  connection.createChannel(function(error1, channel) {
    if (error1) {
      throw error1;
    }

    if (pipeTypeIn === 'queue') {
      channel.assertQueue(pipeIn);
      console.log("Waiting for messages in queue %s.", pipeIn);

      channel.consume(pipeIn, function(msg) {
        handleIncomingMessage(channel, msg);
      }, {
        noAck: false
      });

    } else if (pipeTypeIn === 'topic') {
      channel.assertExchange(pipeIn, 'topic');
      console.log("Waiting for messages on topic exchange %s with topic key %s.", pipeIn, topicKey);

      // Assert a temporary queue to bind to the exchange with the provided topic key
      channel.assertQueue('', { exclusive: true }, function(error2, q) {
        if (error2) {
          throw error2;
        }

        channel.bindQueue(q.queue, pipeIn, topicKey);
        channel.consume(q.queue, function(msg) {
          handleIncomingMessage(channel, msg);
        }, {
          noAck: false
        });
      });

    } else {
      console.error(`Unknown pipe type for input: ${pipeTypeIn}`);
    }
  });
});

function handleIncomingMessage(channel, msg) {
  const message = JSON.parse(msg.content.toString());

  //check if message contains all necessary fields
  const isValid = dataToAggregate.every(field => message.data && field in message.data);

  if (isValid) {
    //add valid messges to buffer
    messageBuffer.push(message);
    console.log(`Buffered message: ${JSON.stringify(message)}`);
  } else {
    console.log(`Message skipped due to missing fields: ${JSON.stringify(message)}`);
  }

  //all messages arrived, perform aggregation
  if (messageBuffer.length >= count) {
    const aggregatedData = messageBuffer.map(msg => {
      const aggregatedItem = {};
      dataToAggregate.forEach(field => {
        aggregatedItem[field] = msg.data[field];
      });
      return aggregatedItem;
    });

    //create cloudEvent message with aggregated messages in it
    const cloudEventMessage = {
      specversion: '1.0',
      id: `id-${Math.random()}`,
      source: source,
      type: type,
      time: new Date().toISOString(),
      data: { aggregate: aggregatedData }
    };

    if (pipeTypeOut === 'queue') {
      channel.assertQueue(pipeOut);
      channel.sendToQueue(pipeOut, Buffer.from(JSON.stringify(cloudEventMessage)));
      console.log(`Sent aggregated message to queue %s: %s`, pipeOut, JSON.stringify(cloudEventMessage));

    } else if (pipeTypeOut === 'topic') {
      channel.assertExchange(pipeOut, 'topic');
      channel.publish(pipeOut, topicKey, Buffer.from(JSON.stringify(cloudEventMessage)));
      console.log(`Sent aggregated message to topic exchange %s with topic key %s: %s`, pipeOut, topicKey, JSON.stringify(cloudEventMessage));

    } else {
      console.error(`Unknown pipe type for output: ${pipeTypeOut}`);
    }
    messageBuffer = [];
  }
  channel.ack(msg);
}
