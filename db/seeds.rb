# Generate some example todos

Todo.create([
  {description: "Read the AnyCable off Rails blog post", deadline: 1.week.ago},
  {description: "Create a new Rails demo app", deadline: 5.days.ago, completed_at: 5.days.ago},
  {description: "Figure out how to convert PCM 24kHz to mulaw 8kHz in Ruby", deadline: 3.days.ago, completed_at: 3.days.ago},
  {description: "Write a blog post about AnyCable, Twilio and OpenAI", deadline: Date.current, completed: true},
  {description: "Deploy application to Fly.io", deadline: 2.days.since},
  # Some future tasks
  {description: "Upgrade to Ruby 3.#{RUBY_VERSION.split(".")[1].to_i.succ}", deadline: "#{Date.current.year}-12-25"}
])
