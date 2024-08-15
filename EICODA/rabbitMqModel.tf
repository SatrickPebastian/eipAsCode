
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


resource "rabbitmq_queue" "InvoiceQueue" {
  name      = "InvoiceQueue"
  vhost     = "/"
  settings {
    durable    = true
    auto_delete = false
  }
}


resource "rabbitmq_queue" "ResultsQueue" {
  name      = "ResultsQueue"
  vhost     = "/"
  settings {
    durable    = true
    auto_delete = false
  }
}


resource "rabbitmq_exchange" "TranslatedTopic" {
  name  = "TranslatedTopic"
  vhost = "/"
  settings {
    type        = "topic"
    durable     = true
    auto_delete = false
  }
}

