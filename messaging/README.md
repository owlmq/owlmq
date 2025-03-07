
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


## wie sieht ein durchlauf einer nachricht intern aus?

1. Producer sendet eine Nachricht an owlmq
2. die nachricht kommt an einem Knoten an
3. in der Nachricht steht ein queue_name
4. durch hashing des queue_name mit dem prefix "q:" lässt sich der TrackerKnoten im Ring finden hash("q:"+queue_name)
5. die nachricht kommt am eintrittsknoten an und wird mit PUT geschrieben, dabei fragt sie beim Tracker nach dem bisherigen HEAD nach und schreibt diesen als vorgänger-Nachricht, gleichzeitig aktualisiert der head den eigenen eintrag zum neuen HEAD der Queue

6. ->Version 1: der eintrittsknoten kennt ja nun die head nachricht und die neue nachricht und ist jetzt verantwortlich dafür dass alle consumer in der definierten weiß die nachrichten bekommen.
   -> Version 2: der eintrittsknoten gibt die nachricht an den tracker und dieser schreibt die nachricht. Anschließend sendet er die nachricht in der gewünschten weiße weiter

7. der consumer meldet sich bei einer queue an
8. der eintrittsknoten informiert die queue_worker (diese hat er im schritt 5 über die subscriber gefunden) dass es eine neue nachricht gibt
9. die queue_worker senden die Nachricht gemäß des gewünschten musters
