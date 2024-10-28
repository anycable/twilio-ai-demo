class CreateTodos < ActiveRecord::Migration[8.0]
  def change
    create_table :todos do |t|
      t.string :description
      t.datetime :completed_at
      t.date :deadline

      t.timestamps
    end
  end
end
