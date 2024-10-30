// Configure your import map in config/importmap.rb. Read more: https://github.com/rails/importmap-rails
import "@hotwired/turbo"

import { start } from "@anycable/turbo-stream"
import { createCable } from "@anycable/web"

const cable = createCable({
  logLevel: document.documentElement.classList.contains("debug") ? "debug" : "info",
  protocol: "actioncable-v1-ext-json"
})

start(cable, { delayedUnsubscribe: true })

import "controllers"
