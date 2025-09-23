# learn-pub-sub-starter (Peril)

This is the starter code used in Boot.dev's [Learn Pub/Sub](https://learn.boot.dev/learn-pub-sub) course.

## Running the code

- Use `./rabbit.sh start/stop/logs` to manage the RabbitMQ container.
    - After running `./rabbit.sh start`, the RabbitMQ UI will be available at [http://localhost:15672](http://localhost:15672).
    - The default username and password are `guest` and `guest`.
- Run `go run cmd/client/main.go` to start the client.
- Run `go run cmd/server/main.go` to start the server.

## Notes about RabbitMQ

Pub/Sub (publish/subscribe) is a pattern software systems can use to communicate. Are particularly useful when many events are happening in real-time and disparate parts of the system need to react to those events.

Publishers publish, and subscribers subscribe. Alternative terms:

- Publisher = Producer = Sender
- Subscriber = Consumer = Receiver

Are often used to enable "event-driven design", or "event-driven architecture". An event-driven architecture uses events to trigger and communicate between decoupled systems.

**Message Broker**

A **message broker** is a middleman that allows different parts of the system to communicate without knowing about each other.

**RabbitMQ** is a popular open-source message broker that implements the **AMQP protocol**. It's open-source, flexible, powerful, and (reasonably) easy to use.

To run a RabbitMQ container, run the following command:

```bash
docker run -it --rm --name rabbitmq -p 5672:5672 -p 15672:15672 rabbitmq:3.13-management
```

The **Advanced Message Queuing Protocol (AMQP)** is an open standard for passing business messages between applications or organizations.

While Kafka, SQS, Google Pub/Sub, and NATS use their own protocols (not AMQP), AMQP is still an open standard, and other message brokers like ActiveMQ, Qpid, Solace and Azure Service Bus also implement it.

Two other popular protocols are:

- MQTT is designed for small IoT devices and as such is optimized to be lightweight and energy efficient.
- STOMP is designed for web applications and is designed to be simple and easy to use.

**Publishers and Queues**

In RabbitMQ, an [**exchange**](https://www.rabbitmq.com/tutorials/amqp-concepts#exchanges) is where publishers send messages, typically with a routing key.

The exchange takes the message, uses the routing key as a filter, and sends the message to any queues that are listening for that routing key.

Publishers don't know about queues at all. They just send messages to exchanges, sometimes with a routing key.

Terminology:

- Exchange: A routing agent that sends messages to queues.
- Binding: A link between an exchange and a queue that uses a routing key to decide which messages go to the queue.
- Queue: A buffer in the RabbitMQ server that holds messages until they are consumed.
- Channel: A virtual connection inside a connection that allows you to create queues, exchanges, and publish messages.
- Connection: A TCP connection to the RabbitMQ server.

RabbitMQ Exchange Types:

- Direct: Messages are routed to the queues based on the message routing key exactly matching the binding key of the queue.
- Topic: Messages are routed to queues based on wildcard matches between the routing key and the routing pattern specified in the binding.
- Fanout: It routes messages to all of the queues bound to it, ignoring the routing key.
- Headers: Routes based on header values instead of the routing key. It's similar to topic but uses message header attributes for routing.

direct and topic are the most commonly useful in backend Pub/Sub architectures.

**Queues** are where the messages are stored after being routed through the exchange. Messages sit in a queue until they are consumed by a subscriber.

Queues can be "durable" or "transient". Durable queues survive a RabbitMQ server restart, while transient queues do not.

The metadata of a durable queue is stored on disk, while transient queues are only stored in memory.

We can also set the auto-delete and exclusive properties of our queues:

- Exclusive: The queue can only be used by the connection that created it.
- Auto-delete: The queue will be automatically deleted when its last connection is closed.

**Subscribers & Routing**

**Consumers** are programs that connect to queues and pull the messages out of them.

A queue can have 0, 1, or many consumers.

- If a queue has no consumers, messages will accumulate in the queue and never be processed.
- If a queue has one consumer, that consumer will process all messages in the queue (assuming it can keep up).
- If a queue has many consumers, messages will be distributed between them in a round-robin fashion (unless you set a priority).

The exclusive flag can be used to tell the RabbitMQ server to only allow one consumer to connect to the queue at a time. Often pub/sub needs fall into one of two categories:

- Process an event once-per-server-instance (good for ephemeral, exclusive queues with one consumer)
- Process an event once, period (good for durable, non-exclusive queues with many consumers)

Routing keys in RabbitMQ are made up of words separated by dots. For example, the routing key `user.created` is made up of two words: user and created.

RabbitMQ supports two types of **wildcards** in routing keys:

- `*` (star) substitutes for exactly one word
- `#` (hash) substitutes for zero or more words

Technically, you can name your exchanges, queues, and routing keys whatever you want, but it's critically important to choose good names. Not only will it make your system easier to understand, but it can also make it more flexible and powerful.

**Exchange Naming**: It's common for one "system" to all use the same exchange. Similar to how you might have a single "database" within a Postgres instance, you might just have a single "exchange" within a RabbitMQ instance.

**Queue Naming**: When I'm working with a direct key -> queue relationship, I'll often name the queue the same as the key, but add a word to describe the intended consumer. For example, if I have a routing key user.created, I might create a queue for my "email notifier" service called user.created.email_notifier.

If I have a queue that consumes all user events, I might name it user.all.billing_service.

If I have temporary queues, I might append a UUID to the queue name to ensure uniqueness. For example, maybe I have web servers that scale up and down based on traffic, and each server needs a copy of "comment created" events. I might name each server's queue one of:

comment.created.bb7a488b-b4e9-4b16-a697-51c20a09b87b
comment.created.8d0a9d3e-5244-460b-bacc-80ae2b802677
comment.created.6814c13f-c33b-4ff7-a4f8-98c718fea980
...

I'll often use auto-generated queue names like this with transient, auto-delete, and exclusive properties so they can be created and destroyed as the system restarts and scales.

**Routing Key Naming**: This is the one that you really want to get right. Not only do you want the routing key names to be descriptive, but you also want them to be flexible for potential wildcard matching. I've found that often a noun.verb pattern works well. For example:

user.created
user.updated
comment.created
comment.deleted
etc.

This allows you to easily bind queues to all events of a certain type, or all events that affect a certain entity.

**Delivery (Dead Letter Exchanges and Queues)**

In an asynchronous system like RabbitMQ, the sender and receiver are decoupled. The sender doesn't need to know if the message was successfully delivered to the receiver. That has benefits, like simplicity and performance, but it also means that the chance of bugs increases.

To address this, it's common in PubSub systems to aggregate messages that fail to be processed into a dead letter queue. Queues can be configured to send messages that fail to be processed to a dead letter exchange, which then routes the message to a dead letter queue.

> Example of creating a dead letter exchange and queue:
> 1. Create a new exchange called peril_dlx of type fanout. Use the default settings.
> 2. Fanout is a good choice because we want all failed messages sent to the exchange to be routed to the queue, without needing to worry about routing keys.
> 3. Using the UI, create a new queue called peril_dlq.
> 4. Go to the queue's page and bind the queue to the peril_dlx exchange with no routing key. Leave the default settings

**Ack and Nack**: how a consumer tells the message broker that an individual message succeeded or failed to be processed. When a consumer receives a message, it must acknowledge it.

If the subscriber crashes or fails to process the message, the message broker can just re-queue the message to be processed again, or discard it (perhaps to a dead-letter queue).

There are really 3 options for acknowledging a message:

- Acknowledge: Processed successfully.
- Nack and requeue: Not processed successfully, but should be requeued on the same queue to be processed again (retry).
- Nack and discard: Not processed successfully, and should be discarded (to a dead-letter queue if configured or just deleted entirely).

Example of declaring a queue with a dead letter exchange:

```go
queue, err := ch.QueueDeclare(
    queueName,                       // name
    queueType == SimpleQueueDurable, // durable
    queueType != SimpleQueueDurable, // delete when unused
    queueType != SimpleQueueDurable, // exclusive
    false,                           // no-wait
    amqp.Table{
        "x-dead-letter-exchange": "peril_dlx",
    },
)
```

**Exact Delivery**: Delivering messages is hard. When you architect a system, you need to decide what guarantees to make. The three main types are:

- At-least-once delivery: If the message broker isn't sure the consumer received the message, it's retried.
- At-most-once delivery: If the message broker isn't sure the consumer received the message, it's discarded.
- Exactly-once delivery: The message is guaranteed to be delivered once and only once.

| Type          | Complexity | Efficiency | Reliability |
| ------------- | ---------- | ---------- | ----------- |
| At-least-once | Medium     | Medium     | Medium      |
| At-most-once  | Low        | High       | Low         |
| Exactly-once  | High       | Low        | High        |

At-least-once: In RabbitMQ, at-least-once delivery is the default. If a consumer fails to process a message, the message broker will just re-queue the message to be processed again. That means that you typically want to write your consumer code in such a way that it can process the same message multiple times without causing problems.

For example, if you have a message that says "Falkor created an account", and your consumer is responsible for sending a verification SMS, you can simply have your consumer check if it already sent an SMS to Falkor in the last 3 minutes before sending another one. That way, even if the message is processed multiple times, only one SMS is sent.

NackRequeue is the default behavior in Rabbit, and it's an example of at-least-once delivery.

At-most-once: At-most-once delivery makes more sense when you're dealing with messages that, frankly, aren't mission-critical. For example, instead of a message that represents a user account, maybe it's just a debug log. At-most-once delivery is more efficient from a performance perspective because it doesn't require the message broker to keep track of which messages have been processed, but obviously, it's less reliable.

Exactly-once: Exactly-once delivery is nearly impossible. That said, there are certainly ways to approximate it to the point of it being reliable from a practical perspective. However, of the three options, exactly-once delivery is the most difficult to implement and the most inefficient (slow).

At-least-once delivery is generally a good "default" choice for most systems.

**Nack Requeue**: As a general rule, you want to split your consumer's errors into two classes:

- Logical errors: Unlikely to be resolved with a retry. For example, a message is malformed JSON, or the ID of a user doesn't exist in the database.
- Transient errors: Likely to be resolved with a retry. For example, a network timeout, or a database connection error.

If you NackRequeue a message, it will be requeued to the primary queue to be processed again. This can be very bad if the error isn't transient as it will just be reprocessed over and over forever, blocking other messages and incurring large processing costs. Only NackRequeue messages if you're confident a retry will resolve the issue!

**Serialization**

JSON isn't the most efficient way to serialize data, so when you're sending massive amounts of data, you might want to consider a more efficient format.

**Gob**: In Go, the standard library has a package called [encoding/gob](https://pkg.go.dev/encoding/gob) that can be used to serialize and deserialize data. It's more efficient than JSON, but it's not human-readable. It's a binary format that's faster to encode and decode.

**Schema**: We've serialized structs to JSON and Gob, but there are many other possible choices like protocol buffers or Avro. While choosing which serialization format to use is important, it's also important to be careful about the shape or "schema" of the data you're serializing. As a general rule, if you make breaking changes to a schema, make sure you handle backward compatibility.

**Nodes and Clusters**

In production, you'd likely have an entire cluster of nodes. Some advantages of a large cluster include:

- High Availability: If one node goes down, other nodes can take over.
- Scalability: You're not constrained by the resources of a single machine.
- Redundancy: If one node goes down, the messages aren't lost.

**Resources**:

- CPU: Faster nodes (more cores, higher clock speed) and more nodes can both help.
- Memory: More RAM per node and more nodes can both help.
- Disk: More disk space per node and more nodes can both help.
- Network Bandwidth: In a cloud setting, bandwidth is usually provisioned in proportion to a node's size.

I've found that using a cluster of 3 nodes is a solid starting point for most production applications, even if you're processing thousands of messages per second. I've also found that when you find your nodes starting to hit limits on CPU, RAM, or Disk, it's generally better to scale vertically first (more powerful nodes) before you go crazy horizontally (larger number of nodes).

More nodes mean more resources, but it also means more management overhead and complexity.

How Do You Know? The overview tab in the RabbitMQ management console is the best place to start. It will show you high-level stats about the resource usage of your cluster.

**Backpressure** is a common problem in Pub/Sub systems. It happens when messages are being published to a queue faster than they can be consumed. This leads to a growing queue size, which can eventually cause the system to run out of memory or disk space.

**Healthy Queues Are Empty**: a full queue is an unhealthy queue because it grows faster than it can be consumed. This is dangerous because it can lead to the system running out of memory or disk space. If Rabbit goes down, your whole system goes down with it. Most of the time a healthy cluster, even if it's processing thousands of messages per second, will have mostly empty queues. You always want to be able to consume messages as fast as they can be published.

**Prefetch**: When you run a consumer, you may have assumed this process for message consumption:

- Fetch a message from the queue (across the network, which can be slow)
- Process the message
- Acknowledge the message
- Repeat

But that would slow everything down to a crawl due to the full network round trip for every message. Instead, RabbitMQ allows you to prefetch messages. When you prefetch messages, RabbitMQ will send you a batch of messages at once, the client library will store them in memory, and you can process them one by one. Much faster. The diagram shows 3 consumers each prefetching batches of 2.

Example in go:

```go
// Set prefetch to control message flow
err = ch.Qos(
    10,    // prefetch count - max messages to deliver without ack
    0,     // prefetch size - 0 means no size limit
    false, // global - apply to entire connection
)
```

**Quorum Queues**: Generally speaking, there are 2 queue types to worry about:

- Classic queues (we've been using these)
- Quorum queues

- Classic queues are the default and are great for most use cases. They are fast and simple. However, they have a single point of failure: the node that the queue is on. If that node goes down, the queue is lost.

You might be thinking, "Wait! You told me Rabbit is a distributed system!" And you're right, Rabbit is distributed, but classic queues are not. They are stored on a single node. If that node goes down, the queue is lost, at least until the node comes back online.

- Quorum queues are designed to be more resilient. They are stored on multiple nodes, so if one node goes down, the queue is still available. The tradeoff is that because quorum queues are stored on multiple nodes, they are slower than classic queues.

As a general rule, use classic queues for my ephemeral queues (transient, auto-delete, etc). I use quorum queues for most of my durable queues.