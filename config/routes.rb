Rails.application.routes.draw do
  resources :todos, except: [:index] do
    collection do
      get :index, defaults: {scope: "current"}
      get :archive, action: :index, defaults: {scope: "archive"}
      get :stale, action: :index, defaults: {scope: "stale"}
    end
  end

  resources :phone_calls, only: [:index]

  namespace :callbacks do
    post "/twilio" => "twilio_status#create"
  end

  # Define your application routes per the DSL in https://guides.rubyonrails.org/routing.html

  # Reveal health status on /up that returns 200 if the app boots with no exceptions, otherwise 500.
  # Can be used by load balancers and uptime monitors to verify that the app is live.
  get "up" => "rails/health#show", as: :rails_health_check

  # Defines the root path route ("/")
  root "todos#index", defaults: {scope: "current"}
end
