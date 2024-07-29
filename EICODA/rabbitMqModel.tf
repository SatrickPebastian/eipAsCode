
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

resource "rabbitmq_queue" "JOOOO" {
  name      = "JOOOO"
  vhost     = "/"
  settings {
    durable    = true
    auto_delete = false
  }
}


resource "rabbitmq_queue" "testOutputOne" {
  name      = "testOutputOne"
  vhost     = "/"
  settings {
    durable    = true
    auto_delete = false
  }
}


resource "rabbitmq_queue" "testOutputTwo" {
  name      = "testOutputTwo"
  vhost     = "/"
  settings {
    durable    = true
    auto_delete = false
  }
}


resource "rabbitmq_exchange" "testTopic" {
  name  = "testTopic"
  vhost = "/"
  settings {
    type        = "topic"
    durable     = true
    auto_delete = false
  }
}

