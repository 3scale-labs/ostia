class AddCredentialsToIdentity < ActiveRecord::Migration[6.0]
  def change
    add_column :identities, :credentials, :json
  end
end
