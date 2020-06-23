class CreateClients < ActiveRecord::Migration[6.0]
  def change
    create_table :clients do |t|
      t.string :registration_client_uri
      t.string :registration_access_token

      t.references :user, null: false, foreign_key: true

      t.timestamps
    end
  end
end
