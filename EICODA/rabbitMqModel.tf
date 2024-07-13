
provider "rabbitmq" {
  endpoint  = "http://localhost:15672"
  username  = "guest"
  password  = "guest"
}

resource "rabbitmq_queue" "myPipe" {
  name      = "myPipe"
  settings {
    durable    = true
    auto_delete = false
  }
}


resource "rabbitmq_queue" "myCoolMqttPipe" {
  name      = "myCoolMqttPipe"
  settings {
    durable    = true
    auto_delete = false
  }
}


resource "rabbitmq_queue" "productionPipe" {
  name      = "productionPipe"
  settings {
    durable    = true
    auto_delete = false
  }
}

