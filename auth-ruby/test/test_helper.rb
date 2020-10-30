# frozen_string_literal: true

require 'pathname'

root = Pathname(__dir__).expand_path.dirname

$LOAD_PATH << root.join('grpc')
$LOAD_PATH << root.join('src')

Module.new do
  def file_fixture(filename)
    Pathname.new(File.join(__dir__, "fixtures/#{filename}"))
  end

  prepend_features(Minitest::Spec)
end

require 'auth'
