# IDEAS

* use plugins based on a grpc connection to the server
    * pro: multi language, only need a start command + they need to expose an api if for example they provide an external api (for example for amqp, or rest)
    * -> they need to "register" them self to the service so that it is aware which plugins are running.
    * -> maybe all nodes should share the same plugins so that loadbalancing over the nodes is easyer

