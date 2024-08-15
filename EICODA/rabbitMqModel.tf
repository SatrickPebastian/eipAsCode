terraform {
  required_providers {
    rabbitmq = {
      source = "cyrilgdn/rabbitmq"
      version = "1.8.0"
    }
  }
}

provider "rabbitmq" {
  endpoint  = "http://localhost:15672"
  username  = "admin"
  password  = "password"
}

resource "rabbitmq_queue" "OrderQueue" {
  name      = "OrderQueue"
  vhost     = "/"
  settings {
    durable    = true
    auto_delete = false
  }
}


resource "rabbitmq_queue" "TaxQueue" {
  name      = "TaxQueue"
  vhost     = "/"
  settings {
    durable    = true
    auto_delete = false
  }
}

