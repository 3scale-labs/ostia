json.extract! client, :id, :client_id, :client_secret, :registration_uri, :registration_access_token, :user_id, :created_at, :updated_at
json.url client_url(client, format: :json)
