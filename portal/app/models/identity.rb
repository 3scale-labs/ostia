class Identity < ApplicationRecord
  belongs_to :user, required: true
end
