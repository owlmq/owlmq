# owlmq


## TODO

- [ ] GET und PUT sollen fingertable benutzen


## idears
### schnittstelle nach außen
* use grpc to connect/interact with the db
    * JWT Token authentication for access managment
        * maybe adding visibility roles (later)
    * use streaming of grpc for constant flows of messages

### einfügen von nachrichten
* wenn der chord ring steht kann werden die einzelnen einträge mittels hashing algorithmus verteilt, oder mittels uuid (zufällige verteilung).
* sendet nun ein teilnehmer eine nachricht wird diese mit einem zeitstempel versehen (entweder am eintrittsknoten im chord ring oder beim letztendlichen speicherort)
    * eintrittsknoten hätte das problem dass je nach eintrittsknoten die reihenfolge der nachrichten verloren gehen kann
    * speicherort hätte das problem dass je nach dem welcher weg im chord ring genommen wird die reihenfolge verloren geht

### speicherung
* die queues können als schlüsselwertpaare abgelegt werden, und jeweils der head, tail zeiger wird auf mehreren knoten im system gespeichert.
    * als schlüssel wird ein json objekt genutzt welches zusätzlich noch einen laufindex beinhaltet; das hat zur folge dass nach dem hashing des schlüssels die einzelnen nachrichten gleichmäßig verteilt im system abgelegt werden.
    * möchte ein beobachter, observer einer queue die letzte nachricht kann er sich auf einen knoten subscriben welcher den head, tail zeiger hat; ändert sich dieser dann bekommt er die letzte nachricht
* die queues können jeweils auf einem knoten im system gespeichert werden, dann könnte man eine gute Datenstruktur wählen und hätte bessere performance lokal
* eine weitere idee ist es nachrichten queues entweder auf der festplatte oder im arbeitsspeicher zu halten, hier könnte auch eine interessante hybrid lösung implementiert werden (für mehr performance werden aktuelle nachrichten im arbeitsspeicher gehalten, alte nachrichten zB älter als eine definierte zeit werden auf die platte geschrieben, wären dann langsamer erreichbar aber würde die geschwindigkeit erhöhen und I/O zugriffe senken)

### lookup des knotens mit dem chord protocol

* [IDEE] der lookup funktioniert so wie im chord protocol und die fingertable wird ergänzt, allerdings wird die fingertable regelmäßig bis auf vorgänger und nachfolger geleert. Alle anderen einträge der fingertable sind mit einem zeitstempel versehen und nach ablauf der zeit aus der liste entfernt. Auf diese weise bleiben nur die aktuell gebrauchten einträge in der liste eine abfrage ist also schneller da kleinere liste. Außerdem ist die wahrscheinlichkeit dass in kurzem abstand ein eintrag benötigt wird höher egal ob lesend oder schriebend.

### verteilugn von nachrichten und queues

#### idee 1
die queues auf die sich beobachter registrieren können jeweils über den namen auch gehashed werden. kommt nun bei einem beliebeigen knoten eine nachricht am system an wird einmal die nachricht gehashed und im chord ring abgelegt anderer seits wird der name der queue gehashed. nun wird einmal die nachricht an den knoten im ring geschickt welcher sie entgültig speichert anderer seits wird sie auch an den knoten geschickt auf welchem sich die beobachter registrieren um sie schneller zustellen zu können.
möchte ein neuer beobachter nach einer zeit alte nachrichten abfragen dann macht er eine anfrage an einen beliebigen knoten dieser stellt eine anfrage an alle knoten und sammelt die nachrichten im gesamt system ein und sortiert sie.

#### idee 2

zu jeder queue wird nur ein zeiger auf die letzte nachricht gespeichert, ergänzt nun ein producer eine nachricht speichert der eintrittsknoten lokal im queue-tail eintrag der queue den ort der speicherung/den hash der nachricht. die anfragen an das system würden dann über den hash direkt durch den eintrittsknoten erfolgen.

### replication/ausfallssicherheit

leader(hauptort eines schlüssel-wert paares) und follower: die replication funktioniert so dass der leader einer information immer der successor des nächsten im chord ring ist. das hat den vorteil dass der leader immer leicht testen kann ob es noch genug kopien gibt falls nicht nummer er sich den nächsten knoten des rings. fällt nun der leader aus ist in chord vorgesehen dass der schlüsselbereich automatisch an den nachfolger geht, er also der neue leader der informationen wird. jetzt fehlt eine kopie für einen teil der daten und diese kann er einfach einrichten.
Ein weiterer vorteil ist dass die plazierung der knoten im chord ring nicht von örtlicher nähe abhängt, so ist das system zusätzlich geschütz vor einen ausfall sobald es genügend knoten beinhaltet.

### ausfallssicherheit (verbindung)

um für ausfallssicherheit bei der verbindung zu sorgen könnte es sinn machen wenn während der verbindung regelmäßig überprüft wird ob wer der direkte nachfolger des knotens ist mit welchem man verbunden ist. fällt der knoten aus kann man versuchen sich mit dem nächsten zu verbinden

### Plugins


* use plugins based on a grpc connection to the server
    * pro: multi language, only need a start command + they need to expose an api if for example they provide an external api (for example for amqp, or rest)
    * -> they need to "register" them self to the service so that it is aware which plugins are running.
    * -> maybe all nodes should share the same plugins so that loadbalancing over the nodes is easyer

### configuration file

* sobald die grundstruktur erstmal steht soll jede node konfigurierbar sein mittels viper und einem konfigurationsfile, wichtig gerade für den cloud context ist aber dass die werte durch envs überschrieben werden können

### loadbalancing

* maybe use envoy (if it supports loadbalancing of grpc)

### Logs

* the system logs of owlmq can be stored inside owlmq itself


### Message Queue on Key-Value

The following structure is the way the objects of owl are stored on the key-value interface.

#### Key-Prefix

2. Queue
Key: "q:uuid"
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
```
4. User
Key: "u:uuid"
Value:
```
name: ""
password_hash: ""
```
5. Node
Key: "n:nodeid"
Value:
```
```

6. Exchange (later)
Key:"e:uuid"
Value:
```
name: ""
type: "direct | fanout | topic"
bindings: "[q:1234,q:2345,q:3456]"
```


### Plugin amqp

https://www.amqp.org/specification/1.0

## Usage API

```
TODO
```



## Links
https://pdos.csail.mit.edu/papers/ton:chord/paper-ton.pdf
http://localhost:8080/
https://github.com/owlmq/owlmq
https://en.wikipedia.org/wiki/Multiversion_concurrency_control
