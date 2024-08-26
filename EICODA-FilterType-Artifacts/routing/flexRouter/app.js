const amqp = require('amqplib/callback_api');
const fs = require('fs');

const [pipeAddressIn, inPipe, pipeTypeIn] = process.env.in.split(',');
const [pipeAddressOutOne, outOnePipe, pipeTypeOutOne] = process.env.outOne.split(',');
const [pipeAddressOutTwo, outTwoPipe, pipeTypeOutTwo] = process.env.outTwo.split(',');
const inRoutingKey = process.env.inRoutingKey || '#';
const outOneRoutingKey = process.env.outOneRoutingKey || '';
const outTwoRoutingKey = process.env.outTwoRoutingKey || '';

// content-based Router --> single
// recipient list --> multiple
const mode = process.env.mode;

const criteriaPath = '/etc/config/criteria';
const routingLogic = JSON.parse(fs.readFileSync(criteriaPath, 'utf8'));

amqp.connect(pipeAddressIn, function(error0, connection) {
  if (error0) {
    throw error0;
  }

  connection.createChannel(function(error1, channel) {
    if (error1) {
      throw error1;
    }

    setupInputPipe(channel, inPipe, pipeTypeIn, inRoutingKey, function(inputQueue) {
      setupOutputPipe(channel, outOnePipe, pipeTypeOutOne);
      setupOutputPipe(channel, outTwoPipe, pipeTypeOutTwo);

      console.log("Waiting for messages in %s.", inPipe);

      channel.consume(inputQueue, function(msg) {
        const message = JSON.parse(msg.content.toString());

        if (mode === 'single') {
          const destination = routeMessageSingle(message);
          if (destination) {
            sendToDestination(channel, destination, message);
          } else {
            console.log("No matching condition, message routed to default.");
            sendToDestination(channel, routingLogic.default, message);
          }
        } else if (mode === 'multiple') {
          const destinations = routeMessageMultiple(message);
          if (destinations.length > 0) {
            destinations.forEach(destination => {
              sendToDestination(channel, destination, message);
            });
          } else {
            console.log("No matching conditions, message routed to default.");
            sendToDestination(channel, routingLogic.default, message);
          }
        } else {
          console.error(`Unknown mode: ${mode}`);
        }

        channel.ack(msg);
      }, {
        noAck: false
      });
    });
  });
});

function setupInputPipe(channel, inPipe, pipeTypeIn, routingKey, callback) {
  if (pipeTypeIn === 'queue') {
    channel.assertQueue(inPipe);
    callback(inPipe);
  } else if (pipeTypeIn === 'topic') {
    channel.assertExchange(inPipe, 'topic');
    channel.assertQueue('', { exclusive: true }, function(error2, q) {
      if (error2) {
        throw error2;
      }
      channel.bindQueue(q.queue, inPipe, routingKey);
      callback(q.queue);
    });
  } else {
    console.error(`Unknown input pipe type: ${pipeTypeIn}`);
  }
}

function setupOutputPipe(channel, pipe, type) {
  if (type === 'queue') {
    channel.assertQueue(pipe);
  } else if (type === 'topic') {
    channel.assertExchange(pipe, 'topic');
  } else {
    console.error(`Unknown output pipe type: ${type}`);
  }
}

function sendToDestination(channel, destination, message) {
  if (destination === outOnePipe && pipeTypeOutOne === 'queue') {
    channel.sendToQueue(outOnePipe, Buffer.from(JSON.stringify(message)));
    console.log("Sent message to queue %s: %s", outOnePipe, JSON.stringify(message));
  } else if (destination === outOnePipe && pipeTypeOutOne === 'topic') {
    channel.publish(outOnePipe, outOneRoutingKey, Buffer.from(JSON.stringify(message)));
    console.log("Sent message to topic %s with routing key %s: %s", outOnePipe, outOneRoutingKey, JSON.stringify(message));
  } else if (destination === outTwoPipe && pipeTypeOutTwo === 'queue') {
    channel.sendToQueue(outTwoPipe, Buffer.from(JSON.stringify(message)));
    console.log("Sent message to queue %s: %s", outTwoPipe, JSON.stringify(message));
  } else if (destination === outTwoPipe && pipeTypeOutTwo === 'topic') {
    channel.publish(outTwoPipe, outTwoRoutingKey, Buffer.from(JSON.stringify(message)));
    console.log("Sent message to topic %s with routing key %s: %s", outTwoPipe, outTwoRoutingKey, JSON.stringify(message));
  } else {
    console.error("Unknown destination or pipe type.");
  }
}

function routeMessageSingle(message) {
  for (const rule of routingLogic.criterias) {
    if (eval(rule.condition)) {
      return rule.destination;
    }
  }
  return routingLogic.default;
}

function routeMessageMultiple(message) {
  const destinations = [];
  for (const rule of routingLogic.criterias) {
    if (eval(rule.condition)) {
      destinations.push(rule.destination);
    }
  }
  
  if (destinations.length === 0 && routingLogic.default) {
    destinations.push(routingLogic.default);
  }
  return destinations;
}
