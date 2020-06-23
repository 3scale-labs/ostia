class Client < ApplicationRecord
  belongs_to :user

  attribute :client_name, :string
  attribute :client_id, :string
  attribute :client_secret, :string

  def assign_keycloak_attributes(response)
    assign_attributes ActionController::Parameters.new(response.attributes).permit(Client.attribute_names)
  end
end
