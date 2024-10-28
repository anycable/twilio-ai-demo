import { Controller } from "@hotwired/stimulus"

// This controller accompanies dialog elements and makes
// sure the `showModal()` method is called when the element appears on the page.
export default class extends Controller {
  static values = { open: Boolean }

  connect() {
    if (this.openValue) {
      this.element.showModal();
    }
  }
}
