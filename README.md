# substation
cloud-native stream processing infra

*Status*: Early days!

*Substation* is a distributed append-only log like Kafka. However, Substation differs in many ways:

 - Substation assumes something like Kubernetes is running underneath. As such, it doesn't have any cluster management logic itself.
 - Substation doesn't deal with records directly. Instead, the API exposes raw byte streams. Applications that want record semantics can layer on their own encodings. This means Substation is ideal for streaming large files, video, application logs, etc, which don't fit neatly into Kafka records.
 - Substation just uses HTTP -- not a custom protocol.
 - Substation apps do not require a fat client library. Instead, per-application Mailboxes keep track of client liveness, perform rebalances, and enforce at-least-once semantics on dumb clients. By analogy, Substation apps are like the `poll()` part of a `KafkaConsumer`, with the rest outsourced to the Mailbox.

Substation is designed to provide a `curl`-able API. Sending files is as easy as:

```
    $ echo hello > hello.txt
    $ curl -d@hello.txt http://substation-broker/topics/foo
    $ echo world > world.txt
    $ curl -d@world.txt http://substation-broker/topics/foo
```

Substation doesn't have a concept of "topic", so in the above example `/topics/foo` is just an arbitrary path. Files POSTed to a broker in this way get replicated to one or more replicas, which can be interrogated directly:

```
    $ curl http://substation-replica-1.substation-replicas/topics/foo > out.txt
    $ cat out.txt
    helloworld
```

Assuming there is only one replica, this will let you read back what you earlier POSTed. If you POST to the same path multiple times, the response will be concatenated (`"helloworld"` in the example above).

Usually there are multiple replicas for each broker, with log segments partitioned across them. For this reason, high-level applications will generally want to talk to a Mailbox, which combines multiple log segments from multiple replicas into a single ordered stream.

Mailboxes keep track of connected clients and their individual progress through the stream. When multiple clients connect to a single Mailbox, they behave like a Kafka consumer group: each log segment will be processed by at least one client. In order to save progress between GETs, Mailboxes leverage HTTP cookies:

```
   # read log segments from /topics/foo, and write word counts to /topics/foo/wc.
   for (( ; ; ))
   do
      curl http://my-mailbox/topics/foo --cookie cookiejar --cookie-jar cookiejar | wc > out.txt
      curl -d@out.txt http://substation-broker/topics/foo/wc
   done
```


