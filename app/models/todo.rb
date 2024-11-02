class Todo < ApplicationRecord
  include Turbo::Broadcastable

  validates :description, :deadline, presence: true

  # That must be implemented via a form object, but for now, we can use a virtual attribute.
  attribute :completed, :boolean, default: false
  before_validation { self.completed_at = completed ? Time.current : nil }
  after_initialize { self.completed = completed_at.present? }

  # Realtime features
  after_create_commit { broadcast_replace_to "todos", target: "new_item_notification", partial: "todos/item_added", locals: {todo: self} }
  after_update_commit { broadcast_replace_to "todos" }
  after_destroy_commit { broadcast_remove_to "todos" }

  scope :completed, -> { where.not(completed_at: nil) }
  scope :incomplete, -> { where(completed_at: nil) }

  scope :current, -> { incomplete.where(deadline: Date.current...).order(deadline: :asc) }
  scope :archive, -> { completed.where(deadline: (1.week.ago)...).order(deadline: :desc) }
  scope :stale, -> { incomplete.where(deadline: (..(Date.current - 1.day))).order(deadline: :desc) }
end
