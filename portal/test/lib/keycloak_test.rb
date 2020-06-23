require "test_helper"

class KeycloakTest < ActiveSupport::TestCase

  test 'client' do
    client = Keycloak::Client.new(server: 'http://localhost:8080', realm: 'portal')

    puts client.discovery
  end

  test 'create_client' do
    client = Keycloak::Client.new(server: 'http://localhost:8080', realm: 'portal')

    pp client.create_client({ client_name: 'foobar' }).attributes.to_h
  end
end
