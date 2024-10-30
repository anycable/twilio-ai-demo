# Pin npm packages by running ./bin/importmap

pin "application"
pin "@hotwired/turbo-rails", to: "turbo.min.js"
pin "@hotwired/stimulus", to: "stimulus.min.js"
pin "@hotwired/stimulus-loading", to: "stimulus-loading.js"
pin_all_from "app/javascript/controllers", under: "controllers"
pin "@anycable/turbo-stream", to: "@anycable--turbo-stream.js" # @0.7.0
pin "@anycable/core", to: "@anycable--core.js" # @0.9.1
pin "@hotwired/turbo", to: "@hotwired--turbo.js" # @8.0.12
pin "nanoevents" # @7.0.1
pin "@anycable/web", to: "@anycable--web.js" # @0.9.0
