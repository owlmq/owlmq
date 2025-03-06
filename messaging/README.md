
# Objects inside the message system

2. Queue
Key: "q:nameHash"
Value:
```
name: ""
persistent: true | false
subscribers: []
```
3. Message
Key: "m:uuid"
Value:
```
queue_name: ""
content: ""
predecessor_message_hash: ""
isGenesis: true | false
```
4. User (later)
Key: "u:uuid"
Value:
```
name: ""
password_hash: ""
```

6. Exchange (later)
Key:"e:uuid"
Value:
```
name: ""
type: "direct | fanout | topic"
bindings: "[q:1234,q:2345,q:3456]"
```

