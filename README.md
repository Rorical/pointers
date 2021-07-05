# pointers
 decentralized kv store based on pubsub in libp2p

# Protocol Specification 

 ### Over View
The pointer would receive updates from pubsub network and broadcast their local latest data so that the network would reach a final consistency.

New nodes who join the network will connect to nearby nodes and ask them for data.

### Operations Message
To put, modify or delete a value from the network, a node would publish an `Operate` message.

```protobuf
enum Operations {
  Mut = 0;
  Del = 1;
}

message Operate {
  Operations op = 1;
  optional string value = 2;

  string sign = 3;
  int64 time = 4;
}
```

Message `Operate` includes the type of operation in the form of enum, the optional new value, the signature of the publisher, and the time it sent.

Nodes would first compare the time of the message with the time in existed local storage. If the message is later than that in storage, nodes would start to check the validity of the signature and if it is valid, the local storage would be updated by the new message.

`Mut` in `Operations` would update the value, and `Del` would wipe out this value in the whole network. Every node that receives the `Del` operation would delete the corresponding value and end the subscription of the corresponding topic in pubsub.

### Ask & Answer Message
To get a value from nodes it connected, a node would send an `Ask` message to and receive `Answer` from them.

```protobuf
message Ask {
  string key = 1;
}

enum Status {
  Ok = 0;
  NotExist = 1;
}

message Answer {
  Status status = 1;
  optional Operate op = 2;
}
```

Message `Ask` would tell other nodes that the value of a specific key is needed. Any node that received this message could check their storage to see if they have this value.

When a node finds the value, it would return an `Answer` message with the `status` field that tells whether it found the corresponding value or not. If `status` is `Ok`, the receiver should be able to read the `op` field and store this value in its local storage.

### Storage
In the need of validating the message, the value in the data store should not be only the single value, so there are other data other than the value.

```protobuf
message StoreValue {
    string value = 2;

    string sign = 3;
    int64 time = 4;
}
```

The message `StoreValue` would be stored in the local key-value database in the form of binary.

### Mechanism Details
A node can store and listen to any data it is interested, so every key-value has its corresponding pubsub topic. The name of the topic would be generated from:

```go
Topic := "/pointer/record/" + Key
```

And then the node would listen to the topic, getting ready to update its local value whenever.

If it is new to this value, it is still possible to get quick update from nearby nodes. The node will discover some nodes with the same topic, and send `Ask` message to them with a protocol id from:

```go
ProtocolId := "/pointer/fetch/0.0.1"
```

Others would send their local value to the new one, and the new one is able to catch up with the network.

Every node in the same topic will send an `Operate` message to the network with a specific time interval, enabling others to catch up with new values.