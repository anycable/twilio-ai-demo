import { Application } from "@hotwired/stimulus"

const application = Application.start()

// Configure Stimulus development experience
application.debug = document.documentElement.classList.contains("debug");
window.Stimulus   = application

export { application }
