require 'yaml/store'

class Config
  def initialize(file)
    @store = YAML::Store.new(file)
  end


end
