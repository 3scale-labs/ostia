Rails.application.routes.draw do
  resources :clients
  match '/auth/:provider', to: 'sessions#new', via: %i[get], as: :auth
  match '/auth/:provider/callback', to: 'sessions#create', via: %i[get post]
  match '/logout', to: 'sessions#destroy', via: %i[get post]
  match '/login', to: 'sessions#new', via: %i[get]
  # For details on the DSL available within this file, see https://guides.rubyonrails.org/routing.html
end
